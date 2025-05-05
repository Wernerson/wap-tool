package main

import (
	"fmt"
	"log"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/signintech/gopdf"
)

func check(err error) {
	if err != nil {
		log.Println("ERROR", err)
		panic(err)
	}
}

// Convert milimeters to pixels
func mmToPx(mm float64) float64 {
	return 3.78 * mm
}

type WAPDrawer interface {
	Draw(wap *Wap, output_path string)
}

type PDFDrawer struct {
	pdf           *gopdf.GoPdf
	wap           *Wap
	pageSize      *gopdf.Rect
	breakOption   *gopdf.BreakOption
	p1            gopdf.Point
	wapBox        gopdf.Rect
	bigColumns    int
	hoursPerDay   int
	minuteHeight  float64
	colWidth      float64
	smallFontSize float64
	largeFontSize float64
	padding       float64
}

func NewPDFDrawer() *PDFDrawer {
	pdf := gopdf.GoPdf{}
	return &PDFDrawer{pdf: &pdf,
		pageSize: gopdf.PageSizeA4Landscape,
		breakOption: &gopdf.BreakOption{
			Mode:           gopdf.BreakModeIndicatorSensitive,
			BreakIndicator: ' ',
			Separator:      "-",
		},
		bigColumns:    8,
		smallFontSize: 4,
		largeFontSize: 8,
		padding:       2,
	}

}

func (d *PDFDrawer) setupDocument() (err error) {
	mm6ToPx := mmToPx(6)
	trimbox := gopdf.Box{Left: mm6ToPx, Top: mm6ToPx, Right: d.pageSize.W - mm6ToPx, Bottom: d.pageSize.H - mm6ToPx}
	d.pdf.Start(gopdf.Config{
		PageSize: *d.pageSize,
		TrimBox:  trimbox,
	})
	err = d.pdf.AddTTFFont("regular", "./ttf/OpenSans-Regular.ttf")
	if err != nil {
		return err
	}
	err = d.pdf.AddTTFFont("bold", "./ttf/OpenSans-Bold.ttf")
	if err != nil {
		return err
	}
	err = d.pdf.AddTTFFont("italic", "./ttf/OpenSans-Italic.ttf")
	if err != nil {
		return err
	}
	err = d.pdf.SetFont("regular", "", d.smallFontSize)
	if err != nil {
		return err
	}

	// Padding left, top, right, and bottom from the page to the WAP Grid
	PL := mmToPx(20)
	PT := mmToPx(15)
	PR := mmToPx(15)
	PB := mmToPx(40)
	// Top-Left starting point
	d.p1 = gopdf.Point{X: PL, Y: PL}
	d.wapBox = gopdf.Rect{W: d.pageSize.W - PL - PR, H: d.pageSize.H - PT - PB}
	duration := d.wap.dayEnd.Sub(d.wap.dayStart)
	d.hoursPerDay = int(duration.Hours())
	d.colWidth = d.wapBox.W / float64(d.bigColumns)
	d.minuteHeight = d.wapBox.H / duration.Minutes()
	return nil
}

