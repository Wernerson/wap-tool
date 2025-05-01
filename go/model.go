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
	columns   [][]string
	firstDay  time.Time
	dayNames  []string
	dayStart  time.Time
	dayEnd    time.Time
}

type Event struct {
	json         *WapJsonDaysElemEventsElem
	start, end   time.Time
	dayOffset    int
	repeats      bool
	parallelCols int
	parallelIdx  int
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
	w.parseColors()
	w.dayStart = DayTime(23, 30)
	w.Days = len(data.Days)
	w.columns = make([][]string, w.Days)
	w.dayNames = make([]string, w.Days)

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
	w.processEvents()
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
	for i, day := range w.data.Days {
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
			if event.Repeats != nil {
				freshEvent.repeats = true
				w.repeating = append(w.repeating, freshEvent)
			}
			w.events = append(w.events, freshEvent)
		}
	}
	sort.Sort(w.events)

	// Validate
	for i, event := range w.events {
		// - Check it has a valid duration
		duration := event.end.Sub(event.start)
		if duration < 0 {
			log.Printf("WARNING event ends before it starts: %v\n", event)
		}
		minimumDurationMin := 10
		if duration.Minutes() < float64(minimumDurationMin) {
			log.Printf("WARNING event length %v min too short and will not be properly displayed. The duration should be at least %d min\n", duration.Minutes(), minimumDurationMin)
		}

		if event.start.Before(w.dayStart) {
			log.Printf("WARNING start time before the day start %v for event %v", w.dayStart, event)
		}

		if w.dayEnd.Before(event.end) {
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
		if cat := event.json.Category; cat != nil {
			if _, ok := w.colors[*cat]; !ok {
				// MAYBE: add helpful message how to fix it
				log.Printf("WARNING category %v is not defined\n", *cat)
				event.json.Category = nil
			}
		}
		// - Check for overlap
		// For example events ev1 and ev2 that overlap in time will have
		//	-----
		// | ev1 |-----|
		// |	 | ev2 |
		// -------------
		// ev1.parallelCols = ev2.parallelCols = 1	(the number of other events in this column)
		// ev1.parallelIdx = 1 and ev2.parallelIdx = 2
		overlapping := event.parallelIdx
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
				w.events[j].parallelIdx++
				overlapping++
			}
		}
		w.events[i].parallelCols = overlapping
		// - TODO validate that appearsIn references a column defined for that day.

		// - TODO note if there is unallocated time
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
