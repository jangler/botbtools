package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	id3 "github.com/bogem/id3v2"
)

// Entry contains tag-relevant data on an entry, provided by the BotB API.
type Entry struct {
	Battle struct {
		Title string `json:"title"`
	} `json:"battle"`
	BotBr struct {
		Name string `json:"name"`
	} `json:"botbr"`
	Datetime string `json:"datetime"`
	Format   struct {
		Title string `json:"title"`
	} `json:"format"`
	Id    string `json:"id"`
	Title string `json:"title"`
}

// initFlag initializes and parses command-line flags.
func initFlag() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			"Usage of %s: file...\n", os.Args[0])
		fmt.Fprint(os.Stderr, `
Attempts to look up and apply tag information from BotB to mp3s from
BotB. The filename should be in the original format from the BotB
donload. Or the first number in it ought to at least be its ID.
`)
		flag.PrintDefaults()
	}
	flag.Parse()
}

// getEntryID returns the first int in a filename, if there is one.
func getEntryID(path string) (int, error) {
	if _, err := os.Stat(path); err != nil {
		return 0, err
	}
	nums := strings.FieldsFunc(filepath.Base(path), func(c rune) bool {
		return c < '0' || c > '9'
	})
	if len(nums) == 0 {
		return 0, fmt.Errorf("bad filename: %s", filepath.Base(path))
	}
	return strconv.Atoi(nums[0])
}

// loadEntry retrieves entry data for the given entry ID using the BotB API.
func loadEntry(id int) (*Entry, error) {
	resp, err := http.Get(
		"http://battleofthebits.org/api/v1/entry/load/" +
			url.QueryEscape(strconv.Itoa(id)))
	if err != nil {
		return nil, err
	}
	var entry Entry
	if err := json.NewDecoder(resp.Body).Decode(&entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

// tagFile applies tag information from an entry to a file at a given path.
func tagFile(path string, entry *Entry) error {
	mp3file, err := id3.Open(path)
	if err != nil {
		return err
	}
	defer mp3file.Close()
	mp3file.SetTitle(entry.Title)
	mp3file.SetArtist(entry.BotBr.Name)
	mp3file.SetAlbum(entry.Battle.Title)
	mp3file.SetYear(entry.Datetime[:4])
	mp3file.SetGenre(entry.Format.Title)
	if err := mp3file.Save(); err != nil {
		return err
	}
	return nil
}

// processFile executes the complete process of tagging a file.
func processFile(path string) error {
	id, err := getEntryID(path)
	if err != nil {
		return err
	}
	entry, err := loadEntry(id)
	if err != nil {
		return err
	}
	if err := tagFile(path, entry); err != nil {
		return err
	}
	return nil
}

// main is the application entry point. It loops through the filenames provided
// on the command line and attempts to tag each as a BotB entry.
func main() {
	initFlag()
	if flag.NArg() < 1 {
		flag.Usage()
	}

	for _, arg := range flag.Args() {
		if err := processFile(arg); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