func (d *PDFDrawer) drawHeaderAndFooter(
	topLeft, topMiddle, topRight string,
	botLeft, botMiddle, botRight string,
) {
	padding := mmToPx(4)
	d.pdf.AddHeader(func() {
		err := d.pdf.SetFontSize(d.largeFontSize)
		check(err)
		d.pdf.SetY(padding)
		d.pdf.SetX(padding)
		err = d.pdf.CellWithOption(nil, topLeft, gopdf.CellOption{Align: gopdf.Left})
		check(err)
		tmW, err := d.pdf.MeasureTextWidth(topMiddle)
		check(err)
		d.pdf.SetX(d.pageSize.W/2 - tmW/2)
		err = d.pdf.CellWithOption(nil, topMiddle, gopdf.CellOption{Align: gopdf.Center})
		check(err)
		trW, err := d.pdf.MeasureTextWidth(topRight)
		check(err)
		d.pdf.SetX(d.pageSize.W - trW - padding)
		err = d.pdf.CellWithOption(nil, topRight, gopdf.CellOption{Align: gopdf.Right})
		check(err)
	})
	d.pdf.AddFooter(func() {
		err := d.pdf.SetFontSize(d.smallFontSize)
		check(err)
		d.pdf.SetY(d.pageSize.H - padding - d.smallFontSize)
		d.pdf.SetX(padding)
		err = d.pdf.CellWithOption(nil, botLeft, gopdf.CellOption{Align: gopdf.Left})
		check(err)
		bmW, err := d.pdf.MeasureTextWidth(botMiddle)
		check(err)
		d.pdf.SetX(d.pageSize.W/2 - bmW/2)
		err = d.pdf.CellWithOption(nil, botMiddle, gopdf.CellOption{Align: gopdf.Center})
		check(err)
		brW, err := d.pdf.MeasureTextWidth(botRight)
		check(err)
		d.pdf.SetX(d.pageSize.W - brW - padding)
		err = d.pdf.CellWithOption(nil, botRight, gopdf.CellOption{Align: gopdf.Right})
		check(err)
	})
}

func (d *PDFDrawer) Draw(wap *Wap, outputPath string) {
	d.wap = wap
	err := d.setupDocument()
	check(err)
	producer := "WAP-tool " + VERSION
	d.pdf.SetInfo(gopdf.PdfInfo{
		Author:   wap.Author,
		Title:    wap.Title,
		Producer: producer,
	})
	d.drawHeaderAndFooter(
		wap.Unit,
		wap.Title,
		wap.Version,
		"",
		"made with "+producer,
		wap.Author)

	layout := d.Layout()
	for week := range wap.Weeks {
		d.setupPage()
		d.drawWeeklyRemarks(week)
		for day := range 7 {
			totalDayIdx := week*7 + day
			d.drawColumnHeader(week*7 + day)
			/// we model the current footnote convetion:
			// for each day the footnote counter resets
			// Monday it starts with 10, Tuesday 20, ...
			footnoteCounter := (day + 1) * 10
			remarks := []string{}
			// draw events for this day
			for _, elem := range layout {
				if elem.dayOffset == totalDayIdx {
					originalEvent := elem.Event
					// Handle footnotes
					linkTo := ""
					if originalEvent.Footnote {
						footNoteSource := *originalEvent
						footNoteSource.Title = strconv.Itoa(footnoteCounter)
						footNoteSource.Description = ""
						elem.Event = &footNoteSource
						ts := MilitaryTime(originalEvent.Start)
						linkTo = fmt.Sprintf("fn:%d", footnoteCounter)
						remarks = append(remarks, fmt.Sprintf("%d  %s %s, %s", footnoteCounter, ts, originalEvent.Title, originalEvent.Description))
						footnoteCounter += 1
					}
					d.drawEvent(elem, linkTo)
				}
			}
			remarks = append(remarks, d.wap.DailyRemarks[totalDayIdx]...)
			d.drawDailyRemarks(day, remarks)
		}
	}
	log.Println("INFO writing pdf to ", outputPath)
	if err := d.pdf.WritePdf(outputPath); err != nil {
		log.Fatal("ERROR failed to write pdf output: ", err.Error())
	}
}

func (d *PDFDrawer) drawDailyRemarks(dayIdx int, remarks []string) {
	rectStart := d.toGridSystem(d.wap.dayEnd, dayIdx)
	rectStart.X += d.padding
	rectStart.Y += d.padding
	remarksHeight := 300.0
	remarksRect := gopdf.Rect{W: d.colWidth - 2*d.padding, H: remarksHeight*d.minuteHeight - d.padding}
	d.pdf.SetStrokeColor(0x00, 0x00, 0x00)
	d.pdf.SetFillColor(0xff, 0xff, 0xff)
	// drawRect(d.pdf, rectStart, remarksRect)
	d.drawMultiLineText(remarks, rectStart, remarksRect)
}

