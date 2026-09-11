package main

import (
	"fmt"

	"github.com/rsa17826/go-argtree"
	"github.com/rsa17826/go-input-lib"
)

var (
	ArgTypeKey = argtree.ArgType{
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
	ArgTypeInt = argtree.ArgType{
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
	ArgTypeKeyModMethod = argtree.ArgType{
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
	ArgTypeLiteral = argtree.ArgType{
		Name: "Literal",
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

func main() {
	state := make(map[string]any)

	// Example CLI tree structure from your setup
	cliTree := argtree.ArgTree{
		Possibilities: []argtree.ArgPossibility{
			{
				Type:   ArgTypeLiteral,
				Values: []any{"modify"},
				Children: argtree.ArgTree{
					Possibilities: []argtree.ArgPossibility{
						{
							Type: ArgTypeKey,
							Children: argtree.ArgTree{
								Possibilities: []argtree.ArgPossibility{
									{
										Type:     ArgTypeLiteral,
										Values:   []any{"from"},
										Children: argtree.ArgTree{},
									},
									{
										Type:     ArgTypeKeyModMethod,
										Children: argtree.ArgTree{},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	input := []string{"modify", "k"}
	err := argtree.Parse(cliTree, input, state)
	if err != nil {
		panic(err)
	}

	fmt.Printf("State updated: %+v\n", state)
}
