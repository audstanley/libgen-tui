package views

import (
	"os"
	"strconv"

	"github.com/audstanley/libgen-tui/libgen"
	"github.com/gdamore/tcell/v2"
	"github.com/pkg/browser"

	"github.com/rivo/tview"
)

// Gui Struct is basically the top level controller
// for the libgen-tui application
type Gui struct {
	App          *tview.Application
	Pages        *tview.Pages
	Flex         *tview.Flex
	Table        *tview.Table
	Form         *tview.Form
	LibgenSearch *libgen.LibGenSearch
}

// New function is a constructor for the Gui struct
func New() *Gui {
	app := *tview.NewApplication()
	pages := *tview.NewPages()
	flex := *tview.NewFlex()
	table := *tview.NewTable()
	form := *tview.NewForm()
	libgenConstructor := *libgen.New()
	return &Gui{
		App:          &app,
		Pages:        &pages,
		Flex:         &flex,
		Table:        &table,
		Form:         &form,
		LibgenSearch: &libgenConstructor,
	}
}

// TableCreatorAfterSearch will generate a table after an initial search is made.
// When the left or right arrow keys are pressed, new page searches are loaded up.
func (g Gui) TableCreatorAfterSearch() {
	g.Table = tview.NewTable().
		SetFixed(1, 1).
		SetSelectable(true, false)
	// create the table from the search
	layout := []string{"", "Author", "Title", "Publisher", "Language", "Year", "Pages", "Format / Size"}
	for i, book := range g.LibgenSearch.GetWebPageOfBooksStruct().Books {
		layout = append(layout, strconv.Itoa(i+1+(25*(int(g.LibgenSearch.GetWebPageOfBooksStruct().CurrentPageNumber)-1))))
		if len(book.Author) > 31 {
			layout = append(layout, book.Author[0:31]+"...")
		} else {
			layout = append(layout, book.Author)
		}
		if len(book.Title) > 47 {
			layout = append(layout, book.Title[0:47]+"...")
		} else {
			layout = append(layout, book.Title)
		}
		if len(book.Publisher) > 31 {
			layout = append(layout, book.Publisher[0:31]+"...")
		} else {
			layout = append(layout, book.Publisher)
		}
		layout = append(layout, book.Language)
		layout = append(layout, book.Year)
		layout = append(layout, book.Pages)
		layout = append(layout, book.FormatAndSize)
	}
	cols, rows := 8, len(g.LibgenSearch.GetWebPageOfBooksStruct().Books)+1
	word := 0
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			color := tcell.ColorWhite
			if c < 1 || r < 1 {
				color = tcell.ColorYellow
			}
			g.Table.SetCell(r, c,
				tview.NewTableCell(layout[word]).
					SetTextColor(color).
					SetAlign(tview.AlignLeft))
			word = (word + 1) % len(layout)
		}
	}
	g.Table.Select(0, 0).SetFixed(1, 1).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			g.App.SetRoot(g.Pages, true).EnableMouse(true).Sync().SetFocus(g.Pages)
		}
		if key == tcell.KeyEnter {
			g.Table.SetSelectable(true, true)
		}
	}).SetSelectedFunc(func(row int, column int) {
		g.Table.GetCell(row, column).SetTextColor(tcell.ColorRed)
		g.Table.SetSelectable(true, false)

		//get link
		downloadLink := g.LibgenSearch.GetWebPageOfBooksStruct().Books[row].DownloadLink
		bookName := g.LibgenSearch.GetWebPageOfBooksStruct().Books[row].Title
		description := g.LibgenSearch.GetWebPageOfBooksStruct().Books[row].Description
		modal := tview.NewModal().
			SetText("Do you want to download \"" + bookName + "\"?\n\n" + description).
			AddButtons([]string{"Cancel", "Download"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {

				//if not then close the modal
				if buttonLabel == "Cancel" {
					g.App.SetRoot(g.Table, true).Sync().SetFocus(g.Table)
				}
				//if yes then download the book
				if buttonLabel == "Download" {
					browser.OpenURL(downloadLink)

				}
			})
		if err := g.App.SetRoot(modal, false).SetFocus(modal).Run(); err != nil {
			panic(err)
		}

	}).SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		switch event.Key() {
		// Listen for Right Arrow Key
		case tcell.KeyRight:
			// next page search
			g.LibgenSearch.NextPageUpdate()
			g.LibgenSearch.ClearAllDataForNextPageSearch()
			channelLength := g.LibgenSearch.Search(*g.LibgenSearch.SearchType, g.LibgenSearch.GetTitle())
			channelOfBooks := make(chan libgen.WebPageOfBooks, channelLength)
			g.LibgenSearch.SaveDownloadLinkToLog(strconv.Itoa(channelLength))
			g.LibgenSearch.GetBookDescription(channelOfBooks)

			for i := 0; i < channelLength; i++ {
				<-channelOfBooks
				g.TableCreatorAfterSearch()
				g.App.SetRoot(g.Table, true).Sync().SetFocus(g.Table)
			}
			g.LibgenSearch.SaveWebPageOfBooks()
		// Listen for Left Arrow Key
		case tcell.KeyLeft:
			// previous page search
			g.LibgenSearch.PreviousPageUpdate()
			g.LibgenSearch.ClearAllDataForNextPageSearch()
			channelLength := g.LibgenSearch.Search(*g.LibgenSearch.SearchType, g.LibgenSearch.GetTitle())
			channelOfBooks := make(chan libgen.WebPageOfBooks, channelLength)
			g.LibgenSearch.SaveDownloadLinkToLog(strconv.Itoa(channelLength))
			g.LibgenSearch.GetBookDescription(channelOfBooks)

			for i := 0; i < channelLength; i++ {
				<-channelOfBooks
				g.TableCreatorAfterSearch()
				g.App.SetRoot(g.Table, true).Sync().SetFocus(g.Table)
			}
			g.LibgenSearch.SaveWebPageOfBooks()
		}
		return event
	})
	if err := g.App.SetRoot(g.Table, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func (g Gui) LibGenSearchFormCreator() {
	g.Form.
		AddDropDown("Genre", []string{"NonFiction", "Fiction", "Scientific"}, 0, nil).
		AddInputField("Search", "", 20, nil, nil).
		AddButton("Search", func() {
			g.LibgenSearch = libgen.New()
			_, searchType := g.Form.GetFormItem(0).(*tview.DropDown).GetCurrentOption()
			query := g.Form.GetFormItem(1).(*tview.InputField).GetText()
			channelLength := g.LibgenSearch.Search(searchType, query)
			// Create a buffered channel for the Page of books
			channelOfBooks := make(chan libgen.WebPageOfBooks, channelLength)
			g.LibgenSearch.SaveDownloadLinkToLog(strconv.Itoa(channelLength))
			go g.LibgenSearch.GetBookDescription(channelOfBooks)

			for i := 0; i < channelLength; i++ {
				<-channelOfBooks
				g.TableCreatorAfterSearch()
				g.App.SetRoot(g.Table, true).Sync().SetFocus(g.Table)
			}
			g.LibgenSearch.SaveWebPageOfBooks()
		}).
		AddButton("Quit", func() {
			g.App.Stop()
			os.Exit(1)
		})
}
