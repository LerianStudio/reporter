package pongo

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/flosch/pongo2/v6"
)

// aggregateTagNode represents a data structure to hold information for custom aggregation operations in templates.
// It includes the operation type, collection expression, field expression, filter expression, and scaling expression.
type aggregateTagNode struct {
	op             string            // Aggregation type ("count", "sum", "avg", "min", "max")
	collectionExpr pongo2.IEvaluator // Expression representing the data collection
	fieldExpr      pongo2.IEvaluator // Field expression for selecting specific fields (e.g., "by field")
	filterExpr     pongo2.IEvaluator // Condition to filter items in the dataset (e.g., "if condition" in the template tag)
	scaleExpr      pongo2.IEvaluator // Expression for scaling results (e.g., formatting decimals like "scale 2")
}

// dateNowNode represents a data structure to hold information for the dateNow template tag.
// It includes the expression representing the date format (e.g., "YYYY-MM-dd").
type dateNowNode struct {
	formatExpr pongo2.IEvaluator // Expression representing the date format (e.g., "YYYY-MM-dd")
}

// calcTagNode represents a calc tag that evaluates arithmetic expressions
type calcTagNode struct {
	expression string
}

// makeAggregateTag returns a pongo2.TagParser for creating custom aggregate template tags based on the specified operation.
func makeAggregateTag(op string) pongo2.TagParser {
	return func(doc *pongo2.Parser, start *pongo2.Token, args *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
		collectionExpr, err := args.ParseExpression()
		if err != nil {
			return nil, err
		}

		var fieldExpr pongo2.IEvaluator

		if op != "count" { // "count"` operation doesn't need a specific field to operate on (simply counts the number of elements)
			if t := args.Match(pongo2.TokenIdentifier, "by"); t == nil {
				return nil, args.Error("Expected 'by' keyword", nil)
			}

			fieldExpr, err = args.ParseExpression()
			if err != nil {
				return nil, err
			}
		}

		var filterExpr pongo2.IEvaluator
		if t := args.Match(pongo2.TokenIdentifier, "if"); t != nil {
			filterExpr, err = args.ParseExpression()
			if err != nil {
				return nil, err
			}
		}

		var scaleExpr pongo2.IEvaluator
		if t := args.Match(pongo2.TokenIdentifier, "scale"); t != nil {
			scaleExpr, err = args.ParseExpression()
			if err != nil {
				return nil, err
			}
		}

		return &aggregateTagNode{
			op:             op,
			collectionExpr: collectionExpr,
			fieldExpr:      fieldExpr,
			filterExpr:     filterExpr,
			scaleExpr:      scaleExpr,
		}, nil
	}
}

// makeDateNowTag creates a custom Pongo2 template tag that outputs the current date formatted according to the provided expression.
func makeDateNowTag() pongo2.TagParser {
	return func(doc *pongo2.Parser, start *pongo2.Token, args *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
		formatExpr, err := args.ParseExpression()
		if err != nil {
			return nil, err
		}

		return &dateNowNode{
			formatExpr: formatExpr,
		}, nil
	}
}

// Execute renders the current date in a specified format and writes it using the provided template writer.
func (node *dateNowNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	formatVal, err := node.formatExpr.Evaluate(ctx)
	if err != nil {
		return err
	}

	format := formatVal.String()
	goLayout := convertToGoDateLayout(format)
	output := time.Now().Format(goLayout)

	_, err2 := writer.WriteString(output)
	if err2 != nil {
		return ctx.Error("Failed to write date", nil)
	}

	return nil
}

// Execute processes a template tag by evaluating a collection and performing aggregation, then writes the result to the output.
func (node *aggregateTagNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	list, err := evaluateCollection(ctx, node.collectionExpr)
	if err != nil {
		return err
	}

	result, err := aggregateResult(ctx, list, node)
	if err != nil {
		return err
	}

	_, err2 := writer.WriteString(result)
	if err2 != nil {
		return ctx.Error("Error writing output", nil)
	}

	return nil
}

