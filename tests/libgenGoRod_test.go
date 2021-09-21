package tests

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/audstanley/libgen-tui/libgen"
	"github.com/stretchr/testify/assert"
)

// Helper function for md5 checksum tests
func md5HashOfPng(filePath string, t *testing.T) (string, error) {
	var returnMD5String string
	file, err := os.Open(fmt.Sprintf("%s.png", filePath))
	if err != nil {
		t.Logf("Could not open File %s.png", filePath)
		return returnMD5String, err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		t.Log("Some Error with creating md5 hash")
		return returnMD5String, err
	}
	hashInBytes := hash.Sum(nil)[:16]
	returnMD5String = hex.EncodeToString(hashInBytes)
	return returnMD5String, nil
}

// Helper function for md5 checksum tests
func cleanUpPngFiles(filePath string) error {
	err := os.Remove(fmt.Sprintf("%s.png", filePath))
	return err
}

// This will perform a search for the book "Mastering Golang" save the search as png, then perform an md5 checksum
// on the png and test if it's the same as the md5 we have in our tests
func TestSearchForMasteringGoLangBook(t *testing.T) {
	// The md5 of the png for the search is:  9fe7fd0e31a39af5315dfdbda1503e15
	search := libgen.New()
	search.SavePng = true
	// after the search is done, the png will be created
	search.Search("NonFiction", "Mastering Golang")
	md5, _ := md5HashOfPng("NonFiction", t)
	t.Log("md5 length", len(md5))
	assert.Contains(t, []string{"9fe7fd0e31a39af5315dfdbda1503e15"}, md5, "the checksum should be cached")
	cleanUpPngFiles("NonFiction")
}

// This will perform a search for the book "Mastering Regular Expressions" save the search as png, then perform an md5 checksum
// on the png and test if it's the same as the md5 we have in our tests
func TestSearchForMasteringRegularExpressionsBook(t *testing.T) {
	// The md5 of the png for the search is: 8c06b90cc3edbd8cb46e8d079fba3a66
	search := libgen.New()
	search.SavePng = true
	// after the search is done, the png will be created
	search.Search("NonFiction", "Mastering Regular Expressions")
	md5, _ := md5HashOfPng("NonFiction", t)
	t.Log("md5 length", len(md5))
	assert.Contains(t, []string{"8c06b90cc3edbd8cb46e8d079fba3a66", "675bf95b6d2f59dcc11dda5e0df8f497"}, md5, "the checksum should be cached")
	cleanUpPngFiles("NonFiction")
}
