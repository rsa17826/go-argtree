package main

import (
	"fmt"

	"github.com/rsa17826/go-argtree"
	"github.com/rsa17826/go-input-lib"
)

var (
	ArgTypeKey = argtree.ArgType{
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
		Transform: func(s string) (any, error) {
			switch s {
			case value:
				return s, nil
			default:
				return nil, fmt.Errorf("is not %s", value)
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
		Name: "modMethod",
		Children: []argtree.ArgPossibility{
			{
				Type:      ArgTypeKey,
				Name:      "endKey",
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
					Name: "sourceKey",
					Children: []argtree.ArgPossibility{
						{
							Type: MakeArgTypeLiteral("from"),
							Children: []argtree.ArgPossibility{
								{
									Type: MakeArgTypeLiteral("TODO device names"),
									Name: "deviceName",
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

	input := []string{"modify", "k", "replace", "s", "modify", "k", "replace", "d"}
	// input := []string{"modify", "k", "replace", "s", "modify", "k", "replace"}
	// input := []string{"modify", "k", "replace", "s"}
	// input := []string{"modify", "k"}
	out, err := argtree.Parse(cliTree, input)
	if err != nil {
		panic(err)
	}

	fmt.Printf("State updated: %+v\n", out)
}
