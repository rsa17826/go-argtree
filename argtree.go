package argtree

import (
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"time"
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
//
// )

// EndActionEnd (the default, zero value) means: once the match chain that
// ends at this leaf possibility completes, stop — Parse returns whatever
// args are left over as an error if any remain.
//
// EndActionLoop means: once the match chain that ends at this leaf
// possibility completes, go back to the ROOT of the whole tree (the slice
// originally passed to Parse) and try to match again against whatever args
// remain, repeating until the args run out. This is what lets a whole
// command shape like "modify k replace s" repeat: "modify k replace s
// modify k replace d".
const (
	EndActionEnd = iota
	EndActionLoop
)

type ArgPossibility struct {
	Type      ArgType
	Name      string
	Children  []ArgPossibility
	EndAction int
	// If, when non-nil, is checked against the state accumulated so far
	// (including whatever this possibility's siblings' ancestors already
	// wrote) before this possibility is even attempted. If it returns
	// false, this possibility is skipped entirely, as if it weren't in the
	// list - it won't be tried and won't appear in "expected one of" error
	// messages.
	If func(OutData) bool
	// IfDescription is shown by ShowHelp to explain when this possibility
	// applies. It exists because an arbitrary If closure can't be
	// introspected - if If is set, you should set this too, or help output
	// for this node just won't say what the condition is.
	IfDescription string
	Repeat        bool // Allows parsing consecutive matching tokens
	AllowZero     bool // If true, 0 matches is allowed. If false, requires at least 1 match.
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
	Pos  int
	Arg  string
	Err  error
	Tree []ArgPossibility
	Path []int
}

func (e *ParseError) Error() string {
	baseMsg := ""
	if e.Arg == "" {
		baseMsg = fmt.Sprintf("argument %d: %s", e.Pos, e.Err)
	} else {
		baseMsg = fmt.Sprintf("argument %d (%q): %s", e.Pos, e.Arg, e.Err)
	}

	if len(e.Tree) > 0 {
		return BuildErrorTree(e.Tree, e.Path) + baseMsg
	}
	return baseMsg
}

func (e *ParseError) Unwrap() error { return e.Err }
func parsePossibility(pos ArgPossibility, args []string) (consumed int, value any, err error) {
	if !pos.Repeat {
		// Normal single-token parsing
		val, err := pos.Type.Transform(args[0])
		if err != nil {
			return 0, nil, err
		}
		return 1, val, nil
	}

	// Repeat mode: Collect all consecutive valid matches into a slice
	var items []any
	idx := 0

	for idx < len(args) {
		val, err := pos.Type.Transform(args[idx])
		if err != nil {
			// Stop repeating on first non-matching token
			break
		}
		items = append(items, val)
		idx++
	}

	// Check minimum match threshold
	if len(items) == 0 && !pos.AllowZero {
		return 0, nil, fmt.Errorf("expected at least 1 argument for %s, got 0", pos.Name)
	}

	return idx, items, nil
}

// Parse matches `tree` against `args` from the start. If the leaf that
// terminates a successful match chain has EndActionLoop, matching restarts
// from the root of `tree` against whatever args remain, repeating until the
// args are exhausted. Each restart gets its own independent OutData, so a
// repeated command shape like "modify k replace s modify k replace d"
// produces one map per repetition rather than one merged map.
func Parse(tree []ArgPossibility, args []string) ([]OutData, error) {
	var results []OutData
	remaining := args
	offset := 0

	for {
		state := make(OutData)
		consumed, endAction, err := parseSubtree(tree, remaining, offset, state)
		if err != nil {
			if pe, ok := err.(*ParseError); ok {
				pe.Tree = tree // Capture the root tree for this pass
				return nil, pe
			}
			return nil, err
		}
		results = append(results, state)

		remaining = remaining[consumed:]
		offset += consumed

		if endAction != EndActionLoop || len(remaining) == 0 {
			break
		}
	}

	if len(remaining) > 0 {
		return nil, &ParseError{
			Pos: offset,
			Arg: remaining[0],
			Err: fmt.Errorf("unexpected trailing argument"),
		}
	}

	return results, nil
}

// parseSubtree matches one possibility from `possibilities` against args[0]
// (and, if it has children, against the args that follow), backtracking to
// the next sibling whenever a match's subtree ultimately fails, and rolling
// back any tentative state writes made by an abandoned branch.
//
// It returns the number of args consumed and the EndAction of the leaf that
// terminated the chain (bubbled up unchanged through parent matches, since
// only the terminal leaf's EndAction determines whether Parse should loop).
// offset is the index into the original top-level args slice that args[0]
// corresponds to, used only for error positions.
func parseSubtree(possibilities []ArgPossibility, args []string, offset int, state OutData) (int, int, error) {
	if len(args) == 0 {
		if hasViable(possibilities, state) {
			return 0, EndActionEnd, &ParseError{
				Pos: offset,
				Err: fmt.Errorf("unexpected end of arguments, expected one of: %s", expectedNames(possibilities, state)),
				// Leave Path empty so backtracking stops at the last matched parent node ("combo")
			}
		}
		return 0, EndActionEnd, nil
	}

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
		if pos.If != nil && !pos.If(state) {
			continue
		}

		// Delegate token matching and repetition handling to parsePossibility
		consumedTokens, transformedVal, err := parsePossibility(*pos, args)
		if err != nil {
			argStr := ""
			if len(args) > 0 {
				argStr = args[0]
			}
			considerFailure(&ParseError{
				Pos: offset,
				Arg: argStr,
				Err: fmt.Errorf("not a valid %s for %s - %v: %w", pos.Type.Name, pos.Name, pos.Type, err),
			})
			continue
		}

		hadPrev, prevVal := false, any(nil)
		if pos.Name != "" {
			prevVal, hadPrev = state[pos.Name]
			state[pos.Name] = transformedVal
		}

		if len(pos.Children) == 0 {
			return consumedTokens, pos.EndAction, nil
		}

		consumedChildren, endAction, err := parseSubtree(pos.Children, args[consumedTokens:], offset+consumedTokens, state)
		if err == nil {
			if consumedChildren == 0 {
				endAction = pos.EndAction
			}
			return consumedTokens + consumedChildren, endAction, nil
		}

		// Branch failed: undo the tentative state write before trying the next sibling
		if pos.Name != "" {
			if hadPrev {
				state[pos.Name] = prevVal
			} else {
				delete(state, pos.Name)
			}
		}

		if pe, ok := err.(*ParseError); ok {
			peCopy := *pe
			peCopy.Path = append([]int{i}, peCopy.Path...)
			considerFailure(&peCopy)
		} else {
			considerFailure(&ParseError{Pos: offset, Err: err, Path: []int{i}})
		}
	}

	if deepest != nil {
		return 0, EndActionEnd, deepest
	}
	return 0, EndActionEnd, &ParseError{
		Pos: offset,
		Arg: args[0],
		Err: fmt.Errorf("unexpected argument, expected one of: %s", expectedNames(possibilities, state)),
	}
}
func BuildErrorTree(tree []ArgPossibility, path []int) string {
	var b strings.Builder
	writeErrorLevel(&b, tree, "", path, 0, true)
	return b.String()
}

