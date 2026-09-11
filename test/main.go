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
)

func MakeArgTypeLiteral(value string) argtree.ArgType {
	return argtree.ArgType{
		Name: "Literal" + value,
		Transform: func(s string) (any, error) {
			switch s {
			case value:
				return s, nil
			default:
				return nil, fmt.Errorf("is not " + value)
			}
		},
		List: func() []string {
			panic("TODO")
		},
		Example: func() string {
			panic("TODO")
		},
	}
}
func main() {
	var postKeySelect = argtree.ArgPossibility{
		Type: ArgTypeKeyModMethod,
		Children: []argtree.ArgPossibility{
			{
				Type:      ArgTypeKey,
				EndAction: argtree.EndActionLoop,
			},
		},
	}
	cliTree := []argtree.ArgPossibility{
		{
			Type: MakeArgTypeLiteral("modify"),
			Children: []argtree.ArgPossibility{
				{
					Type: ArgTypeKey,
					Children: []argtree.ArgPossibility{
						{
							Type: MakeArgTypeLiteral("from"),
							Children: []argtree.ArgPossibility{
								{
									Type: MakeArgTypeLiteral("TODO device names"),
									Children: []argtree.ArgPossibility{
										postKeySelect,
									},
								},
							},
						},
						postKeySelect,
					},
				},
			},
		},
	}

	input := []string{"modify", "k"}
	out, err := argtree.Parse(cliTree, input)
	if err != nil {
		panic(err)
	}

	fmt.Printf("State updated: %+v\n", out)
}
