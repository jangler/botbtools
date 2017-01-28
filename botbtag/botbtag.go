package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	id3 "github.com/bogem/id3v2"
)

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

func initFlag() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			"Usage of %s: file...\n", os.Args[0])
		fmt.Fprint(os.Stderr, `
Attempts to look up and apply tag information from BotB to mp3s from
BotB. The filename should be in the original format from the BotB
donload.
`)
		flag.PrintDefaults()
	}
	flag.Parse()
}

func getEntryID(path string) (string, error) {
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	tokens := strings.Split(path, " ")
	if len(tokens) < 3 {
		return "", fmt.Errorf("bad filename: %s", path)
	}
	return tokens[1], nil
}

func loadEntry(id string) (*Entry, error) {
	resp, err := http.Get(
		"http://battleofthebits.org/api/v1/entry/load/" +
		url.QueryEscape(id))
	if err != nil {
		return nil, err
	}
	var entry Entry
	if err := json.NewDecoder(resp.Body).Decode(&entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

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