var numBeginningRe = regexp.MustCompile("^[0-9]+")

func (d *PDFDrawer) drawMultiLineText(text []string, point gopdf.Point, rect gopdf.Rect) {
	d.pdf.SetXY(point.X, point.Y)
	err := d.pdf.SetFont("regular", "", d.smallFontSize)
	check(err)
	for _, txt := range text {
		// hack: Create anchors for foonote links.
		// we abuse that the footnote lines start the footnote counter
		if match := numBeginningRe.FindString(txt); match != "" {
			d.pdf.SetAnchor("fn:" + match)
		}
		_, h, _ := d.pdf.IsFitMultiCell(&rect, txt)
		err := d.pdf.MultiCellWithOption(&rect, txt, gopdf.CellOption{BreakOption: d.breakOption})
		check(err)
		point.Y += h
		rect.H -= h
		d.pdf.SetXY(point.X, point.Y)
	}
}

func (d *PDFDrawer) drawWeeklyRemarks(weekIdx int) {
	// Remarks
	detHeightMin := 90.0
	rectStart := d.toGridSystem(d.wap.dayStart, d.bigColumns-1)
	headerRect := gopdf.Rect{W: d.colWidth, H: detHeightMin * d.minuteHeight}
	d.pdf.SetStrokeColor(0x00, 0x00, 0x00)
	d.pdf.SetFillColor(0xff, 0xff, 0xff)
	minutes := d.wap.dayEnd.Sub(d.wap.dayStart).Minutes() + 200
	colRect := gopdf.Rect{W: d.colWidth, H: minutes * d.minuteHeight}
	drawRect(d.pdf, rectStart, colRect)

	// Header
	rectStart.Y -= detHeightMin * d.minuteHeight
	d.pdf.SetStrokeColor(0x00, 0x00, 0x00)
	d.pdf.SetFillColor(0xf0, 0xf0, 0xf0)
	drawRect(d.pdf, rectStart, headerRect)
	d.pdf.SetXY(rectStart.X, rectStart.Y)
	d.pdf.SetTextColor(0x00, 0x00, 0x00)
	err := d.pdf.SetFont("bold", "", d.smallFontSize)
	check(err)
	err = d.pdf.CellWithOption(&headerRect, "Bemerkungen",
		gopdf.CellOption{
			Align: gopdf.Center | gopdf.Middle,
		})
	check(err)
	rectStart = d.toGridSystem(d.wap.dayStart, d.bigColumns-1)
	rectStart.X += d.padding
	rectStart.Y += d.padding
	colRect.H -= 2 * d.padding
	colRect.W -= 2 * d.padding
	d.drawMultiLineText(d.wap.Remarks[weekIdx], rectStart, colRect)
}

func (d *PDFDrawer) setupPage() {
	opt := gopdf.PageOption{
		PageSize: d.pageSize,
	}
	d.pdf.AddPageWithOption(opt)
	// The Big Grid
	d.pdf.SetStrokeColor(0, 0, 0)
	d.pdf.SetLineWidth(1)
	drawGrid(d.pdf, d.p1, d.wapBox, d.hoursPerDay, d.bigColumns)
	// Marks at 30 minutes
	d.pdf.SetStrokeColor(0x80, 0x80, 0x80)
	d.pdf.SetLineWidth(.5)
	drawHorizontalLines(d.pdf, d.p1, d.wapBox, d.hoursPerDay*2)
	// Marks at 15 minutes
	d.pdf.SetStrokeColor(0x80, 0x80, 0x80)
	d.pdf.SetLineWidth(.2)
	drawHorizontalLines(d.pdf, d.p1, d.wapBox, d.hoursPerDay*4)

	// Add time scale (mark all hours)
	err := d.pdf.SetFontSize(8)
	check(err)
	d.pdf.SetFillColor(0x00, 0x00, 0x00)
	d.pdf.SetStrokeColor(0x00, 0x00, 0x00)
	for hour := d.wap.dayStart.Hour(); hour <= d.wap.dayEnd.Hour(); hour += 1 {
		p := Add(d.toGridSystem(DayTime(hour, 0), 0), gopdf.Point{X: -20, Y: -d.smallFontSize})
		d.pdf.SetXY(p.X, p.Y)
		// convert to military time format
		err := d.pdf.Cell(nil, fmt.Sprintf("%02d00", hour))
		check(err)
	}
}