func writeErrorLevel(b *strings.Builder, possibilities []ArgPossibility, prefix string, path []int, depth int, active bool) {
	pathIndex := -1
	if depth < len(path) {
		pathIndex = path[depth]
	}

	for i, p := range possibilities {
		isLast := i == len(possibilities)-1

		// A node remains active (not greyed out) if its parent was active AND
		// either we have passed the path end (pathIndex == -1) or this node is on the path (i == pathIndex).
		nodeActive := active && (pathIndex == -1 || i == pathIndex)

		connector := "├─ "
		childPrefix := prefix + "│  "
		if isLast {
			connector = "└─ "
			childPrefix = prefix + "   "
		}

		colorModifier := ""
		if !nodeActive {
			colorModifier = "\033[90m"
		}

		b.WriteString(colorConnector)
		b.WriteString(prefix)
		b.WriteString(connector)
		b.WriteString(ansiReset)

		desc := describePossibility(p)
		if colorModifier == "" {
			b.WriteString(desc)
		} else {
			// Strip individual ANSI colors so the entire line uses dim gray
			desc = strings.ReplaceAll(
				strings.ReplaceAll(
					strings.ReplaceAll(
						strings.ReplaceAll(
							strings.ReplaceAll(
								strings.ReplaceAll(
									strings.ReplaceAll(
										strings.ReplaceAll(
											desc, ansiReset, "",
										), colorConnector, "",
									), colorLabel, "",
								), colorValues, "",
							), colorSep, "",
						), colorTypeName, "",
					), colorRepeat, "",
				), colorCondition, "",
			)
			b.WriteString(colorModifier)
			b.WriteString(desc)
			b.WriteString(ansiReset)
		}

		// Annotate nodes along the active failure path
		if nodeActive && depth == len(path)-1 && i == pathIndex {
			b.WriteString(" \033[31m< failed here\033[0m")
		}

		b.WriteString("\n")

		if len(p.Children) > 0 {
			writeErrorLevel(b, p.Children, childPrefix, path, depth+1, nodeActive)
		}
	}
}
func hasViable(possibilities []ArgPossibility, state OutData) bool {
	for _, p := range possibilities {
		if p.If == nil || p.If(state) {
			return true
		}
	}
	return false
}

