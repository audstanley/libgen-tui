package main

import (
	"github.com/audstanley/libgen-tui/views"
)

func main() {
	g := views.New()
	g.LibGenSearchFormCreator()
	g.Pages.AddPage("Search", g.Form, true, true)
	err := g.App.SetRoot(g.Pages, true).EnableMouse(true).Sync().SetFocus(g.Pages).Run()
	if err != nil {
		panic(err)
	}
}