// evaluateCollection evaluates the given expression in the provided context and extracts a []map[string]any collection.
// Returns the evaluated collection or an error if the type assertion fails or evaluation encounters an issue.
func evaluateCollection(ctx *pongo2.ExecutionContext, expr pongo2.IEvaluator) ([]map[string]any, *pongo2.Error) {
	val, err := expr.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	list, ok := val.Interface().([]map[string]any)
	if !ok {
		return nil, ctx.Error("Expected []map[string]any for collection", nil)
	}

	return list, nil
}

// aggregateResult performs aggregation operations (sum, avg, min, max, count) on a filtered list within a template context.
func aggregateResult(ctx *pongo2.ExecutionContext, list []map[string]any, node *aggregateTagNode) (string, *pongo2.Error) {
	var total int64

	var count int

	var minVal *int64

	var maxVal *int64

	for _, item := range list {
		if !passesFilter(ctx, item, node.filterExpr) {
			continue
		}

		if node.op == "count" {
			count++
			continue
		}

		vInt, skip, err := extractIntValue(ctx, item, node.fieldExpr)
		if err != nil {
			return "", err
		}

		if skip {
			continue
		}

		switch node.op {
		case "sum", "avg":
			total += vInt
			count++
		case "min":
			if minVal == nil || vInt < *minVal {
				minVal = &vInt
			}
		case "max":
			if maxVal == nil || vInt > *maxVal {
				maxVal = &vInt
			}
		}
	}

	scale := getScale(ctx, node.scaleExpr)

	return formatOutput(node.op, total, count, minVal, maxVal, scale), nil
}

// passesFilter evaluates a filter expression on an item within a given execution context and returns true if the condition is met.
func passesFilter(ctx *pongo2.ExecutionContext, item map[string]any, filterExpr pongo2.IEvaluator) bool {
	if filterExpr == nil {
		return true
	}

	localCtx := pongo2.NewChildExecutionContext(ctx)
	for k, v := range item {
		localCtx.Private[k] = v
	}

	cond, err := filterExpr.Evaluate(localCtx)

	return err == nil && cond.IsTrue()
}

// extractIntValue retrieves an integer value from a nested map field specified by a field expression in the given context.
// It evaluates the field expression, extracts the field value, and converts it into int64 if possible.
// Returns the integer value, a boolean indicating if the field should be skipped, and an optional error.
func extractIntValue(ctx *pongo2.ExecutionContext, item map[string]any, fieldExpr pongo2.IEvaluator) (int64, bool, *pongo2.Error) {
	fieldNameVal, err := fieldExpr.Evaluate(ctx)
	if err != nil {
		return 0, false, err
	}

	fieldName := fieldNameVal.String()

	value, ok := getNestedField(item, fieldName)
	if !ok {
		return 0, true, nil
	}

	switch v := value.(type) {
	case int:
		return int64(v), false, nil
	case int64:
		return v, false, nil
	case float64:
		return int64(v), false, nil
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return parsed, false, nil
		}
	}

	return 0, true, nil
}

// getScale determines the scale (number of decimal places) for numerical formatting based on the evaluated expression.
// Returns the scale as an integer, or 0 if the expression is nil or evaluation fails.
func getScale(ctx *pongo2.ExecutionContext, expr pongo2.IEvaluator) int {
	if expr == nil {
		return 0
	}

	if scaleVal, err := expr.Evaluate(ctx); err == nil {
		return scaleVal.Integer()
	}

	return 0
}

// formatOutput formats the result of an aggregation operation (count, sum, avg, min, max) as a scaled string output.
func formatOutput(op string, total int64, count int, minVal, maxVal *int64, scale int) string {
	factor := math.Pow10(scale)

	switch op {
	case "count":
		return fmt.Sprintf("%d", count)
	case "sum":
		return fmt.Sprintf("%.*f", scale, float64(total)/factor)
	case "avg":
		if count > 0 {
			return fmt.Sprintf("%.*f", scale, (float64(total)/float64(count))/factor)
		}

		return fmt.Sprintf("%.*f", scale, 0.0)
	case "min":
		if minVal != nil {
			return fmt.Sprintf("%.*f", scale, float64(*minVal)/factor)
		}

		return fmt.Sprintf("%.*f", scale, 0.0)
	case "max":
		if maxVal != nil {
			return fmt.Sprintf("%.*f", scale, float64(*maxVal)/factor)
		}

		return fmt.Sprintf("%.*f", scale, 0.0)
	default:
		return "NaN"
	}
}

