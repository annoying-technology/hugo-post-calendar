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
		wd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		*hugoSourceDirectory = wd
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

// createCalendarFeed creates an ics string based on all the posts.
func createCalendarFeed(posts []post) (string, error) {
	cal := ics.NewCalendar()

	// If this is set and there are multiple entries in a calendar Calendar.app
	// freaks out: https://github.com/arran4/golang-ical/issues/19. Validators don't catch this...
	//cal.SetMethod(ics.MethodRequest)

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
			event.SetOrganizer("Annoying.Technology")
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

// createCalendarFeed implements this with the ical (github.com/lestrrat-go/ical) library
//func createCalendarFeed(posts []post) (string, error) {
//	c := ical.New()
//	c.AddProperty("SUMMARY", "Post scheduled")
//	tz := ical.NewTimezone()
//	tz.AddProperty("TZID", "Europe/Berlin")
//	c.AddEntry(tz)
//
//	for _, p := range posts {
//		if !(strings.Contains(p.Permalink, "6b47f951d2903438") || strings.Contains(p.Permalink, "885c68a91ffb6b49")) {
//			continue
//		}
//		event := ical.NewEvent()
//		event.AddProperty("uid", p.Permalink)
//		event.AddProperty("summary", "✔︎ Post scheduled")
//		event.AddProperty("description", "A new post is scheduled to appear on annoying.technology.")
//		event.AddProperty("url", p.Permalink)
//
//		_, dtCreated := FormatDateTime("CREATED", p.PublishDate)
//		event.AddProperty("created", dtCreated)
//
//		_, dtStart := FormatDateTime("DTSTART", p.PublishDate)
//		event.AddProperty("dtstart", dtStart)
//
//		// If no end is set it's 1h long
//		//_, dtEnd := FormatDateTime("DTEND", p.PublishDate)
//		//event.AddProperty("dtend", dtEnd)
//		c.AddEntry(event)
//	}
//	return c.String(), nil
//}

//// FormatDateTime as "DTSTART:19980119T070000Z"
//func FormatDateTime(key string, val time.Time) (string, string) {
//	return key, val.UTC().Format("20060102T150405Z")
//}
