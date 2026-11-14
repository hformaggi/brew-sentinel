package tui

import (
	"context"
	"time"
)

// startTask applies results on the UI loop. Only one job may read/change the snapshot.
func (u *UI) startTask(title string, work func(context.Context) (func(), error), args ...any) {
	if !u.ready() {
		return
	}
	u.busy = true
	u.task = title
	u.taskArgs = args
	u.updateChrome()
	ctx, cancel := context.WithTimeout(u.ctx, 5*time.Minute)
	u.cancel = cancel
	go func() {
		defer cancel()
		apply, e := work(ctx)
		u.app.QueueUpdateDraw(func() {
			u.busy = false
			u.cancel = nil
			u.task = ""
			if e != nil {
				u.status = u.tr("Failed: ") + u.errorText(e)
				u.message(u.tr("Could not complete the task"), u.errorText(e))
			} else if apply != nil {
				apply()
			}
			u.updateChrome()
		})
	}()
}

// Insights are independent read-only jobs; mutations/refresh wait for both.
func (u *UI) startInsight(key string, work func(context.Context) (func(), error)) {
	ctx, cancel := context.WithTimeout(u.ctx, 5*time.Minute)
	if u.insightCancels == nil {
		u.insightCancels = map[string]context.CancelFunc{}
	}
	u.insightCancels[key] = cancel
	u.updateChrome()
	go func() {
		defer cancel()
		apply, e := work(ctx)
		u.app.QueueUpdateDraw(func() {
			delete(u.insightCancels, key)
			if key == "disk" {
				u.measuring = false
			} else {
				u.scanning = false
			}
			if e != nil {
				u.status = u.tr("Query unavailable: ") + u.errorText(e)
			} else if apply != nil {
				apply()
			}
			u.render()
		})
	}()
}
