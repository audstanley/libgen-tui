package views

import (
	"github.com/audstanley/libgen-tui/libgen"
	"github.com/gdamore/tcell/v2"

	"github.com/rivo/tview"
)

type state struct {
	stopChans map[string]chan int
}

func newState() *state {
	return &state{
		stopChans: make(map[string]chan int),
	}
}

type Gui struct {
	App          *tview.Application
	Pages        *tview.Pages
	Flex         *tview.Flex
	Form         *tview.Form
	State        *state
	LibgenSearch *libgen.LibGenSearch
}

func New() *Gui {
	libgenConstructor := libgen.New()
	return &Gui{
		App:          tview.NewApplication(),
		Pages:        tview.NewPages(),
		State:        newState(),
		LibgenSearch: libgenConstructor,
	}
}

var info = tview.NewList().AddItem("List item 1", "Some explanatory text", 'a', nil)

func LibGenFlexCreator(g *Gui) {
	// a lot of this code can go -
	// I'm not sure how to get a table into a flex, and we have pages that we
	// need to likely lean also.
	// I'm just trying to get some code in here so that we can get going.
	// the search data should now be accessable in the gui struct.
	// the fiction search data "an array of books" is in
	// g.LibgenSearch.PobFiction.Books [this will be an array of books] based on the first page of results.

	modalShown := true
	pages := tview.NewPages()

	g.Flex = tview.NewFlex().
		AddItem(tview.NewBox().SetBorder(true).SetTitle("LibGen Search Results").SetTitleAlign(tview.AlignLeft), 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(tview.NewBox().SetBorder(true).SetTitle("Middle"), 0, 1, false), 0, 2, false).
		AddItem(info, 0, 1, true).
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Action"), 20, 3, false)

	g.Flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if modalShown {
			//nextSlide()
			modalShown = false
		} else {
			pages.ShowPage("modal")
			modalShown = true
		}
		return event
	})
	modal := tview.NewModal().
		SetText("Resize the window to see the effect of the flexbox parameters").
		AddButtons([]string{"Ok"}).SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		pages.HidePage("modal")
	})
	pages.AddPage("flex", g.Flex, true, true).
		AddPage("modal", modal, false, false)

}

func LibGenSearchFromCreator(g *Gui) {
	g.Form = tview.NewForm().
		AddDropDown("Title", []string{"NonFiction", "Fiction", "Scientific"}, 0, nil).
		AddInputField("Search", "", 20, nil, nil).
		AddButton("Search", func() {
			//fmt.Println("Hello World")
			//g.App.SetRoot(, true).Sync().SetFocus(g.)
			//g.State
			_, searchType := g.Form.GetFormItem(0).(*tview.DropDown).GetCurrentOption()
			query := g.Form.GetFormItem(1).(*tview.InputField).GetText()
			g.LibgenSearch.Search(searchType, query)
			// fmt.Printf("Form Item: %s", g.Form.GetFormItem(1).(*tview.InputField).GetText())
			g.App.SetRoot(g.Flex, true).Sync().SetFocus(g.Flex)
		}).
		AddButton("Quit", func() {
			g.App.Stop()
		})
}
