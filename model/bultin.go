package model

func NewBool(b bool) *bool       { return &b }
func NewInt(n int) *int          { return &n }
func NewInt8(n int8) *int8       { return &n }
func NewInt64(n int64) *int64    { return &n }
func NewString(s string) *string { return &s }

func Int64PositionInSlice(a int64, list []int64) int {
	for i, b := range list {
		if b == a {
			return i
		}
	}
	return -1
}
