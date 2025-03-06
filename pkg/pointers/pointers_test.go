package pointers

import (
	"testing"
)

func TestInt(t *testing.T) {
	num := 42
	result := Int(num)
	if *result != num {
		t.Errorf("Int() = %v, want %v", *result, num)
	}
}
