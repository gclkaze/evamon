// ui/fyne/window.go
package fynerenderer

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
)

type FyneWindow struct {
	w fyne.Window

	tabs *container.AppTabs

	mu     sync.Mutex
	tabMap map[string]*FyneTab // <-- key part
}

func NewFyneWindow(w fyne.Window) *FyneWindow {
	return &FyneWindow{
		w:      w,
		tabs:   container.NewAppTabs(),
		tabMap: make(map[string]*FyneTab), // ← REQUIRED
	}
}

type FyneTab struct {
	item *container.TabItem
	log  *widget.Entry
}

func (fw *FyneWindow) SetTitle(t string) { fw.w.SetTitle(t) }
func (fw *FyneWindow) SetContent(draw port.Drawer /*c fyne.CanvasObject*/) {

	content := container.NewMax(
		draw.Root(),
	)
	fw.w.SetContent(content)
}
func (fw *FyneWindow) Show()       { fw.w.Show() }
func (fw *FyneWindow) Close()      { fw.w.Close() }
func (fw *FyneWindow) CommitTabs() { fw.w.SetContent(fw.tabs) }
func (fw *FyneWindow) SetResizable(resizable bool) {
	// Fyne uses fixed-size toggle; resizable = not fixed
	fw.w.SetFixedSize(!resizable)
}

func (fw *FyneWindow) Resize(width, height float32) {
	fw.w.Resize(fyne.NewSize(width, height))
}

func (fw *FyneWindow) RunOnUI(fn func()) {
	app := fyne.CurrentApp()
	if app == nil {
		fn()
		return
	}

	if drv, ok := app.Driver().(interface {
		RunOnMain(func())
	}); ok {
		drv.RunOnMain(fn)
	} else {
		fn()
	}
}

func (fw *FyneWindow) UpsertTab(tabID, title string) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if tab, ok := fw.tabMap[tabID]; ok {
		tab.item.Text = title
		fw.tabs.Refresh()
		return
	}

	// create new tab
	log := widget.NewMultiLineEntry()
	log.Disable()

	item := container.NewTabItem(title, log)
	fw.tabs.Append(item)

	fw.tabMap[tabID] = &FyneTab{
		item: item,
		log:  log,
	}
}

func (fw *FyneWindow) AssignTab(tabID string, draw port.Drawer) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	tab, ok := fw.tabMap[tabID]
	if !ok {
		return
	}

	content := container.NewMax(
		draw.Root(),
	)
	tab.item.Content = content

	fw.tabs.Refresh()
}

func (fw *FyneWindow) RemoveTab(tabID string) {
	fyne.Do(func() {

		fw.mu.Lock()
		tab, ok := fw.tabMap[tabID]
		if ok {
			delete(fw.tabMap, tabID)
		}
		fw.mu.Unlock()

		if !ok {
			return // nothing to remove
		}

		// Remove from AppTabs
		fw.tabs.Remove(tab.item)

		// Optional: ensure a tab is selected if any remain
		if len(fw.tabs.Items) > 0 && fw.tabs.Selected() == nil {
			fw.tabs.Select(fw.tabs.Items[0])
		}
	})
}

func (fw *FyneWindow) AppendLog(tabID, line string) {
	fyne.Do(func() {

		fw.mu.Lock()
		tab, ok := fw.tabMap[tabID]
		fw.mu.Unlock()
		if !ok {
			return
		}

		tab.log.SetText(tab.log.Text + "\n" + line)
	})
}
