package disk

import (
	"context"
	"os"
	"testing"
)

func TestMeasure(t *testing.T) {
	d := t.TempDir()
	os.WriteFile(d+"/test", make([]byte, 8192), 0600)
	n, e := Measure(context.Background(), []string{d, d})
	if e != nil || n <= 0 {
		t.Fatalf("%d %v", n, e)
	}
}
