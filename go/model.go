package main

import (
	"fmt"
	"log"
	"sort"
	"time"
)

type Wap struct {
	Days      int
	data      *WapJson
	colors    map[string]RGBColor
	repeating Events
	events    Events
	columns   map[int]map[string]struct{}
	dayStart  time.Time
	dayEnd    time.Time
}

type Event struct {
	json       *WapJsonDaysElemEventsElem
	start, end time.Time
	dayOffset  int
}

type Events []Event

// Implement sort.Interface for Events
func (e Events) Len() int      { return len(e) }
func (e Events) Swap(i, j int) { e[i], e[j] = e[j], e[i] }

// Lexicographic order by day, start time and end time
func (e Events) Less(i, j int) bool {
	if e[i].dayOffset < e[j].dayOffset {
		return true
	} else if e[i].dayOffset > e[j].dayOffset {
		return false
	}
	if e[i].start.Compare(e[j].start) == -1 {
		return true
	}
	if e[i].start.Compare(e[j].start) == 0 {
		return e[i].end.Compare(e[j].end) < 0
	}
	return false
}

func (e Event) String() string {
	t1 := e.start.Format("15:04")
	t2 := e.end.Format("15:04")
	return fmt.Sprintf("Event(#%d %v-%v %v)", e.dayOffset, t1, t2, e.json.Title)
}

func NewWAP(data *WapJson) (w *Wap) {
	w = new(Wap)
	w.data = data
	w.colors = make(map[string]RGBColor)
	w.events = []Event{}
	w.repeating = []Event{}
	w.columns = make(map[int]map[string]struct{})
	w.parseColors()
	w.processEvents()
	w.dayStart = DayTime(23, 30)
	w.Days = 7 // TODO
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
	return
}

func (w *Wap) String() string {
	return fmt.Sprintf("raw: %v\ncolors: %v\nevents: %v\ncolumns: %v",
		w.data, w.colors, w.events, w.columns)
}

func (w *Wap) parseColors() {
	for _, cat := range w.data.Categories {
		c, err := parseColor(*cat.Color)
		if err != nil {
			log.Println(err.Error())
			log.Println("WARNING falling back to default colors")
			// MAYBE: pick from a set of predefined columns
			c = RGBColor{127, 127, 127}
		}
		w.colors[cat.Identifier] = c
	}
}

func (w *Wap) processEvents() {

	addCols := func(idx int, cols []string) {
		for _, c := range cols {
			if m, ok := w.columns[idx]; ok {
				m[c] = struct{}{}
			} else {
				w.columns[idx] = map[string]struct{}{c: struct{}{}}
			}
		}
	}

	for i, day := range w.data.Days {
		// TODO day.Offset
		for _, event := range day.Events {
			start, err := parseDayTime(event.Start)
			if err != nil {
				log.Println(err.Error())
				log.Println("WARNING no start time defined. Ignoring event.")
				continue
			}
			end, err := parseDayTime(event.End)
			if err != nil {
				log.Println(err.Error())
				log.Println("WARNING no end time defined. Trying to implicitly find it.")
			}
			freshEvent := Event{
				json:      &event,
				start:     start,
				end:       end,
				dayOffset: i,
			}
			addCols(i, event.AppearsIn)
			if event.Repeats != nil {
				w.repeating = append(w.repeating, freshEvent)
			}
			w.events = append(w.events, freshEvent)
		}
	}
	sort.Sort(w.events)

	// Validate
	for i, event := range w.events {
		// - Check it has a valid duration
		if !event.end.After(event.start) {
			log.Printf("WARNING event ends before it starts: %v %v\n", event.start, event.end)
			event.start, event.end = event.end, event.start
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
		if cat := event.json.Category; cat != nil {
			if _, ok := w.colors[*cat]; !ok {
				// MAYBE: add helpful message how to fix it
				log.Printf("WARNING category %v is not defined\n", *cat)
				event.json.Category = nil
			}
		}
		// - Check for overlap
		for j := i + 1; j < len(w.events); j += 1 {
			nextEvent := w.events[j]
			if nextEvent.dayOffset != event.dayOffset {
				break
			}
			if event.end.Compare(nextEvent.start) <= 0 {
				break
			}
			if o := overlap(nextEvent.json.AppearsIn, event.json.AppearsIn); len(o) > 0 {
				log.Printf("WARNING overlapping events in columns %v %v %v\n", o, event, nextEvent)
			}
		}
	}
}

func overlap(xs []string, ys []string) (res []string) {
	for _, x := range xs {
		for _, y := range ys {
			if x == y {
				res = append(res, x)
			}
		}
	}
	return
}
