package chain

import "testing"

func TestFuncHandlerOrder(t *testing.T) {
	c := New(nestedHandler0, nestedHandler0)
	c = c.Append(nestedHandler1, nestedHandler1)

	mc := New(nestedHandler0, nestedHandler0)
	c = c.Merge(mc)

	h := c.EndFn(endHandler)

	w, err := record(h)
	if err != nil {
		t.Fatalf("unexpected error: %s\n", err.Error())
	}

	resp := w.Body.String()
	wResp := b0 + b0 + b1 + b1 + b0 + b0 + bEnd + b0 + b0 + b1 + b1 + b0 + b0
	if wResp != resp {
		t.Fatalf("want response %s, got %s\n", wResp, resp)
	}
}
