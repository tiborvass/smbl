package rsexpr

import "testing"

func Test1(t *testing.T) {
	tests := [][2]string{
		{
			`a
	b
		c d e
		f
	g
h`,
			`((a (b (c d e) f) g) h)`,
		},
		{
			`a
		b c d
		e
				f
				g
			h
	i`,
			`(a ((b c d) (e (f g) h)) i)`,
		},
		{
			`a
	b
		c d e
		f
g`,
			`((a (b (c d e) f)) g)`,
		},
		{
			`a
	b c d
		e
	f
g`,
			`invalid`,
		},
	}

	for _, test := range tests {
		expr, err := Encode([]byte(test[0]))
		t.Logf("Testing: %s", test[1])
		if (err == nil) == (test[1] == "invalid") {
			t.Error(err)
		}
		if expr.SExpr() != test[1] {
			t.Log("FAILED")
			t.Log(expr)
			t.Log("got:     ", expr.SExpr())
			t.Errorf("expected: %s", test[1])
		} else {
			t.Log("PASSED")
		}
		t.Log("-------------")
	}
}
