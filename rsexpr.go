/*
Package sexpr provides primitives to parse and format S-expressions with different notations.

In addition to the LISP notation, there is also support for a new Tree notation
that uses whitespace (tabs, spaces and newlines) to define structure.

S-expressions are by recursive definition either an atom (usually a string), or a list of S-expressions.

Example of an S-expression in LISP notation:
	((a b (c (d e f) g)) (h i))

Example of the same S-expression in Tree notation:
	a
		b
		c
			d e f
			g
	h i
*/
package sexpr

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

// SExpr is a node in the S-expression.
// If Children is nil, then it represents an atom, available in Value.
// Otherwise, it represents a list of S-expressions.
// If Value is nil, it is an empty node, which is diff
type SExpr struct {
	Value    *string
	Children []*SExpr
}

// String returns an S-expression in the canonical Readable format.
func (expr *SExpr) String() string {
	//return expr.str(0)
	return fmt.Sprintf("%p %s %#v", expr, expr.SExpr(), expr)
	if expr.Children == nil {
		if expr.Value != nil {
			return *expr.Value
		}
		return ""
	}

	children := make([]string, len(expr.Children))
	for i, child := range expr.Children {
		children[i] = child.String()
	}
	return fmt.Sprintf("(%s)", strings.Join(children, " "))
}

func (expr *SExpr) str(depth int) string {
	return ""
	/*var b bytes.Buffer
	t := make([]byte, depth)
	d := 0
	for ; d < depth; d++ {
		t[d] = '\t'
	}
	b.Write(t)
	if len(expr.Children) > 0 {
		b.WriteString(expr.Children[0].str(d + 1))
	}
	sameLine := true
	for i := 1; i < len(expr.Children); i++ {
		if expr.Children[i].Children == nil {
			sameLine = false
			break
		}
	}
	if sameLine {
		for i := 1; i < len(expr.Children); i++ {
			b.WriteByte(' ')
			b.WriteString(expr.Children[i].str(d + 1))
		}
	} else {
		for i := 1; i < len(expr.Children); i++ {
			b.WriteByte('\n')
			b.WriteString(expr.Children[i].str(d + 1))
		}
	}
	return b.String()*/
}

// SExpr returns an S-expression with the parentheses notation.
func (expr *SExpr) SExpr() string {
	if expr.Children == nil {
		if expr.Value != nil {
			return *expr.Value
		}
		return ""
	}

	var b bytes.Buffer
	b.WriteByte('(')
	if len(expr.Children) > 0 {
		b.WriteString(expr.Children[0].SExpr())
	}
	for i := 1; i < len(expr.Children); i++ {
		b.WriteByte(' ')
		b.WriteString(expr.Children[i].SExpr())
	}
	b.WriteByte(')')
	return b.String()
}

func Parse(reader io.Reader) (root *SExpr, err error) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		inAtom := false
		a := 0
		for i, b := range data {
			switch b {
			case '(', ')':
				if inAtom {
					return i - 1, data[a:i], nil
				}
				return i, []byte{b}, nil
			case ' ', '\t', '\n':
				if inAtom {
					return i - 1, data[a:i], nil
				}
				inAtom = false
			default:
				if !inAtom {
					a = i
				}
				inAtom = true
			}
		}
		// if we're at the end of the reader, return the one long atom
		if atEOF {
			return len(data), data[a:], nil
		}
		// ends while still scanning an atom (the atom might be longer)
		if inAtom {
			return 0, nil, nil
		}
		// no atoms nor parens
		return len(data), []byte{}, nil
	})
	root = new(SExpr)
	parseSExpr(scanner, root)
	return root, nil
}

func parseSExpr(s *bufio.Scanner, expr *SExpr) {
	for s.Scan() {
		tok := s.Bytes()
		if len(tok) == 0 {
			break
		}
		if len(tok) == 1 {
			switch tok[0] {
			case '(':
				if expr.Children == nil && expr.Value != nil {
					expr.Children = append(expr.Children, &SExpr{Value: expr.Value})
					expr.Value = nil
				}
			case ')':
			}
		}

	}
}

// ParseTree parses an S-expression in the Tree notation from reader.
func ParseTree(reader io.Reader) (root *SExpr, err error) {
	defer func() {
		if x := recover(); x != nil {
			if e, ok := x.(error); ok {
				err = e
				return
			}
		}
	}()
	scanner := bufio.NewScanner(reader)
	root = new(SExpr)
	line := parseTree(scanner, nil, root, -1, false)
	if root.Children == nil {
		root = &SExpr{Children: []*SExpr{root}}
	}
	debug("--------")
	if len(line) > 0 {
		return root, fmt.Errorf("Recycled line: %q", line)
	}
	return root, nil
}

const DEBUG = true

func debug(s string, args ...interface{}) {
	if DEBUG {
		fmt.Printf(s+"\n", args...)
	}
}

func parseLine(line []byte) (expr *SExpr, depth int) {
	i := 0
	for i < len(line) && line[i] == '\t' {
		i++
	}
	depth = i
	if i == len(line) {
		return
	}
	atoms := bytes.Fields(line[i:])
	if len(atoms) == 1 {
		s := string(atoms[0])
		expr = &SExpr{Value: &s}
		return
	}
	expr = &SExpr{Children: make([]*SExpr, len(atoms))}
	for n, atom := range atoms {
		s := string(atom)
		expr.Children[n] = &SExpr{Value: &s}
	}
	return
}

func parseTree(s *bufio.Scanner, recycled []byte, expr *SExpr, depth int, wasOneLine bool) []byte {
	debug("expr=%s", expr)
	line := recycled
	for len(line) > 0 || s.Scan() {
		if len(line) == 0 {
			// if no recycled lines, read new lines
			line = s.Bytes()
		}
		if len(line) == 0 {
			// skip empty lines
			continue
		}

		node, lineDepth := parseLine(line)
		debug("lineDepth=%d n=%s", lineDepth, node)
		if lineDepth <= depth {
			// TODO: reuse node instead of line
			// end of tree
			// recycle line
			return line
		}

		if lineDepth > depth+1 {
			// add current node to an empty node that will be added to expr
			// recycle line
			node = new(SExpr)
		} else {
			// add current node to expr
			// no need to recycle anything
			line = nil
		}

		// recursion
		line = parseTree(s, line, node, depth+1, node.Children != nil)

		// atom -> (atom) except if empty node
		if expr.Children == nil && expr.Value != nil {
			expr.Children = append(expr.Children, &SExpr{Value: expr.Value})
			expr.Value = nil
		}

		// append node to (atom): (atom NODE)
		expr.Children = append(expr.Children, node)
		debug("new expr=%s", expr)
	}
	// if one-atom list, then unwrap:
	// (atom) -> atom
	if len(expr.Children) == 1 {
		*expr = *expr.Children[0]
	}
	return nil
}
