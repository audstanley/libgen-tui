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
	Publisher         string `json:"publisher"`
	Title             string `json:"title"`
	Language          string `json:"language"`
	Year              string `json:"year"`
	Pages             string `json:"pages"`
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
	PobNonFiction      *WebPageOfBooks
	PobScientific      *WebPageOfBooks
	Fiction            *string
	fictionTitle       *string
	Scientific         *string
	scientificTitle    *string
	SearchType         *string
	SavePng            bool
	DoSearch           bool
	PageOfBooksChannel *chan WebPageOfBooks
}

func New() *LibGenSearch {
	nf := ""
	nft := ""
	pobf := &WebPageOfBooks{}
	pobf.CurrentPageNumber = 1
	pobnf := &WebPageOfBooks{}
	pobnf.CurrentPageNumber = 1
	pobs := &WebPageOfBooks{}
	pobs.CurrentPageNumber = 1
	f := ""
	ft := ""
	s := ""
	st := ""
	pobChannel := make(chan WebPageOfBooks)
	// if runtime.GOOS == "linux" {
	// 	var uname syscall.Utsname
	// }

	SearchType := ""
	SavePng := false
	DoSearch := true
	return &LibGenSearch{
		NonFiction:         &nf,
		nonFictionTitle:    &nft,
		PobFiction:         pobf,
		PobNonFiction:      pobnf,
		PobScientific:      pobs,
		Fiction:            &f,
		fictionTitle:       &ft,
		Scientific:         &s,
		scientificTitle:    &st,
		SearchType:         &SearchType,
		SavePng:            SavePng,
		DoSearch:           DoSearch,
		PageOfBooksChannel: &pobChannel,
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
	*search.NonFiction = fmt.Sprintf("/search.php?req=%s&open=0&res=25&view=simple&phrase=1&column=def&page=%s",
		spacesArePlus,
		strconv.Itoa(int(search.PobNonFiction.CurrentPageNumber)))
}

// the rod search/scrape will be different, since the layout of each search type is different on the website
func (search LibGenSearch) fictionSearchString(q string) {
	spacesArePlus := strings.ReplaceAll(q, " ", "+")
	*search.fictionTitle = spacesArePlus
	*search.Fiction = fmt.Sprintf("/fiction/?q=%s&page=%s",
		spacesArePlus,
		strconv.Itoa(int(search.PobFiction.CurrentPageNumber)))
}

// the rod search/scrape will be different, since the layout of each search type is different on the website
func (search LibGenSearch) scientificSearchString(q string) {
	spacesArePlus := strings.ReplaceAll(q, " ", "+")
	*search.scientificTitle = spacesArePlus
	*search.Scientific = fmt.Sprintf("/scimag/?q=%s&page=%s",
		spacesArePlus,
		strconv.Itoa(int(search.PobScientific.CurrentPageNumber)))
}

// Not quite implemented
// Not that on windows, from some reason, we are unable to write to the filesystem with either GoLang itself.fictionSearch
// OR the go-rod library - We will have to Google this.
func (search LibGenSearch) fictionSearch() int {
	// When a fiction search is made - it always returns 25 results, and there is not way to change that.
	if search.DoSearch {
		page := rod.New().MustConnect().MustPage("http://libgen.rs" + *search.Fiction).MustWindowFullscreen()
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
			})
			if errors.Is(err, context.DeadlineExceeded) {
				// this should me a modal - or something.
				fmt.Println("Could not find that book in a reasonable amount of time")
			} else if err != nil {
				log.Fatal(err)
			}
		}
	}
	return len(search.PobFiction.Books)
}

