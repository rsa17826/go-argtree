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

type OutData map[string]any

func Parse(tree []ArgPossibility, args []string) (OutData, error) {
	state := make(OutData)
	_, err := parseSubtree(tree, args, state)
	if err != nil {
		return nil, err
	}
	return state, nil
}

func parseSubtree(possibilities []ArgPossibility, args []string, state OutData) (int, error) {
	if len(args) == 0 {
		if len(possibilities) == 0 {
			return 0, nil
		}
		return 0, fmt.Errorf("unexpected end of arguments")
	}

	for _, pos := range possibilities {
		matched := false
		var transformedVal any
		var err error

		// Check if it matches literal values first, otherwise use Type.Transform
		if len(pos.Values) > 0 {
			for _, v := range pos.Values {
				if strVal, ok := v.(string); ok && strVal == args[0] {
					matched = true
					transformedVal = strVal
					break
				}
			}
		} else if pos.Type.Transform != nil {
			transformedVal, err = pos.Type.Transform(args[0])
			if err == nil {
				matched = true
			}
		}

		if matched {
			if pos.Type.Name != "" {
				state[pos.Type.Name] = transformedVal
			}

			// If there are children, parse the remaining arguments recursively
			if len(pos.Children) > 0 {
				consumed, err := parseSubtree(pos.Children, args[1:], state)
				if err == nil {
					return 1 + consumed, nil
				}
				continue
			}

			return 1, nil
		}
	}

	return 0, fmt.Errorf("unexpected argument: %s", args[0])
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
