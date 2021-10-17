package libgen

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-rod/rod"
)

type LibGenSearch struct {
	NonFiction      *string
	nonFictionTitle *string
	Fiction         *string
	fictionTitle    *string
	Scientific      *string
	scientificTitle *string
	Rod             *rod.Browser
	pages           *[]string
	pageLinks       *[]string
	SearchType      *string
	SavePng         bool
	DoSearch        bool
}

func New() *LibGenSearch {
	nf := ""
	nft := ""
	f := ""
	ft := ""
	s := ""
	st := ""
	Rod := rod.New()
	// if runtime.GOOS == "linux" {
	// 	var uname syscall.Utsname
	// }

	p := []string{}
	pl := []string{}
	SearchType := ""
	SavePng := false
	DoSearch := true
	return &LibGenSearch{
		NonFiction:      &nf,
		nonFictionTitle: &nft,
		Fiction:         &f,
		fictionTitle:    &ft,
		Scientific:      &s,
		scientificTitle: &st,
		Rod:             Rod,
		pages:           &p,
		pageLinks:       &pl,
		SearchType:      &SearchType,
		SavePng:         SavePng,
		DoSearch:        DoSearch,
	}
}

// This will check the system, if it is running a test, if so, output searches to png file.
func (search LibGenSearch) isTestRun() bool {
	return strings.HasSuffix(os.Args[0], ".test")
}

// the rod seach/scrape will be different, since the layout of each search type is different on the website
func (search LibGenSearch) nonFictionSearchString(q string) {
	spacesArePlus := strings.ReplaceAll(q, " ", "+")
	*search.nonFictionTitle = spacesArePlus
	*search.NonFiction = fmt.Sprintf("/search.php?req=%s&open=0&res=100&view=simple&phrase=1&column=def", spacesArePlus)
}

// the rod seach/scrape will be different, since the layout of each search type is different on the website
func (search LibGenSearch) fictionSearchString(q string) {
	spacesArePlus := strings.ReplaceAll(q, " ", "+")
	*search.fictionTitle = spacesArePlus
	*search.Fiction = fmt.Sprintf("/fiction/?q=%s", spacesArePlus)
}

// the rod seach/scrape will be different, since the layout of each search type is different on the website
func (search LibGenSearch) scientificSearchString(q string) {
	spacesArePlus := strings.ReplaceAll(q, " ", "+")
	*search.scientificTitle = spacesArePlus
	*search.Scientific = fmt.Sprintf("/scimag/?q=%s", spacesArePlus)
}

// Not quite implemented
// Not that on windows, from some reason, we are unable to write to the filesystem with either GoLang itself.fictionSearch
// OR the go-rod library - We will have to Google this.
func (search LibGenSearch) fictionSearch() {
	if search.DoSearch {
		page := search.Rod.MustConnect().MustPage("http://libgen.rs" + *search.Fiction).MustWindowFullscreen()
		page.MustWaitLoad().MustScreenshot("Fiction.png")
		if search.isTestRun() && search.SavePng {
			page.MustWaitLoad().MustScreenshot("Fiction.png")
		}
		// if we are running the software normally (not as a test)
		if !search.isTestRun() {
			el := page.MustElement("body > table")
			search.saveElmentTextToLog(el.MustHTML())
		}
	}
}

// Not quite implemented
func (search LibGenSearch) nonFictionSearch() {
	if search.DoSearch {
		page := search.Rod.MustConnect().MustPage("http://libgen.rs" + *search.NonFiction).MustWindowFullscreen()
		if search.isTestRun() && search.SavePng {
			page.MustWaitLoad().MustScreenshot("NonFiction.png")
		}
		// if we are running the software normally (not as a test)
		if !search.isTestRun() {
			el := page.MustElement("body > table.c > tbody")
			search.saveElmentTextToLog(el.MustHTML())
		}
	}
}

// Not quite implemented
func (search LibGenSearch) scientificSearch() {
	if search.DoSearch {
		page := search.Rod.MustConnect().MustPage("http://libgen.rs" + *search.Scientific).MustWindowFullscreen()
		if search.isTestRun() && search.SavePng {
			page.MustWaitLoad().MustScreenshot("Scientific.png")
		}
		// if we are running the software normally (not as a test)
		if !search.isTestRun() {
			el := page.MustElement("body > table")
			search.saveElmentTextToLog(el.MustHTML())
		}
	}
}

// When you search a Book, this will save the book table data as an html file
// This is a helper function for if you are tyring to parse the html into
// a GoLang struct
func (search LibGenSearch) saveElmentTextToLog(str string) {
	err := os.WriteFile(fmt.Sprintf("rod-%s-%s.html", *search.SearchType, search.GetTitle()), []byte("\n"), 0644)
	if err != nil {
		log.Fatal(err)
	}
	file, err := os.OpenFile(fmt.Sprintf("rod-%s-%s.html", *search.SearchType, search.GetTitle()), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	if _, err := file.WriteString(str); err != nil {
		log.Fatal(err)
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
	*search.SearchType = searchType
	switch *search.SearchType {
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

func (search LibGenSearch) GetTitle() string {
	switch *search.SearchType {
	case "Fiction":
		return *search.fictionTitle
	case "NonFiction":
		return *search.nonFictionTitle
	case "Scientific":
		return *search.scientificTitle
	default:
		log.Fatal("The defaults for GetTile are [Fiction, NonFiction, and Scientific] not:", *search.SearchType)
		return ""
	}
}
