package argtree

import (
	"fmt"

	"github.com/rsa17826/go-input-lib"
)

// ArgType
// const (
// 	Path = iota
// 	String
// 	Int
// 	Float
// 	SignedInt
// 	SignedFloart
// 	Time
// 	SignedTime
// 	Null
// )

type ArgTree struct {
	Possibilities []ArgPossibility
}
type ArgPossibility struct {
	Type     ArgType
	Values   []any
	Children ArgTree
}
type ArgType struct {
	Name      string
	Transform func(string) (any, error)
	List      func() []string
	Example   func() string
}

func Parse(tree ArgTree, args []string, state map[string]any) error {
	for argIdx := range args {
		var lastOut any
		var lastPossibility ArgPossibility
		for i := range tree.Possibilities {
			var parser ArgType = tree.Possibilities[i].Type
			out, err := parser.Transform(args[argIdx])
			if err != nil {
				lastOut = out
				lastPossibility = tree.Possibilities[i]
				break
			}
		}
	}
	return nil
}

var (
	ArgTypeKey = ArgType{
		Name: "Key",
		Transform: func(s string) (any, error) {
			key, ok := input.StringToKey[s]
			if !ok {
				return nil, fmt.Errorf("not a valid key name")
			}
			return key, nil
		},
		List: func() []string {
			panic("TODO")
		},
		Example: func() string {
			panic("TODO")
		},
	}
	ArgTypeInt = ArgType{
		Name: "Int",
		Transform: func(s string) (any, error) {
			return s, nil
		},
		List: func() []string {
			panic("TODO")
		},
		Example: func() string {
			panic("TODO")
		},
	}
)
