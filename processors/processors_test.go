package processors

import "testing"

func TestProcess(t *testing.T) {

	p := NewProcessor()

	res := p.Process([]byte{1, 2, 3})

	if resStr := string(res); resStr != NOOP {
		t.Fatal("response not equals to 'NOOP' string")
	}

}
