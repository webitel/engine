package utils

import (
	"fmt"
	"testing"
)

func TestMapFnToString(t *testing.T) {
	actual := []int{1, 2, 3}
	fn := func(d int) string { return fmt.Sprintf("%d", d) }
	res := MapFn(fn, actual)
	expected := []string{"1", "2", "3"}
	equals(t, expected, res)
}

func TestMapFnInc(t *testing.T) {
	res := MapFn(func(d int) int { return d + 1 }, []int{1, 2, 3})
	equals(t, []int{2, 3, 4}, res)
}
