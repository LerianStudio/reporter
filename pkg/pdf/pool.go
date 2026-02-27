// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package pdf

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"

	cn "github.com/LerianStudio/reporter/pkg/constant"

	"github.com/LerianStudio/lib-commons/v3/commons/log"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

//go:generate mockgen --destination=pool.mock.go --package=pdf --copyright_file=../../COPYRIGHT . PDFGenerator

// Compile-time interface satisfaction check.
var _ PDFGenerator = (*WorkerPool)(nil)

// PDFGenerator defines the interface for submitting PDF generation tasks.
type PDFGenerator interface {
	// Submit sends an HTML string to the pool for PDF generation and blocks until completion.
	Submit(html, filename string) error
}

// Task represents a task to generate a PDF.
type Task struct {
	HTML     string
	Filename string
	Result   chan error
}

// WorkerPool manager multiple Chrome workers to generate PDFs.
type WorkerPool struct {
	tasks   chan Task
	wg      *sync.WaitGroup
	workers int
	timeout time.Duration
	logger  log.Logger
}

// NewWorkerPool creates a new worker pool.
func NewWorkerPool(num int, timeout time.Duration, logger log.Logger) *WorkerPool {
	wp := &WorkerPool{
		tasks:   make(chan Task),
		wg:      &sync.WaitGroup{},
		workers: num,
		timeout: timeout,
		logger:  logger,
	}
	for i := 0; i < num; i++ {
		wp.wg.Add(1)

		go func(workerID int) {
			defer func() {
				if r := recover(); r != nil {
					wp.logger.Errorf("Panic recovered in PDF worker %d: %v\nStack: %s", workerID, r, string(debug.Stack()))
				}
			}()

			wp.startWorker(workerID)
		}(i)
	}

	return wp
}

// startWorker runs a Chrome worker to generate PDFs.
// Creates a single browser process per worker and reuses it for all tasks.
func (wp *WorkerPool) startWorker(_ int) {
	defer wp.wg.Done()

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), wp.getChromeOptions()...)
	defer allocCancel()

	for task := range wp.tasks {
		wp.processTask(allocCtx, task)
	}
}

// getChromeOptions returns optimized Chrome flags for PDF generation in containers with memory limits.
func (wp *WorkerPool) getChromeOptions() []chromedp.ExecAllocatorOption {
	return []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-plugins", true),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-features", "TranslateUI,site-per-process"),

		chromedp.Flag("max-old-space-size", cn.PDFChromeMaxOldSpaceSize),
		chromedp.Flag("js-flags", "--max-old-space-size="+cn.PDFChromeMaxOldSpaceSize),
		chromedp.Flag("disable-software-rasterizer", true),
		chromedp.Flag("disable-namespace-sandbox", true),

		chromedp.Flag("force-fieldtrials", "OmniboxBundledExperimentV1/Disabled"),
	}
}

// processTask handles a single PDF generation task.
func (wp *WorkerPool) processTask(allocCtx context.Context, task Task) {
	htmlSizeKB := float64(len(task.HTML)) / cn.PDFBytesPerKB
	wp.logger.Infof("Starting PDF generation for task: %s (HTML size: %.2f KB, timeout: %v)", task.Filename, htmlSizeKB, wp.timeout)

	if len(task.HTML) > cn.PDFLargeHTMLThreshold {
		wp.logger.Warnf("Large HTML detected (%.2f KB). Consider increasing PDF_TIMEOUT_SECONDS if timeouts occur", htmlSizeKB)
	}

	ctx, ctxCancel := chromedp.NewContext(allocCtx)
	defer ctxCancel()

	ctxTimeout, cancelTimeout := context.WithTimeout(ctx, wp.timeout)
	defer cancelTimeout()

	tmpFileName, err := wp.createTempHTMLFile(task.HTML)
	if err != nil {
		task.Result <- err
		return
	}

	pdfBuf, err := wp.generatePDFFromFile(ctxTimeout, tmpFileName)

	err = wp.processPDFResult(pdfBuf, task.Filename, err)

	err = wp.cleanupTempFile(tmpFileName, err)

	task.Result <- err
}

