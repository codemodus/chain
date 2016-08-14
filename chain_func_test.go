package chain

import "testing"

func TestFuncEnd(t *testing.T) {
	c := New(nestedHandler0, nestedHandler0)
	c = c.Append(nestedHandler1, nestedHandler1)

	xc := New(nestedHandler0, nestedHandler0)
	c = c.Merge(xc)

	h := c.EndFn(endHandler)

	w, err := record(h)
	if err != nil {
		t.Fatalf("unexpected error: %s\n", err.Error())
	}

	resp := w.Body.String()
	wResp := b0 + b0 + b1 + b1 + b0 + b0 + bEnd + b0 + b0 + b1 + b1 + b0 + b0
	if resp != wResp {
		t.Fatalf("want response %s, got %s\n", resp, wResp)
	}
}
