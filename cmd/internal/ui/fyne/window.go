// ui/fyne/window.go
package fynerenderer

import "fyne.io/fyne/v2"

/*type ExecutionWindow interface {
	SetTitle(string)
	SetContent(fyne.CanvasObject)
	Show()
	Close()
}*/

type FyneWindow struct {
	w fyne.Window
}

func (fw *FyneWindow) SetTitle(t string)              { fw.w.SetTitle(t) }
func (fw *FyneWindow) SetContent(c fyne.CanvasObject) { fw.w.SetContent(c) }
func (fw *FyneWindow) Show()                          { fw.w.Show() }
func (fw *FyneWindow) Close()                         { fw.w.Close() }
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
