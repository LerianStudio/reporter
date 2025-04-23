package pongo

import (
	"fmt"
	"math"
	"strconv"

	"github.com/flosch/pongo2/v6"
)

// aggregateTagNode represents a data structure to hold information for custom aggregation operations in templates.
// It includes the operation type, collection expression, field expression, filter expression, and scaling expression.
type aggregateTagNode struct {
	op             string
	collectionExpr pongo2.IEvaluator
	fieldExpr      pongo2.IEvaluator
	filterExpr     pongo2.IEvaluator
	scaleExpr      pongo2.IEvaluator
}

// makeAggregateTag returns a pongo2.TagParser for creating custom aggregate template tags based on the specified operation.
func makeAggregateTag(op string) pongo2.TagParser {
	return func(doc *pongo2.Parser, start *pongo2.Token, args *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
		collectionExpr, err := args.ParseExpression()
		if err != nil {
			return nil, err
		}

		var fieldExpr pongo2.IEvaluator

		if op != "count" {
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
