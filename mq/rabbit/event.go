package rabbit

import (
	"fmt"
	"strconv"
)

//TODO interface
type Variables map[string]interface{}

type REvent Variables

func (e REvent) Name() string {
	v, _ := e["Event-Name"]
	return fmt.Sprintf("%v", v)
}

func (e REvent) Id() string {
	v, _ := e["Unique-ID"]
	return fmt.Sprintf("%v", v)
}

func (e REvent) GetVariable(name string) (string, bool) {
	v, k := e[name]
	return fmt.Sprintf("%v", v), k
}

func (e REvent) GetStringVariable(name string) string {
	v, k := e[name]
	if k {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func (e REvent) GetIntVariable(name string) (int, bool) {
	v, k := e[name]
	if !k {
		return 0, false
	}

	i, err := strconv.Atoi(fmt.Sprintf("%v", v))
	if err != nil {
		return 0, false
	}

	return i, true
}

func (e REvent) ToMapStringInterface() map[string]interface{} {
	return e
}
