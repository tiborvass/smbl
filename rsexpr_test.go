package sexpr

import (
	"strings"
	"testing"
)

func TestParseReadable(t *testing.T) {
	tests := [][2]string{
		{`a`, `(a)`},
		{`a b`, `(a b)`},
		{`
a
b`,
			`(a b)`},
		{`
a b
c`,
			`((a b) c)`},
		{`
a
		b
c
			d
				e
			f
	g`,
			`((a (b)) (c (((d e) f)) g))`},
		{`
a
	b c`,
			`(a (b c))`},
		{`
a
	b
		c d e
		f
	g
h`,
			`((a (b (c d e) f) g) h)`,
		},
		{
			`
a
		b c d
		e
				f
				g
			h
	i`,
			`(a ((b c d) (e (f g) h)) i)`,
		},
		{
			`
a
	b
		c d e
		f
g`,
			`((a (b (c d e) f)) g)`,
		},
		{
			`
a
	b c d
		e
	f
g`,
			`((a ((b c d) (e) f)) g)`,
		},
	}

	for _, test := range tests[len(tests)-1:] {
		expr, err := ParseReadable(strings.NewReader(test[0]))
		t.Logf("Testing: %s", test[1])
		if (err == nil) == (test[1] == "invalid") {
			t.Error(err)
		}
		if expr.SExpr() != test[1] {
			t.Log("FAILED")
			t.Logf("%s", expr)
			t.Logf("%q", expr.SExpr())
			t.Log("got:     ", expr.SExpr())
			t.Errorf("expected: %s", test[1])
		} else {
			t.Log("PASSED")
		}
		/*if expr.String() != test[0] {
			t.Log("FAILED")
			t.Logf("%p %#v", expr, expr)
			t.Logf("got:     %q", expr.String())
			t.Errorf("expected: %q", test[0])
		}*/
		t.Log("-------------")
	}
}