// createTempHTMLFile creates a temporary HTML file with the provided content.
func (wp *WorkerPool) createTempHTMLFile(html string) (string, error) {
	tmpFile, err := os.CreateTemp("", "pdf-*.html")
	if err != nil {
		wp.logger.Errorf("Failed to create temp HTML file: %v", err)
		return "", fmt.Errorf("failed to create temp HTML file: %w", err)
	}

	tmpFileName := tmpFile.Name()

	if err := tmpFile.Close(); err != nil {
		wp.logger.Warnf("Failed to close temp file %s: %v", tmpFileName, err)
	}

	if err := os.WriteFile(tmpFileName, []byte(html), cn.PDFFilePermissions); err != nil {
		wp.logger.Errorf("Failed to write HTML to temp file: %v", err)

		_ = os.Remove(tmpFileName)

		return "", fmt.Errorf("failed to write HTML to temp file: %w", err)
	}

	return tmpFileName, nil
}

// generatePDFFromFile generates a PDF from an HTML file using Chrome.
func (wp *WorkerPool) generatePDFFromFile(ctx context.Context, htmlFilePath string) ([]byte, error) {
	fileURL := "file://" + filepath.ToSlash(htmlFilePath)
	wp.logger.Infof("Navigating to file URL: %s", fileURL)

	var pdfBuf []byte

	err := chromedp.Run(ctx,
		chromedp.Navigate(fileURL),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return network.Enable().Do(ctx)
		}),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(cn.PDFRenderSettleDelay),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error

			pdfBuf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithPaperWidth(cn.PDFPaperWidthInches).
				WithPaperHeight(cn.PDFPaperHeightInches).
				WithMarginTop(cn.PDFMarginInches).
				WithMarginBottom(cn.PDFMarginInches).
				WithMarginLeft(cn.PDFMarginInches).
				WithMarginRight(cn.PDFMarginInches).
				WithDisplayHeaderFooter(false).
				Do(ctx)

			return err
		}),
	)
	if err != nil {
		wp.logPDFGenerationError(ctx, err)
		return nil, err
	}

	return pdfBuf, nil
}

// processPDFResult validates and writes the generated PDF to disk.
func (wp *WorkerPool) processPDFResult(pdfBuf []byte, filename string, err error) error {
	if err != nil {
		return err
	}

	if len(pdfBuf) < cn.PDFMinValidSizeBytes {
		wp.logger.Errorf("Final PDF too small: %d bytes", len(pdfBuf))
		return fmt.Errorf("generated PDF is too small (%d bytes), likely empty", len(pdfBuf))
	}

	if err := os.WriteFile(filename, pdfBuf, cn.PDFFilePermissions); err != nil {
		wp.logger.Errorf("Failed to write PDF file: %v", err)
		return err
	}

	wp.logger.Infof("PDF generated successfully: %d bytes written to %s", len(pdfBuf), filename)

	return nil
}

// cleanupTempFile removes the temporary HTML file and wraps cleanup errors with the original error.
func (wp *WorkerPool) cleanupTempFile(tmpFileName string, originalErr error) error {
	if err := os.Remove(tmpFileName); err != nil {
		wp.logger.Errorf("Failed to remove temp file %s: %v", tmpFileName, err)

		if originalErr == nil {
			return fmt.Errorf("generated PDF successfully but failed to remove temp file %s: %w", tmpFileName, err)
		}

		return fmt.Errorf("%w; additionally failed to remove temp file %s: %v", originalErr, tmpFileName, err)
	}

	return originalErr
}

// logPDFGenerationError logs PDF generation errors with appropriate context.
func (wp *WorkerPool) logPDFGenerationError(ctx context.Context, err error) {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		wp.logger.Errorf("PDF generation timeout (configured timeout: %v): %v", wp.timeout, err)
	} else if errors.Is(ctx.Err(), context.Canceled) {
		wp.logger.Errorf("PDF generation context canceled: %v", err)
	} else {
		wp.logger.Errorf("PDF generation failed: %v", err)
	}
}

// Submit sends a task to the pool and blocks until it is completed.
func (wp *WorkerPool) Submit(html, filename string) error {
	res := make(chan error, 1)
	wp.tasks <- Task{HTML: html, Filename: filename, Result: res}

	return <-res
}

// Close closes the pool and waits for all workers to finish.
func (wp *WorkerPool) Close() {
	close(wp.tasks)
	wp.wg.Wait()
}

// GetStats returns pool statistics
func (wp *WorkerPool) GetStats() map[string]any {
	return map[string]any{
		"workers":       wp.workers,
		"timeout":       wp.timeout,
		"tasks_pending": len(wp.tasks),
	}
}

// IsHealthy returns true if the pool is healthy
func (wp *WorkerPool) IsHealthy() bool {
	return wp.workers > 0 && wp.timeout > 0
}
