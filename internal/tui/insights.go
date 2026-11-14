package tui

import (
	"brew-sentinel/internal/disk"
	"brew-sentinel/internal/security"
	"context"
	"fmt"
)

// scanAfter runs without a confirmation: startup scanning is part of the app's normal use.
func (u *UI) scan() { u.scanAfter(nil) }
func (u *UI) scanAfter(done func()) {
	if u.busy || u.loading || u.scanning {
		return
	}
	u.report = nil
	u.scanError = ""
	u.scanning = true
	u.render()
	u.startInsight("security", func(ctx context.Context) (func(), error) {
		r, err := u.client.Run(ctx, "vulns", "--json")
		var report security.Report
		if err == nil {
			report, err = security.Parse(r.Stdout)
		}
		return func() {
			u.scanning = false
			if err != nil {
				u.scanError = fmt.Sprintf("%s: %s", u.errorText(err), r.Stderr)
				u.status = u.tr("Scan unavailable · v to retry")
			} else {
				u.report = &report
				u.status = fmt.Sprintf(u.tr("Scan complete · %d formulae checked · v to rescan"), report.Checked)
			}
			u.render()
			if done != nil {
				done()
			}
		}, nil
	})
}
func (u *UI) measure() {
	if u.busy || u.loading || u.measuring {
		return
	}
	u.measuring = true
	snapshot := u.snapshot
	u.startInsight("disk", func(ctx context.Context) (func(), error) {
		sizes := map[string]int64{}
		failed := 0
		for _, p := range snapshot.Packages {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			if p.Kind != "formula" {
				continue
			}
			var paths []string
			for _, v := range p.Versions {
				if v.KegPath != "" {
					paths = append(paths, v.KegPath)
				}
			}
			if len(paths) == 0 {
				failed++
				continue
			}
			n, e := disk.Measure(ctx, paths)
			if e != nil {
				failed++
				continue
			}
			sizes[p.ID] = n
		}
		return func() {
			u.sizes = sizes
			u.status = fmt.Sprintf(u.tr("Disk measured · %d formulae · %d unavailable · casks and cache excluded"), len(sizes), failed)
			u.render()
		}, nil
	})
}
