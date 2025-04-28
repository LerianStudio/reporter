package pongo

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/flosch/pongo2/v6"
)

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
