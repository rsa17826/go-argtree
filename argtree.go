package argtree

import (
	"fmt"
	"strings"
)

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

// ParseError reports a failure to match the argument tree. Pos is the index
// (into the original args slice) of the argument that caused the deepest
// failure encountered while backtracking, so it points at the most likely
// actual mistake rather than wherever the recursion happened to unwind to.
type ParseError struct {
	Pos int
	Arg string
	Err error
}

func (e *ParseError) Error() string {
	if e.Arg == "" {
		return fmt.Sprintf("argument %d: %s", e.Pos, e.Err)
	}
	return fmt.Sprintf("argument %d (%q): %s", e.Pos, e.Arg, e.Err)
}

func (e *ParseError) Unwrap() error { return e.Err }

func Parse(tree []ArgPossibility, args []string) (OutData, error) {
	state := make(OutData)
	_, err := parseSubtree(tree, args, 0, state)
	if err != nil {
		return nil, err
	}
	return state, nil
}

// parseSubtree tries each possibility in turn (backtracking on failure) and
// returns the number of args consumed on success. offset is the index into
// the original top-level args slice that args[0] corresponds to, used only
// to produce accurate error positions.
func parseSubtree(possibilities []ArgPossibility, args []string, offset int, state OutData) (int, error) {
	if len(args) == 0 {
		if len(possibilities) == 0 {
			return 0, nil
		}
		return 0, &ParseError{
			Pos: offset,
			Err: fmt.Errorf("unexpected end of arguments, expected one of: %s", expectedNames(possibilities)),
		}
	}

	var deepest *ParseError

	considerFailure := func(candidate *ParseError) {
		if deepest == nil || candidate.Pos > deepest.Pos {
			deepest = candidate
		}
	}

	for _, pos := range possibilities {
		if pos.Type.Transform == nil {
			continue
		}

		transformedVal, err := pos.Type.Transform(args[0])
		if err != nil {
			considerFailure(&ParseError{
				Pos: offset,
				Arg: args[0],
				Err: fmt.Errorf("not a valid %s: %w", nameOrType(pos.Type), err),
			})
			continue
		}

		// Tentatively record this match so children can see it, but keep
		// enough info to roll back if this whole branch ends up failing.
		hadPrev, prevVal := false, any(nil)
		if pos.Type.Name != "" {
			prevVal, hadPrev = state[pos.Type.Name]
			state[pos.Type.Name] = transformedVal
		}

		if len(pos.Children) == 0 {
			return 1, nil
		}

		consumed, err := parseSubtree(pos.Children, args[1:], offset+1, state)
		if err == nil {
			return 1 + consumed, nil
		}

		// Branch failed: undo the tentative state write before trying the
		// next sibling possibility.
		if pos.Type.Name != "" {
			if hadPrev {
				state[pos.Type.Name] = prevVal
			} else {
				delete(state, pos.Type.Name)
			}
		}

		if pe, ok := err.(*ParseError); ok {
			considerFailure(pe)
		} else {
			considerFailure(&ParseError{Pos: offset, Err: err})
		}
	}

	if deepest != nil {
		return 0, deepest
	}
	return 0, &ParseError{
		Pos: offset,
		Arg: args[0],
		Err: fmt.Errorf("unexpected argument, expected one of: %s", expectedNames(possibilities)),
	}
}

func nameOrType(t ArgType) string {
	if t.Name != "" {
		return t.Name
	}
	return "value"
}

func expectedNames(possibilities []ArgPossibility) string {
	names := make([]string, 0, len(possibilities))
	for _, p := range possibilities {
		names = append(names, nameOrType(p.Type))
	}
	return strings.Join(names, ", ")
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
