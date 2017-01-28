package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	id3 "github.com/mikkyang/id3-go"
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

var overwrite bool

func initFlag() {
	flag.BoolVar(&overwrite, "o", false,
		"overwrite existing tags (default: false)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			"Usage of %s: [options] file...\n", os.Args[0])
		fmt.Fprint(os.Stderr, `
Attempts to look up and apply tag information from BotB to mp3s from
BotB. The filename should be in the original format from the BotB
donload.

Options:
`)
		flag.PrintDefaults()
	}
	flag.Parse()
}

func main() {
	initFlag()
	if flag.NArg() < 1 {
		flag.Usage()
	}

	for _, arg := range flag.Args() {
		// check file & parse filename
		if _, err := os.Stat(arg); err != nil {
			fmt.Fprintf(os.Stderr, "file '%s' does not exist!\n", arg)
			continue
		}
		nicename := strings.Replace(
			strings.SplitN(arg, " - ", 2)[1], ".mp3", "", -1)

		// make API requst
		resp, err := http.Get(
			"http://battleofthebits.org/api/v1/entry/search/" +
				url.QueryEscape(nicename) + "?page_length=1")
		if err != nil {
			fmt.Fprintln(os.Stderr, "api error:", err)
			continue
		}

		// decode API response
		var entries []Entry
		if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
			fmt.Fprintln(os.Stderr, "decoding error:", err)
			fmt.Fprintln(os.Stderr, "probably BotB API is mad")
			continue
		}
		if len(entries) == 0 {
			fmt.Fprintln(os.Stderr, "no metadata found...")
			continue
		} else if len(entries) > 1 {
			fmt.Fprintln(os.Stderr, "ambiguous thingy!")
			continue
		}
		entry := entries[0]

		// tag the file!
		mp3file, err := id3.Open(arg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "id3 open error:", err)
			continue
		}
		defer mp3file.Close()
		mp3file.SetTitle(entry.Title)
		mp3file.SetArtist(entry.BotBr.Name)
		mp3file.SetAlbum(entry.Battle.Title)
		mp3file.SetYear(entry.Datetime[:4])
		mp3file.SetGenre(entry.Format.Title)
	}
}
