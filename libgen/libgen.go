package libgen

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-rod/rod"
)

type LibGenSearch struct {
	NonFiction *string
	Fiction    *string
	Scientific *string
	pages      *[]string
	pageLinks  *[]string
	SearchType string
	SavePng    bool
	DoSearch   bool
}

func New() *LibGenSearch {
	nf := ""
	f := ""
	s := ""
	p := []string{}
	pl := []string{}
	st := ""
	SavePng := false
	DoSearch := true
	return &LibGenSearch{
		NonFiction: &nf,
		Fiction:    &f,
		Scientific: &s,
		pages:      &p,
		pageLinks:  &pl,
		SearchType: st,
		SavePng:    SavePng,
		DoSearch:   DoSearch,
	}
}

// This will check the system, if it is running a test, if so, output searches to png file.
func (search LibGenSearch) isTestRun() bool {
	return strings.HasSuffix(os.Args[0], ".test")
}

// the rod seach/scrape will be different, since the layout of each search type is different on the website
func (search LibGenSearch) nonFictionSearchString(q string) {
	spacesArePlus := strings.ReplaceAll(q, " ", "+")
	*search.NonFiction = fmt.Sprintf("/search.php?req=%s&open=0&res=100&view=simple&phrase=1&column=def", spacesArePlus)
}

// the rod seach/scrape will be different, since the layout of each search type is different on the website
func (search LibGenSearch) fictionSearchString(q string) {
	spacesArePlus := strings.ReplaceAll(q, " ", "+")
	*search.Fiction = fmt.Sprintf("/fiction/?q=%s", spacesArePlus)
}

// the rod seach/scrape will be different, since the layout of each search type is different on the website
func (search LibGenSearch) scientificSearchString(q string) {
	spacesArePlus := strings.ReplaceAll(q, " ", "+")
	*search.Scientific = fmt.Sprintf("/scimag/?q=%s", spacesArePlus)
}

// Not quite implemented
func (search LibGenSearch) fictionSearch() {
	if search.DoSearch {
		page := rod.New().MustConnect().MustPage("http://libgen.rs" + *search.Fiction).MustWindowFullscreen()
		if search.isTestRun() && search.SavePng {
			page.MustWaitLoad().MustScreenshot("Fiction.png")
		}
	}
}

// Not quite implemented
func (search LibGenSearch) nonFictionSearch() {
	if search.DoSearch {
		page := rod.New().MustConnect().MustPage("http://libgen.rs" + *search.NonFiction).MustWindowFullscreen()
		if search.isTestRun() && search.SavePng {
			page.MustWaitLoad().MustScreenshot("NonFiction.png")
		}
	}
}

// Not quite implemented
func (search LibGenSearch) scientificSearch() {
	if search.DoSearch {
		page := rod.New().MustConnect().MustPage("http://libgen.rs" + *search.Scientific).MustWindowFullscreen()
		if search.isTestRun() && search.SavePng {
			page.MustWaitLoad().MustScreenshot("Scientific.png")
		}
	}
}

func (search LibGenSearch) GetPages() *[]string {
	return search.pages
}

func (search LibGenSearch) GetPageLinks() *[]string {
	return search.pageLinks
}

// searchType: "Fiction", "NonFiction", "Scientific"
// The second argument is the query
func (search LibGenSearch) Search(searchType string, q string) {
	if searchType == "" {
		log.Fatal("You are making an empty search to libgen")
		os.Exit(1)
	}
	search.SearchType = searchType
	switch search.SearchType {
	case "Fiction":
		search.fictionSearchString(q)
		search.fictionSearch()
	case "NonFiction":
		search.nonFictionSearchString(q)
		search.nonFictionSearch()
	case "Scientific":
		search.scientificSearchString(q)
		search.scientificSearch()
	default:
		log.Fatal("You need to specify the type of libgen search: Libgen.Search(\"[NonFiction, Fiction, or Scientific]\" string, \"Query\")")
	}

}
