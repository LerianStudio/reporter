package pongo

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/flosch/pongo2/v6"
	"github.com/shopspring/decimal"
)

// formatNumber formats a float64 as a string without scientific notation
func formatNumber(num float64) string {
	// Always use %f to avoid scientific notation completely
	return fmt.Sprintf("%.10f", num)
}

// stripZerosFilter formats a numeric value without trailing zeros and without rounding.
// Accepts int, int64, float64 or numeric strings.
func stripZerosFilter(in *pongo2.Value, _ *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	v := in.Interface()

	var dec decimal.Decimal

	switch t := v.(type) {
	case int:
		dec = decimal.NewFromInt(int64(t))
	case int64:
		dec = decimal.NewFromInt(t)
	case float64:
		dec = decimal.NewFromFloat(t)
	case string:
		d, err := decimal.NewFromString(t)
		if err != nil {
			return pongo2.AsSafeValue("NaN"), &pongo2.Error{Sender: "strip_zeros", OrigError: err}
		}

		dec = d
	default:
		// Fallback to string formatting
		s := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%v", v), "0"), ".")
		return pongo2.AsValue(s), nil
	}

	// decimal.String() already removes trailing zeros when possible
	out := dec.String()

	return pongo2.AsValue(out), nil
}

// percentOfFilter calculates the percentage of `in` relative to `param` and returns it as a formatted string.
// Returns "NaN" with an error if inputs are invalid or the denominator is zero.
func percentOfFilter(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	toDec := func(v any) (decimal.Decimal, error) {
		switch t := v.(type) {
		case int:
			return decimal.NewFromInt(int64(t)), nil
		case int64:
			return decimal.NewFromInt(t), nil
		case float64:
			return decimal.NewFromFloat(t), nil
		case string:
			return decimal.NewFromString(t)
		default:
			return decimal.Zero, fmt.Errorf("unsupported type %T", v)
		}
	}

	num, err1 := toDec(in.Interface())
	den, err2 := toDec(param.Interface())

	if err1 != nil || err2 != nil || den.IsZero() {
		return pongo2.AsSafeValue("NaN"), &pongo2.Error{
			Sender:    "percentOfFilter",
			OrigError: errors.New("invalid input or denominator is zero"),
		}
	}

	hundred := decimal.NewFromInt(100)
	pct := num.Mul(hundred).Div(den)

	return pongo2.AsValue(pct.StringFixed(2) + "%"), nil
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
	expression = strings.ReplaceAll(expression, " ", "")

	if expression == "" {
		return 0, fmt.Errorf("empty expression")
	}

	if strings.Contains(expression, "(") {
		return evaluateWithParentheses(expression)
	}

	if strings.Contains(expression, "**") {
		return evaluatePower(expression)
	}

	if strings.Contains(expression, "*") || strings.Contains(expression, "/") {
		return evaluateMultiplicationDivision(expression)
	}

	if strings.Contains(expression, "+") || strings.Contains(expression, "-") {
		return evaluateAdditionSubtraction(expression)
	}

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

	innerExpr := expression[start+1 : end]

	innerResult, err := evaluateArithmeticExpression(innerExpr)
	if err != nil {
		return 0, err
	}

	// Replace the parentheses expression with the result
	// Format the result as a decimal number without scientific notation
	newExpr := expression[:start] + formatNumber(innerResult) + expression[end+1:]

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
		return evaluateAdditionSubtraction(expression)
	}

	leftStart, rightEnd := findOperandBoundaries(expression, opIndex)
	left := expression[leftStart:opIndex]
	right := expression[opIndex+1 : rightEnd]

	leftVal, err := strconv.ParseFloat(left, 64)
	if err != nil {
		return 0, err
	}

	rightVal, err := strconv.ParseFloat(right, 64)
	if err != nil {
		return 0, err
	}

	result, err := performOperation(leftVal, rightVal, operator)
	if err != nil {
		return 0, err
	}

	// Create the new expression with the result
	// Format the result as a decimal number without scientific notation
	newExpr := expression[:leftStart] + formatNumber(result) + expression[rightEnd:]

	return evaluateArithmeticExpression(newExpr)
}

// evaluatePower handles ** operations (power/exponentiation)
func evaluatePower(expression string) (float64, error) {
	// Find the first ** operator
	powerIndex := strings.Index(expression, "**")
	if powerIndex == -1 {
		return evaluateMultiplicationDivision(expression)
	}

	leftStart := 0

	for i := powerIndex - 1; i >= 0; i-- {
		if expression[i] == '*' || expression[i] == '/' || expression[i] == '+' || expression[i] == '-' {
			leftStart = i + 1
			break
		}
	}

	left := expression[leftStart:powerIndex]

	// Find the right operand (everything after the operator until the next operator or end)
	rightStart := powerIndex + 2
	rightEnd := len(expression)

	for i := rightStart; i < len(expression); i++ {
		if expression[i] == '*' || expression[i] == '/' || expression[i] == '+' || expression[i] == '-' {
			rightEnd = i
			break
		}
	}

	right := expression[rightStart:rightEnd]

	leftVal, err := strconv.ParseFloat(left, 64)
	if err != nil {
		return 0, err
	}

	rightVal, err := strconv.ParseFloat(right, 64)
	if err != nil {
		return 0, err
	}

	result := math.Pow(leftVal, rightVal)

	// Create the new expression with the result
	// Format the result as a decimal number without scientific notation
	newExpr := expression[:leftStart] + formatNumber(result) + expression[rightEnd:]

	return evaluateArithmeticExpression(newExpr)
}

// evaluateAdditionSubtraction handles + and - operations
func evaluateAdditionSubtraction(expression string) (float64, error) {
	if strings.HasPrefix(expression, "-") {
		return strconv.ParseFloat(expression, 64)
	}

	var opIndex int

	var operator string

	if plusIndex := strings.Index(expression, "+"); plusIndex != -1 {
		opIndex = plusIndex
		operator = "+"
	} else if minusIndex := strings.Index(expression, "-"); minusIndex != -1 {
		opIndex = minusIndex
		operator = "-"
	} else {
		return strconv.ParseFloat(expression, 64)
	}

	left := expression[:opIndex]
	right := expression[opIndex+1:]

	leftVal, err := strconv.ParseFloat(left, 64)
	if err != nil {
		return 0, err
	}

	rightVal, err := strconv.ParseFloat(right, 64)
	if err != nil {
		return 0, err
	}

	var result float64

	switch operator {
	case "+":
		result = leftVal + rightVal
	case "-":
		result = leftVal - rightVal
	}

	return result, nil
}
