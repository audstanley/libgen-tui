package libgen_test

import (
	"testing"

	"github.com/audstanley/libgen-tui/libgen"
	"github.com/stretchr/testify/assert"
)

var nonFictionPre string = `/search.php?req=`
var nonFictionPost string = `&open=0&res=100&view=simple&phrase=1&column=def`
var fictionPre string = `/fiction/?q=`
var scientificPre string = `/scimag/?q=`

func TestBookNonFiction(t *testing.T) {
	search := libgen.New()
	search.DoSearch = false
	query := "Mastering Golang"
	search.Search("NonFiction", query)
	t.Log("SEARCH_TYPE", *search.SearchType)
	assert.Exactly(t, *search.NonFiction, nonFictionPre+search.GetTitle()+nonFictionPost)
}

func TestBookFiction(t *testing.T) {
	search := libgen.New()
	search.DoSearch = false
	query := "Game of Thrones"
	search.Search("Fiction", query)
	t.Log("SEARCH_TYPE", *search.SearchType)
	assert.Exactly(t, *search.Fiction, fictionPre+search.GetTitle())
}

func TestArticleScientific(t *testing.T) {
	search := libgen.New()
	search.DoSearch = false
	query := "Distributed execution of communicating sequential process-style concurrency: Golang case study"
	search.Search("Scientific", query)
	t.Log("SEARCH_TYPE", *search.SearchType)
	var expected string = scientificPre + search.GetTitle()
	assert.Exactly(t, *search.Scientific, expected)
}