func (d *PDFDrawer) toGridSystem(t time.Time, dayIndex int) gopdf.Point {
	deltaX := float64(dayIndex) * d.colWidth
	deltaY := t.Sub(d.wap.dayStart).Minutes() * d.minuteHeight
	return Add(d.p1, gopdf.Point{X: deltaX, Y: deltaY})
}

// Draw the column header
// For example | Montag, 21.04.2025 |
// For example | Det1 | Det2 | Det3 |
func (d *PDFDrawer) drawColumnHeader(totalDayOffset int) {
	columnLocation := d.assignColumnLocations(d.wap.columns[totalDayOffset], d.colWidth)
	dayInWeek := totalDayOffset % 7
	detHeightMin := 90.0
	dayHeightMin := 20.0
	// Box for the week
	d.pdf.SetStrokeColor(0x00, 0x00, 0x00)
	d.pdf.SetFillColor(0xf0, 0xf0, 0xf0)
	RectStart := Add(d.toGridSystem(d.wap.dayStart, dayInWeek),
		gopdf.Point{X: 0, Y: -(detHeightMin + dayHeightMin) * d.minuteHeight})
	rect := gopdf.Rect{W: d.colWidth, H: dayHeightMin * d.minuteHeight}
	drawRect(d.pdf, RectStart, rect)
	d.pdf.SetXY(RectStart.X, RectStart.Y)
	err := d.pdf.SetFont("bold", "", d.smallFontSize)
	check(err)
	d.pdf.SetTextColor(0x00, 0x00, 0x00)
	dayName := d.wap.dayNames[totalDayOffset]
	err = d.pdf.CellWithOption(&rect, dayName,
		gopdf.CellOption{
			Align: gopdf.Center | gopdf.Middle,
		})
	check(err)
	d.pdf.SetStrokeColor(0x00, 0x00, 0x00)
	d.pdf.SetFillColor(0xf0, 0xf0, 0xf0)
	// empty box if no columns are defined
	if len(columnLocation) == 0 {
		RectStart := Add(d.toGridSystem(d.wap.dayStart, dayInWeek),
			gopdf.Point{X: 0, Y: -detHeightMin * d.minuteHeight})
		rect := gopdf.Rect{W: d.colWidth, H: detHeightMin * d.minuteHeight}
		drawRect(d.pdf, RectStart, rect)
	}
	for colName, opts := range columnLocation {
		RectStart := Add(d.toGridSystem(d.wap.dayStart, dayInWeek),
			gopdf.Point{X: opts.Offset, Y: -detHeightMin * d.minuteHeight})
		rect := gopdf.Rect{W: opts.W, H: detHeightMin * d.minuteHeight}
		drawRect(d.pdf, RectStart, rect)
		d.pdf.SetXY(RectStart.X, RectStart.Y)
		d.pdf.SetTextColor(0x00, 0x00, 0x00)
		d.pdf.Rotate(90.0, RectStart.X+rect.W/2, RectStart.Y+rect.H/2)
		err := d.pdf.SetFont("bold", "", d.smallFontSize)
		check(err)
		err = d.pdf.CellWithOption(&rect, colName,
			gopdf.CellOption{
				Align: gopdf.Center | gopdf.Middle,
			})
		check(err)
		d.pdf.RotateReset()
	}
}

