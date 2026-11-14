package brew

import (
	"strings"
	"testing"
)

func TestStreamBound(t *testing.T) {
	var size int
	w := streamWriter{emit: func(s string) { size += len(s) }}
	p := make([]byte, 2*1024*1024)
	n, e := w.Write(p)
	if e != nil || n != len(p) || size != len(p) || len(w.data) != 1024*1024 {
		t.Fatal("unbounded transcript")
	}
}

func TestStreamPreservesLateDiagnostics(t *testing.T) {
	var emitted strings.Builder
	w := streamWriter{emit: func(s string) { emitted.WriteString(s) }}
	prefix := strings.Repeat("x", 1024*1024)
	w.Write([]byte(prefix))
	const diagnostic = "\nERROR: final build diagnostic\n"
	w.Write([]byte(diagnostic))
	if got := emitted.String(); got != prefix+diagnostic {
		t.Fatal("late output was not forwarded")
	}
	if len(w.data) != 1024*1024 || !strings.HasSuffix(string(w.data), diagnostic) {
		t.Fatal("bounded transcript lost the final diagnostic")
	}
}
