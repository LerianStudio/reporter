package pongo

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/flosch/pongo2/v6"
)

// formatNumber formats a float64 as a string without scientific notation
func formatNumber(num float64) string {
	// Always use %f to avoid scientific notation completely
	return fmt.Sprintf("%.10f", num)
}

// scaleFilter applies a scaling factor to a numeric or string value and formats the result to the specified precision.
// The `in` parameter is the input value, and `param` represents the scaling factor (number of decimal places).
// Returns a scaled and formatted string, or "NaN" with an error on invalid input or unsupported types.
func scaleFilter(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	v := in.Interface()
	scale := param.Integer()

	var intVal int64
	switch t := v.(type) {
	case int:
		intVal = int64(t)
	case int64:
		intVal = t
	case float64:
		intVal = int64(t)
	case string:
		parsed, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			return pongo2.AsSafeValue("NaN"), &pongo2.Error{
				Sender:    "scaleFilter",
				OrigError: fmt.Errorf("failed to parse string to int: %w", err),
			}
		}

		intVal = parsed
	default:
		return pongo2.AsSafeValue("NaN"), &pongo2.Error{
			Sender:    "scaleFilter",
			OrigError: fmt.Errorf("unsupported type %T", v),
		}
	}

	factor := math.Pow10(scale)
	scaled := float64(intVal) / factor

	return pongo2.AsValue(fmt.Sprintf("%.*f", scale, scaled)), nil
}

// percentOfFilter calculates the percentage of `in` relative to `param` and returns it as a formatted string.
// Returns "NaN" with an error if inputs are invalid or the denominator is zero.
func percentOfFilter(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	inVal := in.Interface()
	totalVal := param.Interface()

	getInt := func(v any) (int64, error) {
		switch t := v.(type) {
		case int:
			return int64(t), nil
		case int64:
			return t, nil
		case float64:
			return int64(t), nil
		case string:
			return strconv.ParseInt(t, 10, 64)
		default:
			return 0, fmt.Errorf("unsupported type %T", v)
		}
	}

	num, err1 := getInt(inVal)
	den, err2 := getInt(totalVal)

	if err1 != nil || err2 != nil || den == 0 {
		return pongo2.AsSafeValue("NaN"), &pongo2.Error{
			Sender:    "percentOfFilter",
			OrigError: errors.New("invalid input or denominator is zero"),
		}
	}

	percent := (float64(num) / float64(den)) * 100

	return pongo2.AsValue(fmt.Sprintf("%.2f%%", percent)), nil
}

// sliceFilter extracts a substring from the input string based on the specified "start:end" slice format in the parameter.
// Returns the sliced string or an error if the format is invalid or indices are out of bounds.
func sliceFilter(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	s := in.String()

	parts := strings.Split(param.String(), ":")
	if len(parts) != 2 {
		return nil, &pongo2.Error{
			Sender:    "slice",
			OrigError: fmt.Errorf("invalid slice format, expected 'start:end'"),
		}
	}

	start, err1 := strconv.Atoi(parts[0])
	end, err2 := strconv.Atoi(parts[1])

	if err1 != nil || err2 != nil {
		return nil, &pongo2.Error{
			Sender:    "slice",
			OrigError: fmt.Errorf("invalid start or end in slice"),
		}
	}

	if start < 0 {
		start = 0
	}

	if end > len(s) {
		end = len(s)
	}

	if start > end {
		start = end
	}

	return pongo2.AsValue(s[start:end]), nil
}

// evaluateArithmeticExpression evaluates a mathematical expression string.
// This is a simplified implementation for demonstration purposes.
func evaluateArithmeticExpression(expression string) (float64, error) {
	// Remove all spaces
	expression = strings.ReplaceAll(expression, " ", "")

	// Handle empty expression
	if expression == "" {
		return 0, fmt.Errorf("empty expression")
	}

	// Handle parentheses first
	if strings.Contains(expression, "(") {
		return evaluateWithParentheses(expression)
	}

	// Handle power operations first (highest precedence)
	if strings.Contains(expression, "**") {
		return evaluatePower(expression)
	}

	// Handle multiplication and division (medium precedence)
	if strings.Contains(expression, "*") || strings.Contains(expression, "/") {
		return evaluateMultiplicationDivision(expression)
	}

	// Handle addition and subtraction last (lowest precedence)
	if strings.Contains(expression, "+") || strings.Contains(expression, "-") {
		return evaluateAdditionSubtraction(expression)
	}

	// If no operators, try to parse as a single number
	return strconv.ParseFloat(expression, 64)
}

// evaluateWithParentheses handles expressions with parentheses
func evaluateWithParentheses(expression string) (float64, error) {
	// Find the innermost parentheses
	start := strings.LastIndex(expression, "(")
	if start == -1 {
		return evaluateArithmeticExpression(expression)
	}

	end := strings.Index(expression[start:], ")")
	if end == -1 {
		return 0, fmt.Errorf("unmatched parentheses in expression: %s", expression)
	}

	end += start

	// Extract the expression inside parentheses
	innerExpr := expression[start+1 : end]

	// Evaluate the inner expression
	innerResult, err := evaluateArithmeticExpression(innerExpr)
	if err != nil {
		return 0, err
	}

	// Replace the parentheses expression with the result
	// Format the result as a decimal number without scientific notation
	newExpr := expression[:start] + formatNumber(innerResult) + expression[end+1:]

	// Recursively evaluate the remaining expression
	return evaluateArithmeticExpression(newExpr)
}

