// not implemented
package libgen

import "fmt"

type Book struct {
	Author        string
	Subtitle      string
	Title         string
	Language      string
	FormatAndSize string
}

func PrintWebPageOfBooks(books *[]Book) {
	fmt.Printf("WebPage Of Books:\n\n")
	for _, b := range *books {
		fmt.Println(b.Author, b.Title, b.Subtitle, b.Language, b.FormatAndSize)
	}
}
