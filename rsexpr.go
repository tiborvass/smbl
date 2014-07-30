package rsexpr

import (
	"bytes"
	"fmt"
)

type Expr interface {
	fmt.Stringer
	SExpr() string
	str(depth int) string
}

const DEBUG = false

func debug(s string, args ...interface{}) {
	if DEBUG {
		fmt.Printf(s+"\n", args...)
	}
}

type atom string

func (a atom) SExpr() string {
	return string(a)
}

func (a atom) String() string {
	return string(a)
}

func (a atom) str(depth int) string {
	return string(a)
}

type list []Expr

func (l list) SExpr() string {
	var b bytes.Buffer
	b.WriteByte('(')
	if len(l) > 0 {
		b.WriteString(l[0].SExpr())
	}
	for i := 1; i < len(l); i++ {
		b.WriteByte(' ')
		b.WriteString(l[i].SExpr())
	}
	b.WriteByte(')')
	return b.String()
}

func (l list) String() string {
	return l.str(0)
}

func (l list) str(depth int) string {
	var b bytes.Buffer
	t := make([]byte, depth)
	d := 0
	for ; d < depth; d++ {
		t[d] = '\t'
	}
	b.Write(t)
	if len(l) > 0 {
		b.WriteString(l[0].str(d + 1))
	}
	sameLine := true
	for i := 1; i < len(l); i++ {
		if _, ok := l[i].(atom); !ok {
			sameLine = false
			break
		}
	}
	if sameLine {
		for i := 1; i < len(l); i++ {
			b.WriteByte(' ')
			b.WriteString(l[i].str(d + 1))
		}
	} else {
		for i := 1; i < len(l); i++ {
			b.WriteByte('\n')
			b.WriteString(l[i].str(d + 1))
		}
	}
	return b.String()
}

func Encode(data []byte) (expr Expr, err error) {
	defer func() {
		if x := recover(); x != nil {
			var ok bool
			if err, ok = x.(error); ok {
				expr = nil
				return
			}
			panic(x)
		}
	}()
	e, _ := parseColumn(data, list{}, 0, 0)
	return e, nil
}

func parseColumn(data []byte, l list, depth, i int) (Expr, int) {
	debug("parseColumn: i=%d depth=%d l=%v", i, depth, l)
	var line Expr
	d := depth
	for i < len(data) {
		line, i = parseLine(data, list{}, depth, i)
		debug("line=%v", line)
		d = 0
		for i < len(data) && data[i] == '\t' {
			debug("i=%d, c=%q", i, data[i])
			i++
			d++
		}
		debug("%d > %d", d, depth)
		if d > depth {
			if d == depth+1 {
				line, i = parseColumn(data, line.(list), d, i)
			} else {
				var e Expr
				e, i = parseColumn(data, list{}, d, i)
				for n := depth + 2; n < d; n++ {
					e = list{e}
				}
				line = append(line.(list), e)
			}
		}
		if L, ok := line.(list); ok {
			if len(L) == 1 {
				line = L[0]
			}
		}
		l = append(l, line)
		debug("update l = %v", l)
		if d < depth {
			debug("end of parseColumn: i=%d l=%v", i, l)
			return l, i
		}
	}

	debug("End of parseColumn: i=%d l=%v", i, l)
	return l, i
}

func parseLine(data []byte, l list, depth, i int) (Expr, int) {
	debug("parseLine: i=%d depth=%d l=%v", i, depth, l)
	a := i
	word := false
	for i < len(data) {
		debug("i=%d, c=%q", i, data[i])
		switch data[i] {
		case ' ', '\t':
			if word {
				l = append(l, atom(data[a:i]))
				word = false
			}
		case '\n':
			debug("l=%v", l)
			l = append(l, atom(data[a:i]))
			debug("\\n end of parseLine: i=%d l=%v", i+1, l)
			return l, i + 1
		default:
			if !word {
				a = i
				word = true
			}
		}
		i++
	}
	l = append(l, atom(data[a:i]))
	debug("end of parseLine: i=%d l=%v", i, l)
	return l, i
}