func (d *PDFDrawer) drawEvent(elem EventPosition, linkTo string) {
	event := elem.Event
	if c, ok := d.wap.categories[event.Category]; ok {
		d.pdf.SetFillColor(c.R, c.G, c.B)
	}
	drawRect(d.pdf, elem.P, elem.R)
	title := event.Title
	titleFontSize := d.smallFontSize
	rect := gopdf.Rect{
		W: elem.R.W,
		H: elem.R.H,
	}
	// No padding for small events (like Tagwache/AV)
	if rect.H >= 16 {
		elem.P.Y += d.padding
		rect.H -= d.padding
	}
	d.pdf.SetXY(elem.P.X, elem.P.Y)
	d.pdf.SetTextColor(0x00, 0x00, 0x00)
	err := d.pdf.SetFont("bold", "", titleFontSize)
	check(err)
	ok, heightNeeded, _ := d.pdf.IsFitMultiCell(&rect, title)
	if !ok {
		log.Println("WARNING", "title does not fit in rectangle:", event.Title)
	}
	err = d.pdf.MultiCellWithOption(&rect, title,
		gopdf.CellOption{
			Align:       gopdf.Center,
			BreakOption: d.breakOption,
		})
	check(err)
	d.pdf.SetXY(elem.P.X, elem.P.Y+heightNeeded)
	err = d.pdf.SetFont("regular", "", Min(d.smallFontSize, titleFontSize))
	check(err)
	// TODO check properly whether description fits
	// this does not account for breakOption
	// we get does not fit event though there would be enough space on the next line
	// ok, _, _ = d.pdf.IsFitMultiCellWithNewline(&rect, event.Description)
	if event.Description != "" {
		descriptionRect := gopdf.Rect{
			W: elem.R.W,
			H: elem.R.W - heightNeeded,
		}
		err := d.pdf.MultiCellWithOption(&descriptionRect, event.Description,
			gopdf.CellOption{
				Align:       gopdf.Center,
				BreakOption: d.breakOption,
			})
		check(err)
	}
	if elem.Event.Footnote {
		d.pdf.AddInternalLink(linkTo, elem.P.X, elem.P.Y, elem.R.H, elem.R.W)
	}
}

// TODO(refactor) reuse EventPosition?
type columnInfo struct {
	// Offset from the x of the day
	Offset float64
	// Width of the column
	W float64
}

func (d *PDFDrawer) assignColumnLocations(columns []string, width float64) map[string]columnInfo {
	m := make(map[string]columnInfo)
	// divide evently
	if len(columns) == 0 {
		return m
	}
	// Treat Beso specially: otherwise it looks bad
	// Assumption

	smallWidth := width / 6
	normalCols := 0
	for _, c := range columns {
		if strings.EqualFold(c, "Beso") {
			width -= smallWidth
			m["Beso"] = columnInfo{Offset: width, W: smallWidth}
		} else {
			normalCols += 1
		}
	}
	columnWidth := width / float64(normalCols)
	accumulator := 0.0
	for _, c := range columns {
		if strings.EqualFold(c, "Beso") {
			continue
		}
		m[c] = columnInfo{Offset: accumulator, W: columnWidth}
		accumulator += columnWidth
	}
	return m
}

func Add(p1, p2 gopdf.Point) gopdf.Point {
	return gopdf.Point{X: p1.X + p2.X, Y: p1.Y + p2.Y}
}

// P=(X, y) ------- W	width
// |
// | H	and height
type EventPosition struct {
	// top-left corner
	P gopdf.Point
	// the rectangle
	R gopdf.Rect
	// reference to the original Event
	Event        *Event
	dayOffset    int // hack: need this for repeating events
	parallelCols int // 0 <= parallelCols < len(day.Columns) - 1
	parallelIdx  int // 0 <= parallelCols < len(day.Columns) - 1
}

// TODO can we make this independent of the specific drawer?

