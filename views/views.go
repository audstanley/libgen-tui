package views

import (
	"github.com/audstanley/libgen-tui/libgen"

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
	App   *tview.Application
	Pages *tview.Pages
	Flex  *tview.Flex
	Form  *tview.Form
	State *state
}

func New() *Gui {
	return &Gui{
		App:   tview.NewApplication(),
		Pages: tview.NewPages(),
		State: newState(),
	}
}

var info = tview.NewList().AddItem("List item 1", "Some explanatory text", 'a', nil)

func LibGenFlexCreator(g *Gui) {
	g.Flex = tview.NewFlex().
		AddItem(tview.NewBox().SetBorder(true).SetTitle("LibGen Search Results").SetTitleAlign(tview.AlignLeft), 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(tview.NewBox().SetBorder(true).SetTitle("Middle"), 0, 1, false), 0, 2, false).
		AddItem(info, 0, 1, true).
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Action"), 20, 3, false)
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
			search := libgen.New()
			search.Search(searchType, query)
			// fmt.Printf("Form Item: %s", g.Form.GetFormItem(1).(*tview.InputField).GetText())
			g.App.SetRoot(g.Flex, true).Sync().SetFocus(g.Flex)
		}).
		AddButton("Quit", func() {
			g.App.Stop()
		})
}
