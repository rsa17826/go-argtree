package argtree

import "fmt"

// ArgType
// const (
//
//	Path = iota
//	String
//	Int
//	Float
//	SignedInt
//	SignedFloart
//	Time
//	SignedTime
//	Null
//
// )
const (
	EndActionLoop = iota
	EndActionEnd  = iota
)

type ArgTree struct {
	Possibilities []ArgPossibility
}
type ArgPossibility struct {
	Type      ArgType
	Values    []any
	Children  ArgTree
	EndAction int
}
type ArgType struct {
	Name      string
	Transform func(string) (any, error)
	List      func() []string
	Example   func() string
}
type OutData struct {
}

func Parse(tree ArgTree, args []string, state map[string]any) error {
	var finalOut = OutData{}
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
		fmt.Printf("finalOut: %v\n", finalOut)
	}
	return nil
}

var (
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