func expectedNames(possibilities []ArgPossibility, state OutData) string {
	names := make([]string, 0, len(possibilities))
	for _, p := range possibilities {
		if p.If != nil && !p.If(state) {
			continue
		}
		names = append(names, p.Name)
	}
	return strings.Join(names, ", ")
}

// Complete returns the sorted, deduplicated List() values of every
// possibility that could legally come next, given the args a user has
// already typed in full (not including whatever word they're still in the
// middle of typing - that's a shell-side prefix-filtering concern, not
// this function's). If the already-typed args don't match the tree at all,
// or they exactly complete a valid command with no more root-loop
// repetitions possible, Complete returns nil (nothing to suggest).
func Complete(tree []ArgPossibility, args []string) []string {
	return completeSubtree(tree, tree, args, make(OutData))
}

// completeSubtree walks one pass over possibilities/args, but unlike
// parseSubtree it does not stop at the first possibility that matches -
// it tries every sibling whose Type.Transform accepts args[0], because
// several siblings can legally match the very same token while leading to
// different completions. This tree relies on exactly that shape: a bare
// "<key>" (EndActionLoop, no children, so it loops back to the root) sits
// right next to "<key> from <device>" (children include "from"). Both
// match a trailing key token identically; only by trying both do we learn
// that either "from" or a fresh top-level command can legally follow.
// Results from every matching sibling are merged, deduplicated and sorted.
//
// root is threaded through (rather than tracked via a return flag, as an
// earlier version of this function did) so that when a branch's terminal
// leaf is reached exactly where args run out and that leaf's EndAction is
// EndActionLoop, this function can fold the root's own suggestions
// straight into the merge - a fresh repetition could start next, mirroring
// how Parse starts each repetition fresh from the root.
//
// It doesn't need parseSubtree's backtracking-with-rollback machinery:
// each matching sibling gets its own cloned state, so failed or merely
// alternative branches never pollute each other.
func completeSubtree(root, possibilities []ArgPossibility, args []string, state OutData) []string {
	if len(args) == 0 || (len(args) == 1 && args[0] == "") {
		return viableSuggestions(possibilities, state)
	}

	var suggestions []string

	for i := range possibilities {
		pos := &possibilities[i]
		if pos.Type.Transform == nil {
			continue
		}
		if pos.If != nil && !pos.If(state) {
			continue
		}

		// Delegate token matching and repetition handling to parsePossibility
		consumedTokens, transformedVal, err := parsePossibility(*pos, args)
		if err != nil {
			continue
		}

		branchState := cloneState(state)
		if pos.Name != "" {
			branchState[pos.Name] = transformedVal
		}

		remainingArgs := args[consumedTokens:]
		// A lone trailing "" is the shell's "user just typed a space, next
		// word not started yet" sentinel, not a real leftover argument -
		// treat it the same as "no args left" for deciding whether this
		// branch's pass has fully completed.
		argsExhausted := len(remainingArgs) == 0 || (len(remainingArgs) == 1 && remainingArgs[0] == "")

		if len(pos.Children) == 0 {
			if pos.EndAction == EndActionLoop {
				if argsExhausted {
					suggestions = append(suggestions, viableSuggestions(root, make(OutData))...)
				} else {
					// Real leftover tokens past this loop-back leaf: a
					// fresh repetition starts here, exactly like Parse's
					// own top-level loop restarts from tree against
					// whatever args remain. Recurse from the root with a
					// brand new state.
					suggestions = append(suggestions, completeSubtree(root, root, remainingArgs, make(OutData))...)
				}
			}
			// Otherwise this branch is a dead end for completion purposes:
			// it doesn't loop, so nothing valid can follow (Parse would
			// reject leftover args here as a trailing-argument error).
			continue
		}

		suggestions = append(suggestions, completeSubtree(root, pos.Children, remainingArgs, branchState)...)
	}

	return dedupSort(suggestions)
}

