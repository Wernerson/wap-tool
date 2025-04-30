package main

import (
	"fmt"
	"log"
	"time"
)

type Wap struct {
	data      *WapJson
	colors    map[string]RGBColor
	repeating []Event
	events    []Event
	columns   map[int]map[string]struct{}
	dayStart  time.Time
	dayEnd    time.Time
}

type Event struct {
	json       *WapJsonDaysElemEventsElem
	start, end time.Time
	dayOffset  int
}

func (e Event) String() string {
	t1 := e.start.Format("15:04")
	t2 := e.end.Format("15:04")
	return fmt.Sprintf("#%d %v-%v %v\n", e.dayOffset, t1, t2, e.json.Title)
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
			addCols(i, event.Columns)
			if event.Repeats != nil {
				w.repeating = append(w.repeating, freshEvent)
			}
			w.events = append(w.events, freshEvent)
		}
	}
}
