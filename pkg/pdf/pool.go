package pdf

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/LerianStudio/lib-commons/v2/commons/log"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

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

		go wp.startWorker(i)
	}

	return wp
}

// startWorker runs a Chrome worker to generate PDFs.
func (wp *WorkerPool) startWorker(_ int) {
	defer wp.wg.Done()

	// Optimized Chrome configuration for container environment
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-plugins", true),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-features", "TranslateUI"),
		chromedp.Flag("memory-pressure-off", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	for task := range wp.tasks {
		ctxTimeout, cancelTimeout := context.WithTimeout(ctx, wp.timeout)

		var pdfBuf []byte

		wp.logger.Infof("Starting PDF generation for task: %s (HTML size: %d bytes)", task.Filename, len(task.HTML))

		// Use temporary file approach (more reliable than data URL in containers)
		tmpFile, tmpErr := os.CreateTemp("", "pdf-*.html")
		if tmpErr != nil {
			err := fmt.Errorf("failed to create temp HTML file: %v", tmpErr)
			wp.logger.Errorf("PDF generation failed: %v", err)
			cancelTimeout()

			task.Result <- err

			continue
		}

		tmpFileName := tmpFile.Name()
		if closeErr := tmpFile.Close(); closeErr != nil {
			wp.logger.Warnf("Failed to close temp file %s: %v", tmpFileName, closeErr)
		}

		// Write HTML to temp file
		if writeErr := os.WriteFile(tmpFileName, []byte(task.HTML), 0600); writeErr != nil {
			err := fmt.Errorf("failed to write HTML to temp file: %v", writeErr)
			wp.logger.Errorf("PDF generation failed: %v", err)

			if removeErr := os.Remove(tmpFileName); removeErr != nil {
				wp.logger.Warnf("Failed to remove temp file %s: %v", tmpFileName, removeErr)
			}

			cancelTimeout()

			task.Result <- err

			continue
		}

		// Generate PDF from file
		fileURL := "file://" + filepath.ToSlash(tmpFileName)
		wp.logger.Infof("Navigating to file URL: %s", fileURL)

		err := chromedp.Run(ctxTimeout,
			chromedp.Navigate(fileURL),
			chromedp.WaitVisible("body", chromedp.ByQuery),
			chromedp.ActionFunc(func(ctx context.Context) error {
				return network.Enable().Do(ctx)
			}),
			chromedp.WaitReady("body", chromedp.ByQuery),
			chromedp.Sleep(500*time.Millisecond),
			chromedp.ActionFunc(func(ctx context.Context) error {
				var err error

				pdfBuf, _, err = page.PrintToPDF().
					WithPrintBackground(true).
					WithPaperWidth(8.5).
					WithPaperHeight(11).
					WithMarginTop(0.5).
					WithMarginBottom(0.5).
					WithMarginLeft(0.5).
					WithMarginRight(0.5).
					WithDisplayHeaderFooter(false).
					Do(ctx)

				return err
			}),
		)

		if removeErr := os.Remove(tmpFileName); removeErr != nil {
			wp.logger.Warnf("Failed to remove temp file %s: %v", tmpFileName, removeErr)
		}

		if err == nil {
			if len(pdfBuf) < 1000 {
				err = fmt.Errorf("generated PDF is too small (%d bytes), likely empty", len(pdfBuf))
				wp.logger.Errorf("Final PDF too small: %d bytes", len(pdfBuf))
			} else {
				err = os.WriteFile(task.Filename, pdfBuf, 0600)
				if err != nil {
					wp.logger.Errorf("Failed to write PDF file: %v", err)
				} else {
					wp.logger.Infof("PDF generated successfully: %d bytes written to %s", len(pdfBuf), task.Filename)
				}
			}
		} else {
			if errors.Is(ctxTimeout.Err(), context.DeadlineExceeded) {
				wp.logger.Errorf("PDF generation timeout (configured timeout: %v): %v", wp.timeout, err)
			} else if errors.Is(ctxTimeout.Err(), context.Canceled) {
				wp.logger.Errorf("PDF generation context canceled: %v", err)
			} else {
				wp.logger.Errorf("PDF generation failed: %v", err)
			}
		}

		cancelTimeout()

		task.Result <- err
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
