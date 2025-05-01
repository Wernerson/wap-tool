package main

import (
	"fmt"
	"log"
	"slices"
	"sort"
	"time"
)

var DefaultColor = RGBColor{0xf0, 0xf0, 0xf0}
var MinimumEventDurationMin = 10

// Main type representing a WAP
type Wap struct {
	Days   int
	events Events
	// Styling information
	categories map[string]RGBColor
	columns    [][]string
	firstDay   time.Time
	dayNames   []string
	dayStart   time.Time
	dayEnd     time.Time
	// Metadata
	Unit, Version, Author, Title string
	Remarks                      []string
}

// Represents a valid Event
// This is distinct from WapJsonDaysElemEventsElem that may misses optional fields.
// The following invariants hold.
// w is the parent Wap this event is part of.
// - start.Before(end)
// - 0 <= dayOffset && dayOffset < w.Days
// - set(appearsIn) subset w.Days[dayOffset]
// - w.categories[Category]
type Event struct {
	Start, End  time.Time
	DayOffset   int
	Repeats     bool
	AppearsIn   []string
	Category    string
	Title       string
	Description string
}

type Events []Event

// Implement sort.Interface for Events
func (e Events) Len() int      { return len(e) }
func (e Events) Swap(i, j int) { e[i], e[j] = e[j], e[i] }

// Lexicographic order by (DayOffset, Start, End)
func (e Events) Less(i, j int) bool {
	if e[i].DayOffset < e[j].DayOffset {
		return true
	} else if e[i].DayOffset > e[j].DayOffset {
		return false
	}
	if e[i].Start.Compare(e[j].Start) == -1 {
		return true
	}
	if e[i].Start.Compare(e[j].Start) == 0 {
		return e[i].End.Compare(e[j].End) < 0
	}
	return false
}

func (e Event) String() string {
	t1 := e.Start.Format("15:04")
	t2 := e.End.Format("15:04")
	return fmt.Sprintf("Event(#%d %v-%v %v)", e.DayOffset, t1, t2, e.Title)
}

func NewWAP(data *WapJson) (w *Wap) {
	w = new(Wap)
	w.categories = make(map[string]RGBColor)
	// Default color
	w.categories[""] = DefaultColor
	w.events = []Event{}
	w.parseColors(data.Categories)
	w.dayStart = DayTime(5, 30)
	w.dayEnd = DayTime(23, 30)
	w.Days = len(data.Days)
	w.columns = make([][]string, w.Days)
	w.dayNames = make([]string, w.Days)
	w.Remarks = slices.Clone(data.Remarks)
	firstDayD, err := time.Parse(time.DateOnly, data.Meta.FirstDay)
	if err != nil {
		log.Fatal("ERROR failed to parse date. Use the format YYYY-MM-DD: ", err)
	}
	w.firstDay = firstDayD
	if data.Meta.StartTime != nil {
		t1, err := parseDayTime(*data.Meta.StartTime)
		if err != nil {
			log.Println(err)
			log.Println("WARNING not defined when the day start. Falling back to default.")
		} else {
			w.dayStart = t1
		}
	}
	if data.Meta.EndTime != nil {
		t2, err := parseDayTime(*data.Meta.EndTime)
		if err != nil {
			log.Println(err)
			log.Println("WARNING not defined when the day ends. Falling back to default.")
		} else {
			w.dayEnd = t2
		}
	}

	if data.Meta.Unit != nil {
		w.Unit = *data.Meta.Unit
	}
	if data.Meta.Version != nil {
		w.Version = *data.Meta.Version
	} else {
		w.Version = time.Now().Format(time.DateOnly)
	}
	w.Author = data.Meta.Author
	w.Title = data.Meta.Title

	w.parseEvents(data.Days)
	return
}

func (w *Wap) String() string {
	return fmt.Sprintf("colors: %v\nevents: %v\ncolumns: %v",
		w.categories, w.events, w.columns)
}

func (w *Wap) parseColors(categories []WapJsonCategoriesElem) {
	for _, cat := range categories {
		c, err := parseColor(cat.Color)
		if err != nil {
			log.Println("WARNING falling back to default colors: ", err.Error())
			c = DefaultColor
		}
		w.categories[cat.Identifier] = c
	}
}

func (w *Wap) parseEvents(days []WapJsonDaysElem) {
	for i, day := range days {
		if day.Name != nil {
			w.dayNames[i] = *day.Name
		} else {
			correctedTime := w.firstDay.AddDate(0, 0, i)
			w.dayNames[i] = correctedTime.Weekday().String() + ", " + correctedTime.Format(time.DateOnly)
		}
		w.columns[i] = day.Columns
		for _, event := range day.Events {
			start, err := parseDayTime(event.Start)
			if err != nil {
				log.Println("ERROR: failed to parse start time: ", err.Error())
				continue
			}
			end, err := parseDayTime(event.End)
			if err != nil {
				log.Println("ERROR: failed to parse end time: ", err.Error())
				continue
			}
			if end.Before(start) {
				log.Println("WARNING end before start time. Swapping it.")
				start, end = end, start
			}
			description := ""
			if event.Location != nil {
				description += *event.Location
			}
			if event.Responsible != nil {
				if description != "" {
					description += ", "
				}
				description += *event.Responsible
			}
			freshEvent := Event{
				Start:       start,
				End:         end,
				Title:       event.Title,
				Description: description,
				DayOffset:   i,
				AppearsIn:   []string{},
			}
			for _, col := range event.AppearsIn {
				if slices.Contains(day.Columns, col) {
					freshEvent.AppearsIn = append(freshEvent.AppearsIn, col)
				} else {
					log.Printf("WARNING ignoring column %v that is not defined for day %d\n", col, i)
				}
			}
			if len(event.AppearsIn) == 0 {
				log.Println("WARNING appearsIn is empty. The event implicitly appears in all columns for this day.", event)
				freshEvent.AppearsIn = slices.Clone(day.Columns)
			}
			if event.Category != nil {
				freshEvent.Category = *event.Category
			}
			if event.Repeats != nil {
				freshEvent.Repeats = true
			}
			w.events = append(w.events, freshEvent)
		}
	}
	sort.Sort(w.events)

	// Validate
	for _, event := range w.events {
		// - Check it has a valid duration
		duration := event.End.Sub(event.Start)
		if duration < 0 {
			log.Printf("WARNING event ends before it starts: %v\n", event)
		}
		if duration.Minutes() < float64(MinimumEventDurationMin) {
			log.Printf("WARNING event length %v min too short and will not be properly displayed. The duration should be at least %d min\n", duration.Minutes(), MinimumEventDurationMin)
		}

		if event.Start.Before(w.dayStart) {
			log.Printf("WARNING start time before the day start %v for event %v", w.dayStart, event)
		}

		if w.dayEnd.Before(event.End) {
			log.Printf("WARNING end time before the day end %v for event %v", w.dayEnd, event)
		}

		// MAYBE: Give it an end if it has none
		// if NO_END {
		// 	if i+1 < len(w.events) {
		// 		nextEvent := w.events[i+1]
		// 		if event.dayOffset == nextEvent.dayOffset {
		// 			event.end = nextEvent.start
		// 		} else {
		// 			event.end = w.dayEnd
		// 		}
		// 	}
		// }
		// - Check for valid category
		if _, ok := w.categories[event.Category]; !ok {
			// MAYBE: add helpful message how to fix it
			log.Printf("WARNING category %v is not defined\n", event.Category)
			event.Category = ""
		}
		// - MAYBE note if there is unallocated time
	}
}
