package main

import (
	"fmt"
	"maps"
	"slices"

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
			return slices.Collect(maps.Keys(input.StringToKey))
		},
	}
	ArgTypeKeyModMethod = MakeArgTypeAny([]string{"replace", "toggle", "maxPressTime", "minPressTime", "delay", "invert"})
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
			return []string{value}
		},
	}
}
func MakeArgTypeAny(values []string) argtree.ArgType {
	return argtree.ArgType{
		Transform: func(s string) (any, error) {
			if slices.Contains(values, s) {
				return s, nil
			}
			return nil, fmt.Errorf("is not one of %+v", values)
		},
		List: func() []string {
			return values
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
				If: func(d argtree.OutData) bool {
					switch d["modMethod"] {
					case "replace":
						return true
					default:
						return false
					}
				},
			},
			{
				Type:      argtree.ArgTypeInt,
				Name:      "turboTime",
				EndAction: argtree.EndActionLoop,
				If: func(d argtree.OutData) bool {
					switch d["modMethod"] {
					case "turbo":
						return true
					default:
						return false
					}
				},
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

	// input := []string{"modify", "k", "replace", "s", "modify", "k", "replace", "d"}
	// input := []string{"modify", "k", "replace", "s", "modify", "k", "replace"}
	// input := []string{"modify", "k", "replace", "s"}
	input := []string{"modify", "k"}
	// argtree.ShowHelp(cliTree)
	out, err := argtree.Parse(cliTree, input)
	if err != nil {
		panic(err)
	}

	fmt.Printf("State updated: %+v\n", out)
	for _, v := range out {
		fmt.Printf("1 %+v\n", v["sourceKey"])
		fmt.Printf("2 %+v\n", v["endKey"])
	}
}
