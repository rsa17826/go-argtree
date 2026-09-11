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

// Parse walks the ArgTree using the provided input arguments and updates a state map.
func Parse(tree ArgTree, args []string, state map[string]any) error {
	currentTree := tree

	for i := 0; i < len(args); i++ {
		token := args[i]
		matched := false

		for _, p := range currentTree.Possibilities {
			// Match exact command name or a dynamic placeholder (empty Name)
			if p.Name == token || p.Name == "" {
				matched = true

				// If it's a dynamic node (Name == ""), capture the token as the value
				valToTransform := token
				if p.Name == "" {
					valToTransform = token
				}

				if p.Type.Transform != nil {
					val, err := p.Type.Transform(valToTransform)
					if err != nil {
						return fmt.Errorf("failed to transform %s: %w", token, err)
					}
					// Store based on node name or context
					if p.Name != "" {
						state[p.Name] = val
					} else {
						state["dynamic_key"] = val
					}
				}

				currentTree = p.Children
				break
			}
		}

		if !matched {
			return fmt.Errorf("unexpected argument or command: %s", token)
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
	ArgTypeKeyModMethod = ArgType{
		Name: "KeyModMethod",
		Transform: func(s string) (any, error) {
			switch s {
			case "replace":
				return s, nil
			default:
				return nil, fmt.Errorf("not a valid method")
			}
		},
		List: func() []string {
			panic("TODO")
		},
		Example: func() string {
			panic("TODO")
		},
	}
)
