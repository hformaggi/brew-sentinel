package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	bg      = tcell.GetColor("#0b111b")
	surface = tcell.GetColor("#111c2b")
	line    = tcell.GetColor("#293b50")
	ink     = tcell.GetColor("#dce6f3")
	muted   = tcell.GetColor("#8396ad")
	accent  = tcell.GetColor("#67dfb5")
	blue    = tcell.GetColor("#86baff")
	amber   = tcell.GetColor("#f1c079")
	danger  = tcell.GetColor("#ff919b")
)

func panel(b *tview.Box, title string) {
	b.SetBackgroundColor(surface).SetBorder(true).SetBorderColor(line).SetTitleColor(muted).SetTitle(" "+title+" ").SetBorderPadding(0, 0, 1, 1)
}
func textView() *tview.TextView {
	v := tview.NewTextView().SetDynamicColors(true).SetTextColor(ink)
	v.SetBackgroundColor(surface)
	return v
}
func center(p tview.Primitive, width, height int) tview.Primitive {
	row := tview.NewFlex().AddItem(nil, 0, 1, false).AddItem(p, width, 0, true).AddItem(nil, 0, 1, false)
	row.Box = tview.NewBox().SetBackgroundColor(bg)
	col := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(nil, 0, 1, false).AddItem(row, height, 0, true).AddItem(nil, 0, 1, false)
	col.Box = tview.NewBox().SetBackgroundColor(bg)
	return col
}
