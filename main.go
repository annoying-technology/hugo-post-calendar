package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/peterbourgon/ff"

	"github.com/gocarina/gocsv"
	"github.com/gohugoio/hugo/commands"
)

// post contains the hugo CLI output for the list all command.
type post struct {
	Path        string    `csv:"path"`
	Slug        string    `csv:"slug"`
	Title       string    `csv:"title"`
	Date        time.Time `csv:"date"`
	ExpiryDate  time.Time `csv:"expiryDate"`
	PublishDate time.Time `csv:"publishDate"`
	Draft       bool      `csv:"draft"`
	Permalink   string    `csv:"permalink"`
}

func main() {
	fs := flag.NewFlagSet("at", flag.ExitOnError)
	var (
		hugoSourceDirectory = fs.String("source", "", "full path to hugo source directory")
		destinationFilePath = fs.String("destination", "post-schedule.ics", "path and file name of output file")
	)
	ff.Parse(fs, os.Args[1:],
		ff.WithConfigFileParser(ff.PlainParser),
	)

	if *hugoSourceDirectory == "" {
		fmt.Println("hugo source directory can't be empty, specify with --source")
	}

	// All posts should be included, even the ones in the future (flag: "future")
	flags := []string{"list", "all",
		fmt.Sprintf("--source=%s", *hugoSourceDirectory),
	}

	// Using hugo is a library is way harder than just capturing Stdout
	out, err := captureStdout(func() error {
		commands.Execute(flags)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	var posts []post
	if err := gocsv.UnmarshalString(out, &posts); err != nil {
		log.Fatal(err)
	}
	f, err := os.Create(*destinationFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	feedContent, err := createCalendarFeed(posts)
	if err != nil {
		log.Fatal(err)
	}
	_, err = f.WriteString(feedContent)
	if err != nil {
		log.Fatal(err)
	}

}

func createCalendarFeed(posts []post) (string, error) {
	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodRequest)

	for _, p := range posts {
		location, err := time.LoadLocation("Europe/Berlin")
		if err != nil {
			fmt.Println(err)
			continue
		}

		if strings.Contains(p.Permalink, "posts") {
			event := cal.AddEvent(p.Permalink)
			event.SetCreatedTime(p.PublishDate)
			event.SetDtStampTime(p.PublishDate)
			event.SetModifiedAt(p.PublishDate)
			event.SetStartAt(p.PublishDate.In(location))
			//event.SetEndAt(time.Now())
			event.SetSummary(getPrefix(p.Draft) + "✔︎ Post scheduled")
			//event.SetLocation(p.Permalink)
			event.SetDescription("A new post is scheduled to appear on annoying.technology.")
			event.SetURL(p.Permalink)
		}
	}
	return cal.Serialize(), nil
}

func captureStdout(f func() error) (string, error) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String(), err
}

func getPrefix(isDraft bool) string {
	if isDraft {
		return "[Draft] "
	} else {
		return ""
	}
}
