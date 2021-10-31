package libgen

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
)

type Book struct {
	Author            string `json:"author"`
	Subtitle          string `json:"subtitle"`
	Title             string `json:"title"`
	Language          string `json:"language"`
	FormatAndSize     string `json:"formatAndSize"`
	Mirror1           string `json:"mirror1"`
	DownloadLink      string `json:"downloadLink"`
	LinkToDescription string `json:"linkToDescription"`
	Description       string `json:"description"`
}

type WebPageOfBooks struct {
	Books                   []Book `json:"books"`
	CurrentPageNumber       uint64 `json:"currentPageNumber"`
	NumberOfResults         uint64 `json:"NumberOfResults"`
	NumberOfResultsAsString string `json:"NumberOfResultsAsString"`
	TopPageNumber           uint64 `json:"topPageNumber"`
}

type LibGenSearch struct {
	NonFiction         *string
	nonFictionTitle    *string
	PobFiction         *WebPageOfBooks
	Fiction            *string
	fictionTitle       *string
	Scientific         *string
	scientificTitle    *string
	Rod                *rod.Browser
	pages              *[]string
	pageLinks          *[]string
	SearchType         *string
	SavePng            bool
	DoSearch           bool
	PageOfBooksChannel chan WebPageOfBooks
}

func New() *LibGenSearch {
	nf := ""
	nft := ""
	pobf := WebPageOfBooks{}
	f := ""
	ft := ""
	s := ""
	st := ""
	Rod := rod.New()
	pobChannel := make(chan WebPageOfBooks)
	// if runtime.GOOS == "linux" {
	// 	var uname syscall.Utsname
	// }

	p := []string{}
	pl := []string{}
	SearchType := ""
	SavePng := false
	DoSearch := true
	return &LibGenSearch{
		NonFiction:         &nf,
		nonFictionTitle:    &nft,
		PobFiction:         &pobf,
		Fiction:            &f,
		fictionTitle:       &ft,
		Scientific:         &s,
		scientificTitle:    &st,
		Rod:                Rod,
		pages:              &p,
		pageLinks:          &pl,
		SearchType:         &SearchType,
		SavePng:            SavePng,
		DoSearch:           DoSearch,
		PageOfBooksChannel: pobChannel,
	}
}

// This will check the system, if it is running a test, if so, output searches to png file.
func (search LibGenSearch) isTestRun() bool {
	return strings.HasSuffix(os.Args[0], ".test")
}

// the rod search/scrape will be different, since the layout of each search type is different on the website
func (search LibGenSearch) nonFictionSearchString(q string) {
	spacesArePlus := strings.ReplaceAll(q, " ", "+")
	*search.nonFictionTitle = spacesArePlus
	*search.NonFiction = fmt.Sprintf("/search.php?req=%s&open=0&res=100&view=simple&phrase=1&column=def", spacesArePlus)
}

// the rod search/scrape will be different, since the layout of each search type is different on the website
func (search LibGenSearch) fictionSearchString(q string) {
	spacesArePlus := strings.ReplaceAll(q, " ", "+")
	*search.fictionTitle = spacesArePlus
	*search.Fiction = fmt.Sprintf("/fiction/?q=%s", spacesArePlus)
}

// the rod search/scrape will be different, since the layout of each search type is different on the website
func (search LibGenSearch) scientificSearchString(q string) {
	spacesArePlus := strings.ReplaceAll(q, " ", "+")
	*search.scientificTitle = spacesArePlus
	*search.Scientific = fmt.Sprintf("/scimag/?q=%s", spacesArePlus)
}

