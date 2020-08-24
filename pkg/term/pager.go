package term

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type PageItem struct {
	Name    string
	Content string
}

func Page(items []PageItem) error {
	app := tview.NewApplication()

	// convert color codes
	for i, p := range items {
		p.Name = tview.TranslateANSI(p.Name)
		p.Content = tview.TranslateANSI(p.Content)
		items[i] = p
	}

	// right side: text view
	text := tview.NewTextView().SetText(items[0].Content).SetDynamicColors(true)
	text.Box = text.Box.SetBorder(true)

	// left side: resource chooser
	list := tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true)
	list.Box = list.Box.SetBorder(true).SetTitle("Resources")

	selectItem := func(i int) {
		text.SetText(items[i].Content)
		text.ScrollToBeginning()
	}

	list.SetSelectedFunc(func(i int, _ string, _ string, _ rune) {
		selectItem(i)
	})

	for _, i := range items {
		list = list.AddItem(i.Name, "", 0, nil)
	}

	// layout container
	flex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(list, 0, 1, false).
		AddItem(text, 0, 4, true)

	// custom key handler
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTAB, tcell.KeyBacktab:
			list.InputHandler()(event, nil)
			selectItem(list.GetCurrentItem())
		case tcell.KeyCtrlC:
			return event
		default:
			text.InputHandler()(event, nil)
		}

		return nil
	})

	return app.SetRoot(flex, true).EnableMouse(true).Run()
}
