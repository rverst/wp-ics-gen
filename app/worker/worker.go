package worker

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html"
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"com.github.rverst.wp-ics-gen/app/wp"
)

const (
	entryTimePattern = "2006-01-02T15:04:05"
	eventTimePattern = "2006-01-02 15:04:05"
)

var (
	lastHash       [32]byte
	eventLocations = make(map[string]EventLocation)
)

type Worker struct {
	ticker    *time.Ticker
	stop      chan struct{}
	calFile   string
	eventsURL string
	location  *time.Location

	Update chan string
}

func New(wd string, interval time.Duration, eventsUrl string) Worker {
	l, _ := time.LoadLocation("Europe/Berlin")
	return Worker{
		calFile:   wd + "/events.ics",
		eventsURL: eventsUrl,
		ticker:    time.NewTicker(interval),
		stop:      make(chan struct{}),
		location:  l,
		Update:    make(chan string),
	}
}

func (w Worker) StartWorker() {
	go func() {
		ics, hash := loadContent(w.calFile)
		w.updateContent(ics, hash)
		w.generateIcs()
		for {
			select {
			case <-w.stop:
				slog.Info("worker stopped")
				return
			case <-w.ticker.C:
				go w.generateIcs()
			}
		}
	}()
}

func (w Worker) StopWorker() {
	w.ticker.Stop()
	close(w.stop)
	close(w.Update)
}

func (w Worker) generateIcs() {
	data, err := wp.FetchData(w.eventsURL)
	if err != nil {
		fmt.Printf("failed to fetch data: %v\n", err)
		return
	}

	entries := parseData(data, w.location)
	ics, hash := entries.ToIcs()

	w.updateContent(ics, hash)
	saveContent(w.calFile, ics)
}

func (w Worker) updateContent(ics string, hash [32]byte) {
	if bytes.Equal(hash[:], lastHash[:]) {
		return
	}
	fmt.Println("changes detected", hex.EncodeToString(hash[:]), hex.EncodeToString(lastHash[:]))
	w.Update <- ics
	lastHash = hash
}

func parseData(data []map[string]any, location *time.Location) Entries {
	entries := make(Entries, 0)
	for _, item := range data {
		if item["status"].(string) != "publish" {
			continue
		}
		if item["type"].(string) != "events" {
			continue
		}
		acf := item["acf"].(map[string]any)
		if len(acf) == 0 || acf["from_date"] == nil {
			continue
		}

		created, err := time.ParseInLocation(entryTimePattern, item["date"].(string), location)
		if err != nil {
			fmt.Printf("failed to parse created date: %v\n", err)
			continue
		}
		modified, err := time.ParseInLocation(entryTimePattern, item["modified"].(string), location)
		if err != nil {
			modified = created
		}

		fromDate, err := time.ParseInLocation(eventTimePattern, acf["from_date"].(string), location)
		if err != nil {
			fmt.Printf("failed to parse from date: %v\n", err)
			continue
		}
		toDate, _ := time.ParseInLocation(eventTimePattern, acf["to_date"].(string), location)

		allDay := false
		if fromDate.Hour() == 0 && fromDate.Minute() == 0 && toDate.Hour() == 0 && toDate.Minute() == 0 {
			allDay = true
		}

		wpTerm := item["_links"].(map[string]any)["wp:term"].([]any)
		locUrl := ""
		for _, term := range wpTerm {
			t := term.(map[string]any)
			if t == nil || t["taxonomy"].(string) != "eventloc" {
				continue
			}
			locUrl = t["href"].(string)

			if _, ok := eventLocations[locUrl]; !ok {
				locData, err := wp.FetchData(locUrl)
				if err != nil {
					fmt.Printf("failed to fetch location data: %v\n", err)
					continue
				}
				eventLocations[locUrl] = parseEventLocation(locData)
			}
		}

		entry := Entry{
			Guid:     guid(item["id"].(float64)),
			Created:  created,
			Modified: modified,
			FromDate: fromDate,
			ToDate:   toDate,
			AllDay:   allDay,
			URL:      item["link"].(string),
			Title:    item["title"].(map[string]any)["rendered"].(string),
			Content:  stripHtml(item["content"].(map[string]any)["rendered"].(string)),
			Excerpt:  stripHtml(item["excerpt"].(map[string]any)["rendered"].(string)),
		}
		if locUrl != "" {
			entry.Location = eventLocations[locUrl]
		}
		entries = append(entries, entry)
	}
	return entries
}

func parseEventLocation(data []map[string]any) EventLocation {
	if len(data) == 0 {
		return EventLocation{}
	}

	return EventLocation{
		ID:          int64(data[0]["id"].(float64)),
		Name:        data[0]["name"].(string),
		Description: data[0]["description"].(string),
	}
}

func stripHtml(s string) string {
	s = strings.ReplaceAll(s, "</p>", "\n")
	s = strings.ReplaceAll(s, "<br>", "\n")
	s = strings.ReplaceAll(s, "<br/>", "\n")
	s = strings.ReplaceAll(s, "<br />", "\n")

	re := regexp.MustCompile(`<[^>]*>`)
	s = re.ReplaceAllString(s, "")

	re = regexp.MustCompile(`\n\s*\n+`)
	s = re.ReplaceAllString(s, "\n")

	s = strings.Trim(s, "\n")
	return html.UnescapeString(s)
}

func guid(f float64) string {
	sum := sha256.Sum256([]byte(strconv.Itoa(int(f))))
	return fmt.Sprintf("%x", sum)
}

func loadContent(path string) (string, [32]byte) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("failed to read file: %v\n", err)
		return "", [32]byte{}
	}
	return string(data), sha256.Sum256(data)
}

func saveContent(path, content string) {
	if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
		err := os.Remove(path)
		if err != nil {
			fmt.Printf("failed to remove file: %v\n", err)
		}
	}
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		fmt.Printf("failed to write file: %v\n", err)
	}
}
