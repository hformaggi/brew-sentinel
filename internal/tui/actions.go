package tui

import (
	"brew-sentinel/internal/brew"
	"brew-sentinel/internal/inventory"
	"context"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strings"
)

func (u *UI) confirm(title, text string, run func()) {
	u.actionDialog("confirm", title, text, []menuItem{{label: u.tr("Cancel")}, {label: u.tr("Confirm"), run: run}})
}
func (u *UI) mutate(action, kind, name string) {
	args, e := brew.ActionArgs(action, kind, name)
	if e != nil {
		u.message(u.tr("Invalid operation"), u.errorText(e))
		return
	}
	u.startTask("Preparing preview for %s…", func(ctx context.Context) (func(), error) {
		fresh, e := u.client.Inventory(ctx)
		if e != nil {
			return nil, e
		}
		preview := ""
		if action != "uninstall" {
			pa, e := brew.PreviewArgs(action, kind, name)
			if e != nil {
				return nil, e
			}
			out, e := u.client.Read(ctx, pa...)
			if e != nil {
				return nil, e
			}
			preview = string(out)
		}
		return func() {
			freshGraph := inventory.NewGraph(fresh)
			text := u.tr("COMMAND\nbrew ") + strings.Join(args, " ") + "\n\n" + preview
			if action == "uninstall" {
				deps := freshGraph.Dependents(kind + ":" + name)
				text += u.tr("KNOWN DEPENDENTS\n")
				if len(deps) == 0 {
					text += u.tr("No direct dependents recorded.\n")
				} else {
					text += strings.Join(deps, "\n") + "\n"
				}
				text += u.tr("\nHomebrew will validate the removal. Dependencies will not be ignored.")
			}
			text += u.tr("\n\nReview the preview before confirming. External changes may affect the result.")
			u.confirm(u.tr("Review · ")+name, text, func() { u.perform(args, kind) })
		}, nil
	}, name)
}
func (u *UI) perform(args []string, kind string) {
	if !u.ready() {
		return
	}
	u.busy = true
	u.task = "Running brew %s"
	u.taskArgs = []any{strings.Join(args, " ")}
	u.updateChrome()
	if kind == "cask" {
		u.app.Suspend(func() {
			e := u.client.Interactive(args)
			if e == nil {
				fmt.Print(u.tr("\nOperation completed."))
			} else {
				fmt.Printf(u.tr("\nFailed: %v"), e)
			}
			fmt.Print(u.tr("\nPress Enter to return…"))
			var input string
			fmt.Scanln(&input)
		})
		u.busy = false
		u.task = ""
		u.refresh()
		return
	}
	output := textView().SetDynamicColors(false).SetScrollable(true)
	panel(output.Box, u.tr("OPERATION · brew ")+strings.Join(args, " "))
	help := textView().SetText(u.tr(" [#f1c079]● Running[-] · scroll output with ↑/↓"))
	help.SetBackgroundColor(bg)
	root := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(output, 0, 1, true).AddItem(help, 2, 0, false)
	root.SetBackgroundColor(bg)
	close := u.showPage("operation", root, output)
	go func() {
		ctx, cancel := context.WithCancel(u.ctx)
		defer cancel()
		r, e := u.client.Stream(ctx, args, func(part string) { u.app.QueueUpdateDraw(func() { fmt.Fprint(output, part); output.ScrollToEnd() }) })
		u.app.QueueUpdateDraw(func() {
			u.busy = false
			u.task = ""
			status := u.tr("[#67dfb5]✓ Completed[-]")
			if e != nil || r.ExitCode != 0 {
				status = u.tr("[#ff919b]! Operation failed[-]")
			}
			help.SetText(" " + status + u.tr(" · [#67dfb5]Esc[-] back · inventory will be refreshed"))
			fmt.Fprintf(output, "\n\n"+u.tr("Exit code: %d"), r.ExitCode)
			if e != nil {
				fmt.Fprintf(output, "\n%s", e)
			}
			root.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
				if ev.Key() == tcell.KeyEscape {
					close()
					return nil
				}
				return ev
			})
			u.refresh()
		})
	}()
}
func (u *UI) cleanup(action string) { u.executeCleanup(action) }
func (u *UI) executeCleanup(action string) {
	if action != "cleanup" && action != "autoremove" {
		return
	}
	u.startTask("Preparing preview for %s…", func(ctx context.Context) (func(), error) {
		out, e := u.client.Read(ctx, action, "--dry-run")
		if e != nil {
			return nil, e
		}
		text := strings.TrimSpace(string(out))
		if text == "" {
			text = u.tr("Homebrew did not list any items to remove.")
		}
		return func() {
			u.confirm(u.tr("Maintenance · ")+action, u.tr("HOMEBREW PREVIEW\n\n")+text+u.tr("\n\nCOMMAND\nbrew ")+action+u.tr("\n\nConfirm only the items you want to remove. The state may change externally."), func() { u.perform([]string{action}, "formula") })
		}, nil
	}, action)
}
