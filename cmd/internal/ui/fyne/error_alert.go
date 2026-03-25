package fynerenderer

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type ErrorAlert struct {
	message string
}

func NewErrorAlert(message string) *ErrorAlert {
	return &ErrorAlert{message: message}
}

func (e *ErrorAlert) Show() {
	win := fyne.CurrentApp().NewWindow("Error")

	label := widget.NewLabel(e.message)
	label.Wrapping = fyne.TextWrapWord

	closeBtn := widget.NewButton("Close", func() {
		win.Close()
	})
	closeBtn.Importance = widget.DangerImportance

	content := container.NewBorder(
		nil,
		container.NewHBox(layout.NewSpacer(), closeBtn, layout.NewSpacer()),
		nil, nil,
		label,
	)

	win.SetContent(content)
	win.Resize(fyne.NewSize(400, 200))
	win.CenterOnScreen()
	win.Show()
}