// Not quite implemented
// Not that on windows, from some reason, we are unable to write to the filesystem with either GoLang itself.fictionSearch
// OR the go-rod library - We will have to Google this.
func (search LibGenSearch) fictionSearch() {
	// When a fiction search is made - it always returns 25 results, and there is not way to change that.
	if search.DoSearch {
		page := search.Rod.MustConnect().MustPage("http://libgen.rs" + *search.Fiction).MustWindowFullscreen()
		page.MustWaitLoad().MustScreenshot("Fiction.png")
		if search.isTestRun() && search.SavePng {
			page.MustWaitLoad().MustScreenshot("Fiction.png")
		}
		// if we are running the software normally (not as a test)
		if !search.isTestRun() {
			err := rod.Try(func() {
				tbody := page.Timeout(6 * time.Second).MustSearch("tbody")
				//pob := WebPageOfBooks{}
				if tbody != nil {
					trs := tbody.MustElements("tr")
					for _, tr := range trs {
						tds := tr.MustElements("td")
						book := Book{}
						for i, td := range tds {
							switch i {
							case 0:
								book.Author = td.MustText()
							case 1:
								book.Subtitle = td.MustText()
							case 2:
								book.Title = td.MustText()
								link := td.MustElement("a")
								book.LinkToDescription = fmt.Sprint(link.MustProperty("href"))
							case 3:
								book.Language = td.MustText()
							case 4:
								book.FormatAndSize = td.MustText()
							case 5:
								// This is the first mirror URL in the list of td elements
								link := td.MustElement("a")
								book.Mirror1 = fmt.Sprint(link.MustProperty("href"))
							}
						}
						search.PobFiction.Books = append(search.PobFiction.Books, book)
					}
				}
				// we need to search for div class: catalog_paginator
				// body > div:nth-child(7) > div:nth-child(1)

				// (?P<Results>\d*\W*\d*\W*\d*\W*\d*)  - regex for just under a trillion results.
				// there was some unicode characters therefore \s doesn't work in regex
				// 3 955 files found
				numberOfResultsHtml := page.MustSearch("body > div:nth-child(7) > div:nth-child(1)")
				if numberOfResultsHtml != nil {
					text := numberOfResultsHtml.MustText()
					reg := regexp.MustCompile(`(?P<Results>\d*\W*\d*\W*\d*\W*\d*)`)
					matched := reg.MatchString(text)
					if matched {
						//names := reg.SubexpNames()
						results := reg.FindStringSubmatch(text)[1]
						// any character that is not 0-9, get rid of. [ie unicode stuff]
						reg2 := regexp.MustCompile("[^0-9]+")
						processedString := reg2.ReplaceAllString(results, "")
						// now we can parse to a digit [finally!]
						search.PobFiction.NumberOfResultsAsString = processedString
						topNumberOfResults, err := strconv.Atoi(processedString)
						search.PobFiction.NumberOfResults = uint64(topNumberOfResults)
						topNumberOfResultsAsFloat := float64(topNumberOfResults)
						if err != nil {
							log.Fatal(err)
						}
						search.PobFiction.TopPageNumber = uint64(math.Ceil(topNumberOfResultsAsFloat / 25))
						if search.PobFiction.TopPageNumber == 0 {
							// if no search results, we still have a top page number of 1.
							search.PobFiction.TopPageNumber = 1
						}
						if search.PobFiction.CurrentPageNumber == 0 {
							search.PobFiction.CurrentPageNumber = 1
						}

					} else {
						search.PobFiction.NumberOfResultsAsString = "shoot:" + numberOfResultsHtml.MustHTML()
					}

				} else {
					search.PobFiction.NumberOfResultsAsString = "nope"
				}

				// Here we can go to each mirror link, and try to get the description of the Book data.
				search.getBookDescription()
				search.SaveWebPageOfBooks(*search.PobFiction)

			})
			if errors.Is(err, context.DeadlineExceeded) {
				// this should me a modal - or something.
				fmt.Println("Could not find that book in a reasonable amount of time")
			} else if err != nil {
				log.Fatal(err)
			}
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

// saveElmentTextToLog is a private module that is used for testing by saving some
// of the search data to an html file.
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

func (search LibGenSearch) saveDwonloadLinkToLog(str string) {
	file, err := os.OpenFile(fmt.Sprintf("rod-%s-%s.log", *search.SearchType, search.GetTitle()), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	if _, err := file.WriteString(str + "\n"); err != nil {
		log.Fatal(err)
	}
}

// GetPages should return a list of pages - not fully implemented just yet
func (search LibGenSearch) GetPages() *[]string {
	return search.pages
}

// GetPageLinks should return a list of page links - not fully implemented just yet
func (search LibGenSearch) GetPageLinks() *[]string {
	return search.pageLinks
}

// The Search module makes a libgen search
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

// GetTitle returns the title of the book based on what the search types is.
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

// SaveWebPageOfBooks will save a json file of the parsed/scraped data
// and save it to a json file.
func (search LibGenSearch) SaveWebPageOfBooks(pob WebPageOfBooks) {
	for _, b := range pob.Books {
		// trim up author new line before saving to file:
		b.Author = strings.Replace(b.Author, "\n", " ", -1)
	}

	jObj, err := json.MarshalIndent(pob, "", "  ")
	if err != nil {
		log.Println(err)
	}

	err = os.WriteFile(fmt.Sprintf("rod-%s-%s.json", *search.SearchType, search.GetTitle()), []byte("\n"), 0644)
	if err != nil {
		log.Fatal(err)
	}
	file, err := os.OpenFile(fmt.Sprintf("rod-%s-%s.json", *search.SearchType, search.GetTitle()), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	if _, err := file.WriteString(string(jObj)); err != nil {
		log.Fatal(err)
	}

}

// getBookDescription switches between search types and will
// call other more specific private modules
func (search LibGenSearch) getBookDescription() {
	switch *search.SearchType {
	case "Fiction":
		search.getBookDescriptionForFiction()
	case "NonFiction":
		search.getBookDescriptionForFiction()
	case "Scientific":
		search.getBookDescriptionForFiction()
	}
}

// getBookDescriptionForNonFiction tries to get descriptions for nonfiction books
func (search LibGenSearch) getBookDescriptionForNonFiction() {

}

// getBookDescriptionForFiction tries to get descriptions for fiction books
// it also gets the mirror-1 link - for downloading
func (search LibGenSearch) getBookDescriptionForFiction() {
	search.PageOfBooksChannel = make(chan WebPageOfBooks)
	var wg sync.WaitGroup
	wg.Add(len(search.PobFiction.Books))
	for i, book := range search.PobFiction.Books {
		go func(differentI int, theBook Book) {
			defer wg.Done()
			time.Sleep((time.Duration(differentI) * 250) * time.Millisecond)
			page := rod.New().MustConnect().MustPage()
			err := rod.Try(func() {
				page.Timeout(5 * time.Second).MustNavigate(theBook.Mirror1)

				// #download > h2 > a
				download := page.MustElement("#download > h2 > a")
				search.PobFiction.Books[differentI].DownloadLink = fmt.Sprint(download.MustProperty("href"))
				search.saveDwonloadLinkToLog(search.PobFiction.Books[differentI].Title + " " + search.PobFiction.Books[differentI].DownloadLink)

				descriptionTbody := page.Timeout(4 * time.Second).MustSearch("tbody")
				if descriptionTbody != nil {
					trs := descriptionTbody.MustElements("tr")
					for _, tr := range trs {
						tds := tr.MustElements("td")
						for _, td := range tds {
							descriptionIndicator := td.MustText()
							splitData := strings.Split(descriptionIndicator, "\n")
							descFound := false
							for _, data := range splitData {
								if data == "Description:" {
									descFound = true
								} else if descFound {
									search.PobFiction.Books[differentI].Description = data
									descFound = false
								}
							}
						}
					}
				}
				search.saveDwonloadLinkToLog(search.PobFiction.Books[differentI].Title + " " + search.PobFiction.Books[differentI].DownloadLink + " : DONE")
			})

			if errors.Is(err, context.DeadlineExceeded) {
				// if rod times out - we'll just crash the app, for now.
				search.PobFiction.Books[differentI].Description = fmt.Sprintf("outerDescriptionError: %s", err.Error())
			} else if err != nil {
				// other errors - we'll just crash the app, for now.
				log.Fatal(err)
			}

			// Send the books to the Page Of Books Channel.
			// This can allow the view to listen to a channel and update the view dynamically as
			// new books come in from the search.
			search.PageOfBooksChannel <- *search.PobFiction

		}(i, book)
	}
	wg.Wait()
	close(search.PageOfBooksChannel)
}

// getBookDescriptionForScientific tries to get descriptions for scientific articles
func (search LibGenSearch) getBookDescriptionForScientific() {

}
