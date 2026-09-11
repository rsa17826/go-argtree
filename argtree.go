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

type ArgPossibility struct {
	Type      ArgType
	Values    []any
	Children  []ArgPossibility
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

func Parse(tree []ArgPossibility, args []string) (OutData, error) {
	return parseSubtree(tree, args, state)
}
func parseSubtree(tree []ArgPossibility, args []string) (OutData, error) {
	var finalOut = OutData{}
	for argIdx := range args {
		var lastOut any
		var lastPossibility ArgPossibility
		for _, pos := range tree {
			var parser ArgType = pos.Type
			out, err := parser.Transform(args[argIdx])
			if err != nil {
				lastOut = out
				lastPossibility = pos
				out, err = parseSubtree(pos.Children, args[argIdx:])
				if err != nil {
					fmt.Printf("finalOut: %v\n", finalOut)
					return out
				}
			}
		}
	}
	return finalOut, nil
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
