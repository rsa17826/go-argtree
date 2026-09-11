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

// EndActionEnd (the default, zero value) means: after this possibility
// matches (and its children, if any, finish matching), stop trying to match
// more from this list. EndActionLoop means: after this possibility matches,
// go back to the start of the *same* possibility list and try to match
// again against the remaining args, repeating until the args run out.
const (
	EndActionEnd = iota
	EndActionLoop
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

// parseSubtree repeatedly matches one possibility from `possibilities`
// against the front of `args`. It keeps going (from the start of
// `possibilities` again) as long as the most recently matched possibility
// has EndAction == EndActionLoop and there are args left; otherwise it stops
// after the first match. offset is the index into the original top-level
// args slice that args[0] corresponds to, used only for error positions.
func parseSubtree(possibilities []ArgPossibility, args []string, offset int, state OutData) (int, error) {
	totalConsumed := 0

	for {
		if len(args) == 0 {
			if totalConsumed == 0 && len(possibilities) > 0 {
				return 0, &ParseError{
					Pos: offset,
					Err: fmt.Errorf("unexpected end of arguments, expected one of: %s", expectedNames(possibilities)),
				}
			}
			return totalConsumed, nil
		}

		matchedPos, consumed, err := matchOnce(possibilities, args, offset, state)
		if err != nil {
			return 0, err
		}

		totalConsumed += consumed
		args = args[consumed:]
		offset += consumed

		if matchedPos.EndAction != EndActionLoop {
			return totalConsumed, nil
		}
		// EndActionLoop: go around again, matching from the start of the
		// same `possibilities` list against whatever args remain.
	}
}

// matchOnce tries each possibility against args[0] (and, if it has children,
// against the args that follow), backtracking to the next sibling whenever
// a match's subtree ultimately fails. It rolls back any tentative state
// writes made by an abandoned branch. On success it returns the matched
// possibility (so the caller can check its EndAction) and the total number
// of args consumed by it and its children.
func matchOnce(possibilities []ArgPossibility, args []string, offset int, state OutData) (*ArgPossibility, int, error) {
	var deepest *ParseError
	considerFailure := func(candidate *ParseError) {
		if deepest == nil || candidate.Pos > deepest.Pos {
			deepest = candidate
		}
	}

	for i := range possibilities {
		pos := &possibilities[i]
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

		hadPrev, prevVal := false, any(nil)
		if pos.Type.Name != "" {
			prevVal, hadPrev = state[pos.Type.Name]
			setMatchedValue(state, pos, transformedVal)
		}

		if len(pos.Children) == 0 {
			return pos, 1, nil
		}

		consumed, err := parseSubtree(pos.Children, args[1:], offset+1, state)
		if err == nil {
			return pos, 1 + consumed, nil
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
		return nil, 0, deepest
	}
	return nil, 0, &ParseError{
		Pos: offset,
		Arg: args[0],
		Err: fmt.Errorf("unexpected argument, expected one of: %s", expectedNames(possibilities)),
	}
}

// setMatchedValue records a match's value in state. For EndActionLoop
// possibilities, repeated matches accumulate into a []any under the same
// key instead of overwriting each other.
func setMatchedValue(state OutData, pos *ArgPossibility, val any) {
	if pos.EndAction == EndActionLoop {
		existing, _ := state[pos.Type.Name].([]any)
		state[pos.Type.Name] = append(existing, val)
		return
	}
	state[pos.Type.Name] = val
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