// Not quite implemented
func (search LibGenSearch) nonFictionSearch() int {
	if search.DoSearch {
		page := rod.New().MustConnect().MustPage("http://libgen.rs" + *search.NonFiction).MustWindowFullscreen()
		if search.isTestRun() && search.SavePng {
			page.MustWaitLoad().MustScreenshot("NonFiction.png")
		}
		// if we are running the software normally (not as a test)
		if !search.isTestRun() {
			err := rod.Try(func() {
				tbody := page.Timeout(60 * time.Second).MustSearch("body > table.c > tbody")
				if tbody != nil {
					trs := tbody.MustElements("tr")
					for i, tr := range trs {
						// the first tr is a category bar, ignore.
						if i != 0 {
							tds := tr.MustElements("td")
							book := Book{}
							for j, td := range tds {
								switch j {
								case 0:
									// ID
									// We don't need this
									continue
								case 1:
									book.Author = td.MustText()
								case 2:
									aElements := td.MustElements("a")
									for _, aElement := range aElements {
										link := fmt.Sprint(aElement.MustProperty("href"))
										reg := regexp.MustCompile(`http://libgen.rs/book`)
										matched := reg.Match([]byte(link))
										if matched {
											title := strings.Split(aElement.MustText(), "\n")
											book.Title = title[0]
											book.LinkToDescription = link
										}
									}
								case 3:
									book.Publisher = td.MustText()
									continue
								case 4:
									book.Year = td.MustText()
									book.FormatAndSize = td.MustText()
								case 5:
									book.Pages = td.MustText()
								case 6:
									book.Language = td.MustText()
								case 7:
									book.FormatAndSize = td.MustText()
								case 8:
									book.FormatAndSize += "/" + td.MustText()
									formatThenSize := strings.Split(book.FormatAndSize, "/")
									book.FormatAndSize = formatThenSize[1] + " / " + formatThenSize[0]
								case 9:
									// This is the first mirror URL in the list of td elements
									link := td.MustElement("a")
									book.Mirror1 = fmt.Sprint(link.MustProperty("href"))
								}
							}
							search.PobNonFiction.Books = append(search.PobNonFiction.Books, book)
						}

					}
				}

				// we need to search for div class: catalog_paginator
				// body > div:nth-child(7) > div:nth-child(1)

				// (?P<Results>\d*\W*\d*\W*\d*\W*\d*)  - regex for just under a trillion results.
				// there was some unicode characters therefore \s doesn't work in regex
				// 3 955 files found

				numberOfResultsHtml := page.MustSearch("tbody > tr > td:nth-child(1) > font")
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
						search.PobNonFiction.NumberOfResultsAsString = processedString
						topNumberOfResults, err := strconv.Atoi(processedString)
						search.PobNonFiction.NumberOfResults = uint64(topNumberOfResults)
						topNumberOfResultsAsFloat := float64(topNumberOfResults)
						if err != nil {
							log.Fatal(err)
						}
						search.PobNonFiction.TopPageNumber = uint64(math.Ceil(topNumberOfResultsAsFloat / 25))
						if search.PobNonFiction.TopPageNumber == 0 {
							// if no search results, we still have a top page number of 1.
							search.PobNonFiction.TopPageNumber = 1
						}
						if search.PobNonFiction.CurrentPageNumber == 0 {
							search.PobNonFiction.CurrentPageNumber = 1
						}

					} else {
						search.PobNonFiction.NumberOfResultsAsString = "shoot:" + numberOfResultsHtml.MustHTML()
					}

				} else {
					search.PobNonFiction.NumberOfResultsAsString = "nope"
				}
			})

			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("Could not find that book in a reasonable amount of time")
			} else if err != nil {
				log.Fatal(err)
			}
		}
	}
	return len(search.PobNonFiction.Books)
}

// Not quite implemented
func (search LibGenSearch) scientificSearch() int {
	if search.DoSearch {
		page := rod.New().MustConnect().MustPage("http://libgen.rs" + *search.Scientific).MustWindowFullscreen()
		if search.isTestRun() && search.SavePng {
			page.MustWaitLoad().MustScreenshot("Scientific.png")
		}
		// if we are running the software normally (not as a test)
		if !search.isTestRun() {
			el := page.MustElement("body > table")
			search.saveElementTextToLog(el.MustHTML())
		}
	}
	return len(search.PobScientific.Books)
}

