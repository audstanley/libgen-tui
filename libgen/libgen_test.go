package libgen_test

import (
	"strings"
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
	assert.Exactly(t, *search.NonFiction, nonFictionPre+strings.ReplaceAll(query, " ", "+")+nonFictionPost)
}

func TestBookFiction(t *testing.T) {
	search := libgen.New()
	search.DoSearch = false
	query := "Game of Thrones"
	search.Search("Fiction", query)
	assert.Exactly(t, *search.Fiction, fictionPre+strings.ReplaceAll(query, " ", "+"))
}

func TestArticleScientific(t *testing.T) {
	search := libgen.New()
	search.DoSearch = false
	query := "Distributed execution of communicating sequential process-style concurrency: Golang case study"
	search.Search("Scientific", query)
	assert.Exactly(t, *search.Fiction, fictionPre+strings.ReplaceAll(query, " ", "+"))
}
