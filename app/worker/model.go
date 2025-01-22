package worker

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

const calendarID = "f4bd341e-d8ea-11ef-b96e-3b8bd9e2b02f"

type EventLocation struct {
	ID          int64
	Name        string
	Description string
}

type Entry struct {
	ID       int64
	Guid     string
	Title    string
	Content  string
	Excerpt  string
	URL      string
	Created  time.Time
	Modified time.Time
	FromDate time.Time
	ToDate   time.Time
	AllDay   bool
	Location EventLocation
}

type Entries []Entry

func (e Entries) ToIcs() (string, [32]byte) {
	var b strings.Builder
	b.WriteString("BEGIN:VCALENDAR\nVERSION:2.0\nPRODID:-//hegering-gronau-epe//termine//DE\n")
	b.WriteString("CALSCALE:GREGORIAN\nNAME:Hegering Gronau-Epe Termine\n")
	b.WriteString("X-WR-CALNAME:Hegering Gronau-Epe Termine\n")
	b.WriteString("X-WR-RELCALID:" + calendarID + "\n")
	b.WriteString("X-WR-PUBLISHED-TTL:PT1H\n")
	b.WriteString("X-WR-TIMEZONE:Europe/Berlin\n")
	b.WriteString("X-WR-CALDESC:Termine des Hegering Gronau-Epe\n")

	for _, entry := range e {
		fmt.Fprintf(&b, "BEGIN:VEVENT\nUID:%s\n", entry.Guid)
		fmt.Fprintf(&b, "METHOD:PUBLISH\nSTATUS:CONFIRMED\n")
		fmt.Fprintf(&b, "COLOR:GREEN\n")
		fmt.Fprintf(&b, "DTSTAMP:%s\n", entry.Created.Format("20060102T150405Z"))
		if entry.AllDay {
			fmt.Fprintf(&b, "DTSTART;TZID=Europe/Berlin:%s\n", entry.FromDate.Format("20060102"))
			if !entry.ToDate.Before(entry.FromDate) {
				fmt.Fprintf(&b, "DTEND;TZID=Europe/Berlin:%s\n", entry.ToDate.Format("20060102"))
			}
		} else {
			fmt.Fprintf(&b, "DTSTART;VALUE=DATE:%s\n", entry.FromDate.Format("20060102T150405"))
			if entry.ToDate.After(entry.FromDate) {
				fmt.Fprintf(&b, "DTEND;VALUE=DATE:%s\n", entry.ToDate.Format("20060102T150405"))
			}
		}
		if entry.Location.ID > 0 {
			fmt.Fprintf(&b, "LOCATION:%s, %s\n", entry.Location.Name, entry.Location.Description)
		}

		fmt.Fprintf(&b, "SUMMARY:%s\nDESCRIPTION:%s\nURL:%s\n", entry.Title, entry.Excerpt, entry.URL)
		fmt.Fprintf(&b, "LAST-MODIFIED:%s\n", entry.Modified.Format("20060102T150405Z"))
		fmt.Fprintf(&b, "X-MICROSOFT-CDO-ALLDAYEVENT:%v\n", entry.AllDay)
		fmt.Fprintf(&b, "X-MICROSOFT-CDO-INTENDEDSTATUS:FREE\n")
		fmt.Fprintf(&b, "REFRESH-INTERVAL;VALUE=DURATION:P4D\n")
		fmt.Fprintf(&b, "ORGANIZER;CN=\"Hegering Gronau-Epe\":mailto:HegeringGronauEpe@gmail.com\n")

		fmt.Fprintf(&b, "END:VEVENT\n")
	}
	b.WriteString("END:VCALENDAR")
	s := b.String()
	return s, sha256.Sum256([]byte(s))
}