// saveElementTextToLog is a private module that is used for testing by saving some
// of the search data to an html file.
// When you search a Book, this will save the book table data as an html file
// This is a helper function for if you are tyring to parse the html into
// a GoLang struct
func (search LibGenSearch) saveElementTextToLog(str string) {
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

func (search LibGenSearch) SaveDownloadLinkToLog(str string) {
	file, err := os.OpenFile(fmt.Sprintf("rod-%s-%s.log", *search.SearchType, search.GetTitle()), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	if _, err := file.WriteString(str + "\n"); err != nil {
		log.Fatal(err)
	}
}

// The Search module makes a libgen search
// searchType: "Fiction", "NonFiction", "Scientific"
// The second argument is the query
func (search LibGenSearch) Search(searchType string, q string) int {
	if searchType == "" {
		log.Fatal("You are making an empty search to libgen")
		os.Exit(1)
	}
	*search.SearchType = searchType
	switch *search.SearchType {
	case "Fiction":
		search.fictionSearchString(q)
		return search.fictionSearch()
	case "NonFiction":
		search.nonFictionSearchString(q)
		return search.nonFictionSearch()
	case "Scientific":
		search.scientificSearchString(q)
		return search.scientificSearch()
	default:
		log.Fatal("You need to specify the type of libgen search: Libgen.Search(\"[NonFiction, Fiction, or Scientific]\" string, \"Query\")")
		return 0
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

// GetTitle returns the title of the book based on what the search types is.
func (search LibGenSearch) GetWebPageOfBooksStruct() *WebPageOfBooks {
	switch *search.SearchType {
	case "Fiction":
		return search.PobFiction
	case "NonFiction":
		return search.PobNonFiction
	case "Scientific":
		return search.PobScientific
	default:
		return search.PobFiction
	}
}

// SaveWebPageOfBooks will save a json file of the parsed/scraped data
// and save it to a json file.
func (search LibGenSearch) SaveWebPageOfBooks() {
	var pob *WebPageOfBooks
	switch *search.SearchType {
	case "Fiction":
		pob = search.PobFiction
	case "NonFiction":
		pob = search.PobNonFiction
	case "Scientific":
		pob = search.PobScientific
	}

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
func (search LibGenSearch) GetBookDescription(ch chan WebPageOfBooks) {
	switch *search.SearchType {
	case "Fiction":
		search.getBookDescriptionForFiction(ch)
	case "NonFiction":
		search.getBookDescriptionForNonFiction(ch)
	case "Scientific":
		search.getBookDescriptionForScientific(ch)
	}
}

// getBookDescriptionForNonFiction tries to get descriptions for nonfiction books
func (search LibGenSearch) getBookDescriptionForNonFiction(ch chan WebPageOfBooks) {
	wg := sync.WaitGroup{}
	wg.Add(len(search.PobNonFiction.Books))
	for i, book := range search.PobNonFiction.Books {
		go func(differentI int, theBook Book, waitG *sync.WaitGroup) {
			defer waitG.Wait()
			time.Sleep((time.Duration(differentI) * 250) * time.Millisecond)
			page := rod.New().MustConnect().MustPage()
			err := rod.Try(func() {
				page.Timeout(2 * time.Second).MustNavigate(theBook.Mirror1)
				download := page.MustElement("#download > h2 > a")
				search.PobNonFiction.Books[differentI].DownloadLink = fmt.Sprint(download.MustProperty("href"))
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
									search.PobNonFiction.Books[differentI].Description = data
									descFound = false
								}
							}
						}
					}
				}
				search.SaveDownloadLinkToLog(search.PobNonFiction.Books[differentI].Title + " " +
					search.PobNonFiction.Books[differentI].DownloadLink + " " +
					search.PobNonFiction.Books[differentI].Description)
				ch <- *search.PobNonFiction
			})
			if errors.Is(err, context.DeadlineExceeded) {
				search.PobNonFiction.Books[differentI].Description = fmt.Sprintf("outerDescriptionError: %s", err.Error())
			}
		}(i, book, &wg)
		go func() {
			wg.Wait()
			close(ch)
		}()
	}
}