// cloneState returns a shallow copy of state, so a completion branch can
// tentatively extend it without affecting sibling branches explored from
// the same starting state.
func cloneState(state OutData) OutData {
	clone := make(OutData, len(state))
	for k, v := range state {
		clone[k] = v
	}
	return clone
}

// dedupSort returns values deduplicated and sorted, or nil if values is
// empty - matching Complete's documented "nil means nothing to suggest".
func dedupSort(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(values))
	out := make([]string, 0, len(values))
	for _, v := range values {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	sort.Strings(out)
	return out
}
func MakeArgTypeLiteral(value string) ArgType {
	return ArgType{
		Name: "Keyword",
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
func MakeArgTypeAny(values []string) ArgType {
	return ArgType{
		Name: "AnyOf",
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

// viableSuggestions collects List() values from every possibility whose If
// (if any) passes against state, deduplicated and sorted.
func viableSuggestions(possibilities []ArgPossibility, state OutData) []string {
	seen := make(map[string]bool)
	var out []string
	for _, p := range possibilities {
		if p.If != nil && !p.If(state) {
			continue
		}
		if p.Type.List == nil {
			continue
		}
		for _, v := range p.Type.List() {
			if !seen[v] {
				seen[v] = true
				out = append(out, v)
			}
		}
	}
	sort.Strings(out)
	return out
}

// printCompletions writes Complete's result one per line to stdout, which
// is the format zsh's compadd (fed via command substitution) expects.
func printCompletions(tree []ArgPossibility, args []string) {
	for _, s := range Complete(tree, args) {
		fmt.Println(s)
	}
}
func CheckCompletionRequest(tree []ArgPossibility) bool {
	if len(os.Args) > 1 && os.Args[1] == "__complete" {
		printCompletions(tree, os.Args[2:])
		return true
	}
	return false
}

// ShowHelp prints a human-readable rendering of the whole argument tree to
// stdout: one indented line per possibility, showing its name, the values
// its type accepts (via Type.List), whether it repeats (EndAction ==
// EndActionLoop), and - if set - IfDescription for conditional nodes.
//
// Note: this calls every reachable Type.List(), so a List that panics
// (e.g. the "TODO" placeholders in this package) will panic ShowHelp too -
// that's intentional: it's telling you that type needs a real List before
// it can be documented, rather than quietly leaving it out of the help
// output.
func ShowHelp(tree []ArgPossibility) {
	fmt.Print(BuildHelp(tree))
}

// BuildHelp is the same rendering ShowHelp prints, returned as a string
// instead of written to stdout.
func BuildHelp(tree []ArgPossibility) string {
	var b strings.Builder
	writeHelpLevel(&b, tree, "")
	return b.String()
}

// writeHelpLevel renders one level of the tree using box-drawing connectors
// (├─, └─, │) the way `tree`/`ls -R` style output does, so a possibility's
// place in the branching structure is visible at a glance. prefix is the
// exact string to print before each line at this depth, already carrying
// the "│  " / "   " continuation from every ancestor level.
// ANSI color codes used by ShowHelp/BuildHelp to visually separate the
// different kinds of information on each line: the tree connectors, the
// <name> label, the accepted-values list, the (repeats) marker, and the
// [only if ...] condition.
const (
	ansiReset      = "\033[0m"
	colorConnector = "\033[90m"   // dim gray
	colorLabel     = "\033[1;36m" // bold cyan
	colorValues    = "\033[32m"   // green
	colorSep       = "\033[31m"   // ?
	colorTypeName  = "\033[34m"   // ?
	colorRepeat    = "\033[33m"   // yellow
	colorCondition = "\033[35m"   // magenta
)

func writeHelpLevel(b *strings.Builder, possibilities []ArgPossibility, prefix string) {
	for i, p := range possibilities {
		isLast := i == len(possibilities)-1

		connector := "├─ "
		childPrefix := prefix + "│  "
		if isLast {
			connector = "└─ "
			childPrefix = prefix + "   "
		}

		b.WriteString(colorConnector)
		b.WriteString(prefix)
		b.WriteString(connector)
		b.WriteString(ansiReset)
		b.WriteString(describePossibility(p))
		b.WriteString("\n")

		if len(p.Children) > 0 {
			writeHelpLevel(b, p.Children, childPrefix)
		}
	}
}

// maxHelpListItems caps how many of a Type's List() values ShowHelp will
// print inline. List exists for tab-completion, where showing everything is
// the point; help text is read by a human, so a type with hundreds of valid
// values (like key names) needs to be summarized instead of dumped in full.
const maxHelpListItems = 8

func describePossibility(p ArgPossibility) string {
	label := p.Name

	parts := []string{}
	var part string = fmt.Sprintf("%s<%s", colorLabel, label)
	// parts := []string{fmt.Sprintf("%s<%s>%s", colorLabel, label, ansiReset)}

	if p.Type.Name == "" {
		part += fmt.Sprintf(">%s", ansiReset)
	} else {
		if label == "" {
			part += fmt.Sprintf("%s%s%s%s>%s", colorSep, colorTypeName, p.Type.Name, colorLabel, ansiReset)
		} else {
			part += fmt.Sprintf("%s:%s%s%s>%s", colorSep, colorTypeName, p.Type.Name, colorLabel, ansiReset)
		}
	}
	if p.Type.List != nil {
		if values := p.Type.List(); len(values) > 0 {
			var joiner string = fmt.Sprintf("%s, %s", colorSep, colorValues)
			if len(values) == 1 {
				part += fmt.Sprintf("%s:%s %s%s", colorSep, colorValues, values[0], ansiReset)
				// part += fmt.Sprintf("%skeyword%s:%s %s%s", colorTypeName, colorSep, colorValues, values[0], ansiReset)
			} else if len(values) > maxHelpListItems {
				shown := strings.Join(values[:maxHelpListItems], joiner)
				part += fmt.Sprintf("%s:%s %s%s, ...%s (%d total)%s", colorSep, colorValues, shown, colorSep, colorValues, len(values), ansiReset)
				// part += fmt.Sprintf("%sone of%s:%s %s%s, ...%s (%d total)%s", colorTypeName, colorSep, colorValues, shown, colorSep, colorValues, len(values), ansiReset)
			} else {
				part += fmt.Sprintf("%s:%s %s%s", colorSep, colorValues, strings.Join(values, joiner), ansiReset)
				// part += fmt.Sprintf("%sone of%s:%s %s%s", colorTypeName, colorSep, colorValues, strings.Join(values, joiner), ansiReset)
			}
		}
	}

	parts = append(parts, part)
	if p.EndAction == EndActionLoop {
		parts = append(parts, fmt.Sprintf("%s(repeats)%s", colorRepeat, ansiReset))
	}

	if p.IfDescription != "" {
		parts = append(parts, fmt.Sprintf("%s[only if %s]%s", colorCondition, p.IfDescription, ansiReset))
	}

	return strings.Join(parts, " ")
}

var (
	ArgTypeInt = ArgType{
		Name: "Int",
		Transform: func(s string) (any, error) {
			return s, nil
		},
		List: func() []string {
			return []string{}
		},
	}
	// ArgTypeTime parses standard non-negative duration strings (e.g., "1h3m", "250ms", "1.5h").
	// Rejects negative durations.
	ArgTypeTime = ArgType{
		Name: "Time",
		Transform: func(s string) (any, error) {
			d, err := time.ParseDuration(s)
			if err != nil {
				return nil, fmt.Errorf("invalid time duration format: %w", err)
			}
			if d < 0 {
				return nil, fmt.Errorf("time duration cannot be negative: %s", s)
			}
			return d, nil
		},
		List: func() []string {
			return []string{} // Dynamic type, no static completion list
		},
		Example: func() string {
			return "1h3m"
		},
	}

	// ArgTypeSignedTime parses signed duration strings, explicitly allowing
	// both positive and negative values (e.g., "-2.5ms", "+10s", "1h3m").
	ArgTypeSignedTime = ArgType{
		Name: "SignedTime",
		Transform: func(s string) (any, error) {
			d, err := time.ParseDuration(s)
			if err != nil {
				return nil, fmt.Errorf("invalid signed time duration format: %w", err)
			}
			return d, nil
		},
		List: func() []string {
			return []string{} // Dynamic type, no static completion list
		},
		Example: func() string {
			return "-2.5ms"
		},
	}
)
