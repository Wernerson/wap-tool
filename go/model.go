package main

import (
	"fmt"
	"log"
	"slices"
	"sort"
	"time"
)

const (
	daysInWeek              = 7
	minimumEventDurationMin = 10
)

var defaultColor = RGBColor{0xf0, 0xf0, 0xf0}

// Main type representing a WAP
// Days < Weeks * 7
// len(Remarks) == Weeks
// len(DailyRemarks) == Weeks * 7
// len(columns) == Weeks * 7
// dayStart and dayEnd have minutes (00, 15, 30, 45)
type Wap struct {
	// total number of days
	Days   int
	Weeks  int
	events Events
	// Styling information
	categories map[string]RGBColor
	// Columns for each day
	columns  [][]string
	firstDay time.Time
	dayNames []string
	dayStart time.Time
	dayEnd   time.Time
	// Metadata
	Unit, Version, Author, Title string
	// Remarks for each week
	Remarks [][]string
	// Remarks for each day
	DailyRemarks [][]string
}

// Represents a valid Event
// This is distinct from WapJsonDaysElemEventsElem that may misses optional fields.
// The following invariants hold.
// w is the parent Wap this event is part of.
// - start.Before(end)
// - 0 <= dayOffset && dayOffset < w.Days
// - w.categories[Category]
type Event struct {
	Start, End  time.Time
	OpenEnd     bool
	DayOffset   int
	Repeats     bool
	AppearsIn   []string
	Category    string
	Title       string
	Description string
	Footnote    bool
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
	return fmt.Sprintf("Event(day #%d %v-%v %v)", e.DayOffset+1, t1, t2, e.Title)
}

func NewWAP(data *WapJson) (w *Wap) {
	w = new(Wap)
	w.categories = make(map[string]RGBColor)
	// Default color
	w.categories[""] = defaultColor
	w.parseColors(data.Categories)
	w.dayStart = DayTime(5, 30)
	w.dayEnd = DayTime(23, 30)
	firstDayD, err := time.Parse(time.DateOnly, data.Meta.FirstDay)
	if err != nil {
		log.Fatal("ERROR failed to parse date. Use the format YYYY-MM-DD: ", err)
	}
	w.firstDay = firstDayD
	if data.Meta.StartTime != nil {
		t1, err := parseDayTime(*data.Meta.StartTime)
		if err != nil {
			log.Println(err)
			log.Printf("WARNING not defined when the day starts. Falling back to default: %s\n", MilitaryTime(w.dayStart))
		} else {
			w.dayStart = RoundToQuarterHour(t1)
		}
	}
	if data.Meta.EndTime != nil {
		t2, err := parseDayTime(*data.Meta.EndTime)
		if err != nil {
			log.Println(err)
			log.Printf("WARNING not defined when the day ends. Falling back to default: %s\n", MilitaryTime(w.dayEnd))
		} else {
			w.dayEnd = RoundToQuarterHour(t2)
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
	w.Weeks = len(data.Weeks)
	for weekIdx, week := range data.Weeks {
		w.Days += len(week.Days)
		w.Remarks = append(w.Remarks, week.Remarks)
		for i := range 7 {
			correctedTime := w.firstDay.AddDate(0, 0, weekIdx*daysInWeek+i)
			localDay := TranslateWeekDay(correctedTime.Weekday())
			name := localDay + ", " + SwissDate(correctedTime)
			w.dayNames = append(w.dayNames, name)
			w.columns = append(w.columns, []string{})
			if len(w.columns) == 0 {
				log.Printf("WARNING no columns defined on %s\n", name)
			}
			w.DailyRemarks = append(w.DailyRemarks, []string{})
		}
		for dayIdx, day := range week.Days {
			w.columns[weekIdx*daysInWeek+dayIdx] = day.Columns
			w.DailyRemarks[weekIdx*daysInWeek+dayIdx] = day.Remarks
		}
	}
	w.parseEvents(data.Weeks)
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
			log.Println("WARNING falling back to default color. Failed to parse: ", err.Error())
			c = defaultColor
		}
		w.categories[cat.Identifier] = c
	}
}

func (w *Wap) parseEvents(weeks []WapJsonWeeksElem) {
	for weekIdx, week := range weeks {
		if len(week.Days) > daysInWeek {
			log.Printf("ERROR: too many days %d, expected at most %d", len(week.Days), daysInWeek)
			continue
		}
		for i, day := range week.Days {
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
				description := ""
				if event.Description != nil {
					description = *event.Description
				}
				freshEvent := Event{
					Start:       start,
					End:         end,
					Title:       event.Title,
					Description: description,
					DayOffset:   weekIdx*daysInWeek + i,
					AppearsIn:   []string{},
				}
				if end.Before(start) {
					log.Printf("WARNING end before start time for %v. Swapping it.\n", freshEvent)
					freshEvent.Start, freshEvent.End = freshEvent.End, freshEvent.Start
				}
				if len(event.AppearsIn) == 0 {
					log.Printf("INFO appearsIn is empty. The %v implicitly appears in all columns for this day.\n", freshEvent)
				}
				if event.Category != nil {
					freshEvent.Category = *event.Category
				}
				if event.Repeats != nil {
					freshEvent.Repeats = true
				}
				if event.Footnote != nil {
					freshEvent.Footnote = *event.Footnote
				}
				if event.OpenEnd != nil {
					freshEvent.OpenEnd = *event.OpenEnd
				}
				for _, col := range event.AppearsIn {
					if slices.Contains(day.Columns, col) {
						freshEvent.AppearsIn = append(freshEvent.AppearsIn, col)
					} else {
						if !freshEvent.Repeats {
							log.Printf("WARNING ignoring column '%s' in event '%s' that is not defined for week %d day %d: %v \n", col, event.Title, weekIdx+1, i+1, w.columns[i])
						}
						freshEvent.AppearsIn = append(freshEvent.AppearsIn, col)
					}
				}
				w.events = append(w.events, freshEvent)
			}
		}
	}
	sort.Sort(w.events)

	// Validate
	for _, event := range w.events {
		duration := event.End.Sub(event.Start)
		if duration.Minutes() < float64(minimumEventDurationMin) {
			log.Printf("WARNING %v length %v min too short and will not be properly displayed. Recommended duration %d min\n", event, duration.Minutes(), minimumEventDurationMin)
		}

		if event.Start.Before(w.dayStart) {
			log.Printf("WARNING start time before the day start %v for event %v", MilitaryTime(w.dayStart), event)
		}

		if w.dayEnd.Before(event.End) {
			log.Printf("WARNING end time before the day end %v for event %v", MilitaryTime(w.dayEnd), event)
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
