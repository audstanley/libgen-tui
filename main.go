package main

import (
	"github.com/audstanley/libgen-tui/views"
)

func main() {
	g := views.New()
	views.LibGenFlexCreator(g)
	views.LibGenSearchFromCreator(g)
	g.Pages.AddPage("Flex", g.Flex, true, true)
	g.Pages.AddPage("Search", g.Form, true, true)

	// Form.SetBorder(true).SetTitle("LibGen Downloader").SetTitleAlign(tview.AlignLeft)
	// item := Form.GetFormItem(0)
	// fmt.Println(item)

	err := g.App.SetRoot(g.Pages, true).EnableMouse(true).Sync().SetFocus(g.Pages).Run()
	if err != nil {
		panic(err)
	}
}
