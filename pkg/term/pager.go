package term

import (
	"fmt"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// PageItem represents a single item to be viewed
type PageItem struct {
	Name    string
	Content string
}

// Page shows an application viewer allowing the review of specific resources
func Page(items []PageItem) error {
	if len(items) == 0 {
		fmt.Println("No resources found")
		return nil
	}
	app := tview.NewApplication()

	// convert color codes
	for i, p := range items {
		p.Name = tview.TranslateANSI(p.Name)
		p.Content = tview.TranslateANSI(p.Content)
		items[i] = p
	}

	// right side: text view
	text := tview.NewTextView().SetText(items[0].Content).SetDynamicColors(true)
	text.Box = text.Box.SetBorder(true).SetTitle("Resource")

	// left side: resource chooser
	list := tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true)
	list.Box = list.Box.SetBorder(true).SetTitle("Resources").SetBorderAttributes(tcell.AttrBold)

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
		AddItem(list, 0, 1, true).
		AddItem(text, 0, 4, false)

	isListSelected := true

	// custom key handler
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyDown, tcell.KeyUp:
			if isListSelected {
				list.InputHandler()(event, nil)
				selectItem(list.GetCurrentItem())
			} else {
				text.InputHandler()(event, nil)
			}
		case tcell.KeyTAB, tcell.KeyRight:
			isListSelected = false
			app.SetFocus(text)
		case tcell.KeyBacktab, tcell.KeyLeft:
			isListSelected = true
			app.SetFocus(list)
		case tcell.KeyCtrlC, tcell.KeyEscape:
			return tcell.NewEventKey(tcell.KeyCtrlC, event.Rune(), event.Modifiers())
		default:
			text.InputHandler()(event, nil)
		}

		return nil
	})

	return app.SetRoot(flex, true).EnableMouse(true).Run()
}
