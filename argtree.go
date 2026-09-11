package argtree

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
	Name     string
	Children ArgTree
}
type ArgType struct {
	Value     string
	@abstract
	Transform func(string) any
	@abstract
	List func() []string
	@abstract
	Example func() string
}

func init() {
	type Key ArgType {
		List
		Transform:func(str string){
			return str
		}
	}
	println(ArgTree{Possibilities: []ArgPossibility{
		{
			Name: "modify",
			Type: Null,
			Children: ArgTree{Possibilities: []ArgPossibility{{
				Type:     Key,
				Name:     "",
				Children: ArgTree{},
			}}},
		},
	}})
}