// getBookDescriptionForFiction tries to get descriptions for fiction books
// it also gets the mirror-1 link - for downloading
func (search LibGenSearch) getBookDescriptionForFiction(ch chan WebPageOfBooks) {
	//search.PageOfBooksChannel = make(chan WebPageOfBooks, len(search.PobFiction.Books))
	wg := sync.WaitGroup{}
	wg.Add(len(search.PobFiction.Books))
	for i, book := range search.PobFiction.Books {
		go func(differentI int, theBook Book, waitG *sync.WaitGroup) {
			defer waitG.Wait()
			time.Sleep((time.Duration(differentI) * 250) * time.Millisecond)
			page := rod.New().MustConnect().MustPage()
			err := rod.Try(func() {
				page.Timeout(2 * time.Second).MustNavigate(theBook.Mirror1)

				// #download > h2 > a
				download := page.MustElement("#download > h2 > a")
				search.PobFiction.Books[differentI].DownloadLink = fmt.Sprint(download.MustProperty("href"))
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
				search.SaveDownloadLinkToLog(search.PobFiction.Books[differentI].Title + " " +
					search.PobFiction.Books[differentI].DownloadLink +
					search.PobFiction.Books[differentI].Description)
				ch <- *search.PobFiction
			})

			if errors.Is(err, context.DeadlineExceeded) {
				search.PobFiction.Books[differentI].Description = fmt.Sprintf("outerDescriptionError: %s", err.Error())
			}
		}(i, book, &wg)

		// let's try BUFFERING the channel before we close it.
		// by sending our wait off on a conncurrent goroutine,
		// we can let the function be non-blocking while we listen to the books come into
		// the view.  We can now render the view and once all the requests are finished,
		// we close the channel 'ch' which will allow the receiver to finish and also not block.
		go func() {
			wg.Wait()
			close(ch)
		}()
	}
}

// getBookDescriptionForScientific tries to get descriptions for scientific articles
func (search LibGenSearch) getBookDescriptionForScientific(ch chan WebPageOfBooks) {

}

func (search LibGenSearch) ClearAllDataForNextPageSearch() {
	search.PobFiction.Books = search.PobFiction.Books[:0]
	search.PobNonFiction.Books = search.PobNonFiction.Books[:0]
	search.PobScientific.Books = search.PobScientific.Books[:0]
	*search.PageOfBooksChannel = make(chan WebPageOfBooks)
}

func (search LibGenSearch) NextPageUpdate() {
	switch *search.SearchType {
	case "Fiction":
		if search.PobFiction.CurrentPageNumber < search.PobFiction.TopPageNumber {
			search.PobFiction.CurrentPageNumber += 1
		}
	case "NonFiction":
		if search.PobNonFiction.CurrentPageNumber < search.PobNonFiction.TopPageNumber {
			search.PobNonFiction.CurrentPageNumber += 1
		}
	case "Scientific":
		if search.PobScientific.CurrentPageNumber < search.PobScientific.TopPageNumber {
			search.PobScientific.CurrentPageNumber += 1
		}
	}
}

func (search LibGenSearch) PreviousPageUpdate() {
	switch *search.SearchType {
	case "Fiction":
		if search.PobFiction.CurrentPageNumber > 1 {
			search.PobFiction.CurrentPageNumber -= 1
		}
	case "NonFiction":
		if search.PobNonFiction.CurrentPageNumber > 1 {
			search.PobNonFiction.CurrentPageNumber -= 1
		}
	case "Scientific":
		if search.PobScientific.CurrentPageNumber > 1 {
			search.PobScientific.CurrentPageNumber -= 1
		}
	}
}