// findNextOperator finds the next * or / operator in the expression
func findNextOperator(expression string) (int, string) {
	mulIndex := strings.Index(expression, "*")
	divIndex := strings.Index(expression, "/")

	if mulIndex == -1 && divIndex == -1 {
		return -1, ""
	}

	if mulIndex == -1 {
		return divIndex, "/"
	}

	if divIndex == -1 {
		return mulIndex, "*"
	}

	if mulIndex < divIndex {
		return mulIndex, "*"
	}

	return divIndex, "/"
}

// findOperandBoundaries finds the start and end positions of operands around an operator
func findOperandBoundaries(expression string, opIndex int) (int, int) {
	// Find left operand start (everything before the operator, but only until the previous operator)
	leftStart := 0

	for i := opIndex - 1; i >= 0; i-- {
		if expression[i] == '*' || expression[i] == '/' || expression[i] == '+' || expression[i] == '-' {
			leftStart = i + 1
			break
		}
	}

	// Find right operand end (everything after the operator until the next operator or end)
	rightStart := opIndex + 1
	rightEnd := len(expression)

	for i := rightStart; i < len(expression); i++ {
		if expression[i] == '*' || expression[i] == '/' || expression[i] == '+' || expression[i] == '-' {
			rightEnd = i
			break
		}
	}

	return leftStart, rightEnd
}

// performOperation performs the specified arithmetic operation
func performOperation(leftVal, rightVal float64, operator string) (float64, error) {
	switch operator {
	case "*":
		return leftVal * rightVal, nil
	case "/":
		if rightVal == 0 {
			return 0, fmt.Errorf("division by zero")
		}

		return leftVal / rightVal, nil
	default:
		return 0, fmt.Errorf("unknown operator: %s", operator)
	}
}

// evaluateMultiplicationDivision handles * and / operations
func evaluateMultiplicationDivision(expression string) (float64, error) {
	opIndex, operator := findNextOperator(expression)
	if opIndex == -1 {
		// No multiplication/division, evaluate as addition/subtraction
		return evaluateAdditionSubtraction(expression)
	}

	leftStart, rightEnd := findOperandBoundaries(expression, opIndex)
	left := expression[leftStart:opIndex]
	right := expression[opIndex+1 : rightEnd]

	// Evaluate left and right parts
	leftVal, err := strconv.ParseFloat(left, 64)
	if err != nil {
		return 0, err
	}

	rightVal, err := strconv.ParseFloat(right, 64)
	if err != nil {
		return 0, err
	}

	// Perform the operation
	result, err := performOperation(leftVal, rightVal, operator)
	if err != nil {
		return 0, err
	}

	// Create the new expression with the result
	// Format the result as a decimal number without scientific notation
	newExpr := expression[:leftStart] + formatNumber(result) + expression[rightEnd:]

	// Recursively evaluate the remaining expression
	return evaluateArithmeticExpression(newExpr)
}

// evaluatePower handles ** operations (power/exponentiation)
func evaluatePower(expression string) (float64, error) {
	// Find the first ** operator
	powerIndex := strings.Index(expression, "**")
	if powerIndex == -1 {
		// No power operator, evaluate as multiplication/division
		return evaluateMultiplicationDivision(expression)
	}

	// Find the left operand (everything before the operator, but only until the previous operator)
	leftStart := 0

	for i := powerIndex - 1; i >= 0; i-- {
		if expression[i] == '*' || expression[i] == '/' || expression[i] == '+' || expression[i] == '-' {
			leftStart = i + 1
			break
		}
	}

	left := expression[leftStart:powerIndex]

	// Find the right operand (everything after the operator until the next operator or end)
	rightStart := powerIndex + 2 // ** is 2 characters
	rightEnd := len(expression)

	for i := rightStart; i < len(expression); i++ {
		if expression[i] == '*' || expression[i] == '/' || expression[i] == '+' || expression[i] == '-' {
			rightEnd = i
			break
		}
	}

	right := expression[rightStart:rightEnd]

	// Evaluate left and right parts
	leftVal, err := strconv.ParseFloat(left, 64)
	if err != nil {
		return 0, err
	}

	rightVal, err := strconv.ParseFloat(right, 64)
	if err != nil {
		return 0, err
	}

	// Perform the power operation
	result := math.Pow(leftVal, rightVal)

	// Create the new expression with the result
	// Format the result as a decimal number without scientific notation
	newExpr := expression[:leftStart] + formatNumber(result) + expression[rightEnd:]

	// Recursively evaluate the remaining expression
	return evaluateArithmeticExpression(newExpr)
}

// evaluateAdditionSubtraction handles + and - operations
func evaluateAdditionSubtraction(expression string) (float64, error) {
	// Handle negative numbers at the beginning
	if strings.HasPrefix(expression, "-") {
		// This is a negative number, parse it directly
		return strconv.ParseFloat(expression, 64)
	}

	// Find the first + or - operator
	var opIndex int

	var operator string

	// Look for + first
	if plusIndex := strings.Index(expression, "+"); plusIndex != -1 {
		opIndex = plusIndex
		operator = "+"
	} else if minusIndex := strings.Index(expression, "-"); minusIndex != -1 {
		opIndex = minusIndex
		operator = "-"
	} else {
		// No operators, parse as a single number
		return strconv.ParseFloat(expression, 64)
	}

	// Split the expression
	left := expression[:opIndex]
	right := expression[opIndex+1:]

	// Evaluate left and right parts
	leftVal, err := strconv.ParseFloat(left, 64)
	if err != nil {
		return 0, err
	}

	rightVal, err := strconv.ParseFloat(right, 64)
	if err != nil {
		return 0, err
	}

	// Perform the operation
	var result float64

	switch operator {
	case "+":
		result = leftVal + rightVal
	case "-":
		result = leftVal - rightVal
	}

	return result, nil
}
