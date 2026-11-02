package model

import "strconv"

var (
	CurrentVersion string = "dev"
	BuildNumber    string
)

func BuildNumberInt() int {
	n, _ := strconv.Atoi(BuildNumber)
	return n
}
