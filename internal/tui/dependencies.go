package tui

import (
	"brew-sentinel/internal/inventory"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (u *UI) dependencies(p inventory.Package) {
	root := tview.NewTreeNode(p.Name + " · " + u.sourceLabel(p.DependencySource)).SetReference(p.ID).SetColor(accent)
	tree := tview.NewTreeView().SetRoot(root).SetCurrentNode(root).SetGraphicsColor(line)
	panel(tree.Box, u.tr("Dependencies of ")+p.Name)
	index := map[string]inventory.Package{}
	for _, p := range u.snapshot.Packages {
		index[p.ID] = p
	}
	parents := map[*tview.TreeNode]*tview.TreeNode{}
	nodes := 0
	var expand func(*tview.TreeNode, string, map[string]bool, int)
	expand = func(node *tview.TreeNode, id string, seen map[string]bool, depth int) {
		if seen[id] {
			node.AddChild(tview.NewTreeNode(u.tr("↩ circular relationship")).SetColor(amber))
			return
		}
		if depth >= 20 {
			node.AddChild(tview.NewTreeNode(u.tr("… depth limit reached")).SetColor(muted))
			return
		}
		seen[id] = true
		defer delete(seen, id)
		for _, dep := range u.graph.Dependencies(id) {
			if nodes >= 1000 {
				node.AddChild(tview.NewTreeNode(u.tr("… 1,000-node limit reached")).SetColor(muted))
				break
			}
			nodes++
			label := dep
			color := muted
			if item, ok := index[dep]; ok {
				label = item.Name + "  " + Versions(item)
				color = ink
			} else {
				label += u.tr(" · unresolved")
			}
			child := tview.NewTreeNode(label).SetReference(dep).SetSelectable(true).SetColor(color)
			node.AddChild(child)
			parents[child] = node
			expand(child, dep, seen, depth+1)
			if depth >= 1 {
				child.SetExpanded(false)
			}
		}
	}
	expand(root, p.ID, map[string]bool{}, 0)
	if nodes == 0 {
		root.AddChild(tview.NewTreeNode(u.tr("No direct dependencies recorded.")).SetColor(muted))
	}
	heading := textView().SetText(" [#67dfb5::b]" + tview.Escape(p.Name) + u.tr("[-::-]  [#8396ad]Dependency tree[-]"))
	heading.SetBackgroundColor(bg)
	footer := textView().SetText(u.tr(" [#67dfb5]↑↓[-] browse  [#67dfb5]←[-] collapse  [#67dfb5]→[-] expand  [#67dfb5]Enter[-] details  [#67dfb5]Esc[-] back"))
	footer.SetBackgroundColor(bg)
	rootView := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(heading, 2, 0, false).AddItem(tree, 0, 1, true).AddItem(footer, 2, 0, false)
	rootView.Box = tview.NewBox().SetBackgroundColor(bg)
	close := u.showPage("tree", rootView, tree)
	tree.SetSelectedFunc(func(n *tview.TreeNode) {
		if id, ok := n.GetReference().(string); ok {
			if item, ok := index[id]; ok {
				u.detail(item)
			}
		}
	})
	tree.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		if e.Key() == tcell.KeyEscape {
			close()
			return nil
		}
		if e.Key() == tcell.KeyRight {
			if n := tree.GetCurrentNode(); n != nil {
				children := n.GetChildren()
				if !n.IsExpanded() {
					n.SetExpanded(true)
				} else if len(children) > 0 {
					tree.SetCurrentNode(children[0])
				}
			}
			return nil
		}
		if e.Key() == tcell.KeyLeft {
			if n := tree.GetCurrentNode(); n != nil {
				if n.IsExpanded() && len(n.GetChildren()) > 0 {
					n.SetExpanded(false)
				} else if p := parents[n]; p != nil {
					tree.SetCurrentNode(p)
				}
			}
			return nil
		}
		if e.Rune() == ' ' {
			if n := tree.GetCurrentNode(); n != nil {
				n.SetExpanded(!n.IsExpanded())
			}
			return nil
		}
		return e
	})
}