// Computes a layout for each event
func (d *PDFDrawer) Layout() (res []EventPosition) {
	// Initialize
	for _, event := range d.wap.events {
		// sanitize the event before printing (only local here)
		if event.Start.Before(d.wap.dayStart) {
			event.Start = d.wap.dayStart
		}
		if d.wap.dayEnd.Before(event.End) {
			event.End = d.wap.dayEnd
		}
		minutes := event.End.Sub(event.Start).Minutes()
		if minutes <= 0 {
			continue
		}
		height := minutes * d.minuteHeight
		RectStart := d.toGridSystem(event.Start, event.DayOffset%7)

		res = append(res, EventPosition{
			dayOffset: event.DayOffset,
			P:         RectStart,
			R:         gopdf.Rect{W: d.colWidth, H: height},
			Event:     &event})
	}
	// First pass
	for i, ev := range res {
		if ev.Event.Repeats {
			continue
		}
		// - Check for overlap
		// Assumption: this can only happen for events in a single columns
		if len(ev.Event.AppearsIn) != 1 {
			continue
		}
		// For example events ev1 and ev2 that overlap in time will have
		//	-----
		// | ev1 |-----|
		// |	 | ev2 |
		// -------------
		// ev1.parallelCols = ev2.parallelCols = 1	(the number of other events in this column)
		// ev1.parallelIdx = 1 and ev2.parallelIdx = 2
		overlapping := ev.parallelIdx
		for j := i + 1; j < len(res); j += 1 {
			next := res[j]
			if next.Event.DayOffset != ev.Event.DayOffset {
				break
			}
			if ev.Event.End.Compare(next.Event.Start) <= 0 {
				break
			}
			if next.Event.Repeats {
				continue
			}
			if o := overlap(next.Event.AppearsIn, ev.Event.AppearsIn); len(o) > 0 {
				log.Printf("WARNING overlapping events in columns %v %v %v\n", o, ev, next)
				res[j].parallelIdx++
				overlapping++
			}
		}
		res[i].parallelCols = overlapping
	}
	// Second pass
	newPositions := []EventPosition{}
	for i, elem := range res {
		event := elem.Event
		columnLocation := d.assignColumnLocations(d.wap.columns[event.DayOffset], d.colWidth)
		if event.Repeats {
			res[i].R.W = d.colWidth
			for day := range d.wap.Days {
				if event.DayOffset <= day {
					elem.dayOffset = day
					elem.P = d.toGridSystem(event.Start, day%7)
					newPositions = append(newPositions, elem)
				}
			}
			continue
		}
		if elem.parallelCols > 0 {
			// Example:
			// |   Det         |
			// | ev1 | ev2 |ev3|
			width := d.colWidth
			for _, c := range event.AppearsIn {
				width = columnLocation[c].W
			}
			eventWidth := width / float64(elem.parallelCols+1)
			offset := float64(elem.parallelIdx) * eventWidth

			elem.P.X += offset
			elem.R.W = eventWidth

			for _, c := range d.wap.columns[event.DayOffset] {
				if !slices.Contains(event.AppearsIn, c) {
					elem.P.X += columnLocation[c].W
				} else {
					break
				}
			}
			newPositions = append(newPositions, elem)
			continue
		}
		eventWidth := 0.0
		offset := 0.0
		// if an event appears in multiple consecutive columns they can be merged
		// for ev1 that appears in columns A and B:
		// | A | B |
		// |  ev1  |
		// if the columns are not adjacent, the event is printed in multiple ones
		// for ev1 that appears in columns A and C:
		// | A   | B   | C   |
		// | ev1 | ... | ev2 |
		active := false // true if we are expanding a column
		// TODO ugly cloning: the original event also still exists!
		for _, c := range d.wap.columns[event.DayOffset] {
			if slices.Contains(event.AppearsIn, c) {
				eventWidth += columnLocation[c].W
				if !active {
					active = true
					offset = columnLocation[c].Offset
				}
			} else if active {
				active = false
				clone := res[i]
				clone.P.X += offset
				clone.R.W = eventWidth
				newPositions = append(newPositions, clone)
				eventWidth = 0.0
			}
		}
		if active {
			clone := res[i]
			clone.P.X += offset
			clone.R.W = eventWidth
			newPositions = append(newPositions, clone)
		}
	}
	return newPositions
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
