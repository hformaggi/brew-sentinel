package disk

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"brew-sentinel/internal/i18n"
)

func Measure(ctx context.Context, paths []string) (int64, error) {
	var total int64
	var roots []string
	for _, p := range paths {
		if !filepath.IsAbs(p) {
			return 0, i18n.Errorf("path is not absolute")
		}
		p = filepath.Clean(p)
		skip := false
		for _, q := range paths {
			q = filepath.Clean(q)
			if p != q && strings.HasPrefix(p, q+string(filepath.Separator)) {
				skip = true
			}
		}
		for _, q := range roots {
			if p == q {
				skip = true
			}
		}
		if skip {
			continue
		}
		roots = append(roots, p)
		out, e := exec.CommandContext(ctx, "du", "-sk", p).Output()
		if e != nil {
			return 0, e
		}
		f := strings.Fields(string(out))
		if len(f) == 0 {
			return 0, i18n.Errorf("du returned no result")
		}
		n, e := strconv.ParseInt(f[0], 10, 64)
		if e != nil {
			return 0, e
		}
		total += n * 1024
	}
	return total, nil
}
func Format(n int64) string {
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	}
	v := float64(n)
	for _, u := range []string{"KiB", "MiB", "GiB", "TiB"} {
		v /= 1024
		if v < 1024 || u == "TiB" {
			return fmt.Sprintf("%.1f %s", v, u)
		}
	}
	return "?"
}