// convertToGoDateLayout converts a date format string from a custom format (e.g., "YYYY-MM-dd") to Go's date layout format.
func convertToGoDateLayout(layout string) string {
	replacer := strings.NewReplacer(
		"YYYY", "2006",
		"MM", "01",
		"dd", "02",
		"HH", "15",
		"mm", "04",
		"ss", "05",
	)

	return replacer.Replace(layout)
}

func makeCalcTag(_ *pongo2.Parser, _ *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	calcNode := &calcTagNode{}

	// Get the raw expression as string
	// We need to collect all remaining tokens to build the full expression
	var tokens []string

	for arguments.Remaining() > 0 {
		token := arguments.Current()
		tokens = append(tokens, token.Val)

		arguments.Consume()
	}

	expression := strings.Join(tokens, " ")

	// Remove spaces around dots to fix variable paths
	expression = strings.ReplaceAll(expression, " . ", ".")
	calcNode.expression = expression

	return calcNode, nil
}

func (node *calcTagNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	// Get the template context to access variables
	context := ctx.Public

	// Replace variables in the expression with their values
	expression := node.replaceVariables(node.expression, context, ctx.Private)

	// Evaluate the arithmetic expression
	result, err := evaluateArithmeticExpression(expression)
	if err != nil {
		return &pongo2.Error{
			Sender:    "calc",
			OrigError: err,
		}
	}

	// Write the result
	// Round to 10 decimal places to avoid artifacts (e.g., ...0000000001)
	rounded := math.Round(result*1e10) / 1e10
	// Format without scientific notation
	out := formatNumber(rounded)
	// Trim trailing zeros and trailing dot
	out = strings.TrimRight(out, "0")
	out = strings.TrimRight(out, ".")

	if _, err := writer.WriteString(out); err != nil {
		return ctx.Error("Error writing output", nil)
	}

	return nil
}

// shouldSkipMatch determines if a match should be skipped (operators, parentheses, numbers)
func shouldSkipMatch(match string) bool {
	// Skip operators and parentheses
	operators := []string{"+", "-", "*", "/", "**", "(", ")", "[", "]", "{", "}"}
	for _, op := range operators {
		if match == op {
			return true
		}
	}

	// Skip if it's a number (but only if it doesn't contain dots)
	if !strings.Contains(match, ".") {
		if _, err := strconv.ParseFloat(match, 64); err == nil {
			return true
		}
	}

	return false
}

// resolveVariableFromContext attempts to resolve a variable from a given context
func resolveVariableFromContext(match string, context pongo2.Context) (string, bool) {
	templateStr := fmt.Sprintf("{{ %s }}", match)

	template, err := pongo2.FromString(templateStr)
	if err != nil {
		return "", false
	}

	result, err := template.ExecuteBytes(context)
	if err != nil {
		return "", false
	}

	value := strings.TrimSpace(string(result))
	if value == "" {
		return "", false
	}

	// Try to parse as number to validate it's a valid numeric value
	if _, err := strconv.ParseFloat(value, 64); err != nil {
		return "", false
	}

	return value, true
}

func (node *calcTagNode) replaceVariables(expression string, context pongo2.Context, privateContext pongo2.Context) string {
	// Find all variable patterns in the expression (words with dots)
	// This regex matches patterns like: balance.available, midaz_transaction.balance.0.initial_balance, etc.
	re := regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*(\.[a-zA-Z0-9_]+)*\b`)

	expression = re.ReplaceAllStringFunc(expression, func(match string) string {
		// Skip operators, parentheses, and numbers
		if shouldSkipMatch(match) {
			return match
		}

		// Try to resolve the variable using Pongo2's template engine
		// First try the private context (loop variables)
		if value, ok := resolveVariableFromContext(match, privateContext); ok {
			return value
		}

		// If private context failed, try public context
		if value, ok := resolveVariableFromContext(match, context); ok {
			return value
		}

		// If both contexts failed, return "0"
		return "0"
	})

	return expression
}
