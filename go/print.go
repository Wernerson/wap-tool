package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/signintech/gopdf"
)

func checkPDFerror(err error) {
	if err != nil {
		log.Println("ERROR while printing", err)
	}
}

// Convert milimeters to pixels
func mmToPx(mm float64) float64 {
	return 3.78 * mm
}

type WAPDrawer interface {
	Draw(wap *Wap, outputPath string)
}

type PDFDrawer struct {
	pdf         *gopdf.GoPdf
	wap         *Wap
	pageSize    *gopdf.Rect
	breakOption *gopdf.BreakOption
	// Points are offset from the top-left of the page
	// p1 ------+
	//  | wapBox|
	//  + ------+
	p1            gopdf.Point
	wapBox        gopdf.Rect
	bigColumns    int
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
		bigColumns:    8, // weekly remarks is a column too
		smallFontSize: 4,
		largeFontSize: 8,
		padding:       2,
	}

}

// Embed the font files in the executable
//
//go:embed ttf/OpenSans-Regular.ttf
var openSansRegular []byte

//go:embed ttf/OpenSans-Bold.ttf
var openSansBold []byte

//go:embed ttf/OpenSans-Italic.ttf
var openSansItalic []byte

func (d *PDFDrawer) setupDocument() (err error) {
	mm6ToPx := mmToPx(6)
	trimbox := gopdf.Box{Left: mm6ToPx, Top: mm6ToPx, Right: d.pageSize.W - mm6ToPx, Bottom: d.pageSize.H - mm6ToPx}
	d.pdf.Start(gopdf.Config{
		PageSize: *d.pageSize,
		TrimBox:  trimbox,
	})
	err = d.pdf.AddTTFFontByReader("regular", bytes.NewReader(openSansRegular))
	if err != nil {
		return err
	}
	err = d.pdf.AddTTFFontByReader("bold", bytes.NewReader(openSansBold))
	if err != nil {
		return err
	}
	err = d.pdf.AddTTFFontByReader("italic", bytes.NewReader(openSansItalic))
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
	d.colWidth = d.wapBox.W / float64(d.bigColumns)
	d.minuteHeight = d.wapBox.H / duration.Minutes()
	return nil
}

func (d *PDFDrawer) drawHeaderAndFooter(
	topLeft, topMiddle, topRight string,
	botLeft, botMiddle, botRight string,
) {
	padding := mmToPx(4)
	weekCounter := 1
	d.pdf.AddHeader(func() {
		err := d.pdf.SetFontSize(d.largeFontSize)
		checkPDFerror(err)
		d.pdf.SetY(padding)
		d.pdf.SetX(padding)
		err = d.pdf.CellWithOption(nil, topLeft, gopdf.CellOption{Align: gopdf.Left})
		checkPDFerror(err)
		topMiddleCopy := fmt.Sprintf("%s - Woche %d", topMiddle, weekCounter)
		weekCounter++
		tmW, err := d.pdf.MeasureTextWidth(topMiddleCopy)
		checkPDFerror(err)
		d.pdf.SetX(d.pageSize.W/2 - tmW/2)
		err = d.pdf.CellWithOption(nil, topMiddleCopy, gopdf.CellOption{Align: gopdf.Center})
		checkPDFerror(err)
		trW, err := d.pdf.MeasureTextWidth(topRight)
		checkPDFerror(err)
		d.pdf.SetX(d.pageSize.W - trW - padding)
		err = d.pdf.CellWithOption(nil, topRight, gopdf.CellOption{Align: gopdf.Right})
		checkPDFerror(err)
	})
	d.pdf.AddFooter(func() {
		err := d.pdf.SetFontSize(d.smallFontSize)
		checkPDFerror(err)
		d.pdf.SetY(d.pageSize.H - padding - d.smallFontSize)
		d.pdf.SetX(padding)
		err = d.pdf.CellWithOption(nil, botLeft, gopdf.CellOption{Align: gopdf.Left})
		checkPDFerror(err)
		bmW, err := d.pdf.MeasureTextWidth(botMiddle)
		checkPDFerror(err)
		d.pdf.SetX(d.pageSize.W/2 - bmW/2)
		err = d.pdf.CellWithOption(nil, botMiddle, gopdf.CellOption{Align: gopdf.Center})
		checkPDFerror(err)
		brW, err := d.pdf.MeasureTextWidth(botRight)
		checkPDFerror(err)
		d.pdf.SetX(d.pageSize.W - brW - padding)
		err = d.pdf.CellWithOption(nil, botRight, gopdf.CellOption{Align: gopdf.Right})
		checkPDFerror(err)
	})
}

func (d *PDFDrawer) Draw(wap *Wap, outputPath string) {
	d.wap = wap
	err := d.setupDocument()
	checkPDFerror(err)
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
		for day := range daysInWeek {
			totalDayIdx := week*daysInWeek + day
			d.drawColumnHeader(week*daysInWeek + day)
			// we model the current footnote convention:
			// for each day the footnote counter resets
			// Monday it starts with 10, Tuesday 20, ...
			footnoteCounter := (day + 1) * 10
			remarks := []string{}
			// draw events for this day
			for _, elem := range layout {
				if elem.dayOffset == totalDayIdx {
					event := elem.Event
					// Handle footnotes
					if event.Footnote {
						footNoteSource := *event
						footNoteSource.Title = strconv.Itoa(footnoteCounter)
						footNoteSource.Description = ""
						elem.Event = &footNoteSource
						ts := MilitaryTime(event.Start)
						remarks = append(remarks, fmt.Sprintf("%d  %s %s, %s", footnoteCounter, ts, event.Title, event.Description))
						footnoteCounter += 1
					}
					d.drawEvent(elem)
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
	remarksHeight := 300.0
	d.pdf.SetStrokeColor(0x00, 0x00, 0x00)
	d.pdf.SetFillColor(0xff, 0xff, 0xff)
	drawRect(d.pdf, rectStart, gopdf.Rect{W: d.colWidth, H: remarksHeight * d.minuteHeight})
	rectStart.X += d.padding
	rectStart.Y += d.padding
	remarksRect := gopdf.Rect{W: d.colWidth - 2*d.padding, H: remarksHeight*d.minuteHeight - d.padding}
	d.drawMultiLineText(remarks, rectStart, remarksRect)
}

func (d *PDFDrawer) drawMultiLineText(text []string, point gopdf.Point, rect gopdf.Rect) {
	d.pdf.SetXY(point.X, point.Y)
	err := d.pdf.SetFont("regular", "", d.smallFontSize)
	checkPDFerror(err)
	for _, txt := range text {
		if txt == "" {
			point.Y += d.smallFontSize
			rect.H -= d.smallFontSize
			d.pdf.SetXY(point.X, point.Y)
			continue
		}
		for s := range strings.SplitSeq(txt, "\n") {
			if s == "" {
				continue
			}
			_, h, _ := d.pdf.IsFitMultiCell(&rect, s)
			err := d.pdf.MultiCellWithOption(&rect, s, gopdf.CellOption{BreakOption: d.breakOption})
			checkPDFerror(err)
			point.Y += h
			rect.H -= h
			d.pdf.SetXY(point.X, point.Y)
		}
	}
}

func (d *PDFDrawer) drawWeeklyRemarks(weekIdx int) {
	// Remarks
	detHeightMin := 90.0
	rectStart := d.toGridSystem(d.wap.dayStart, d.bigColumns-1)
	headerRect := gopdf.Rect{W: d.colWidth, H: detHeightMin * d.minuteHeight}
	d.pdf.SetStrokeColor(0x00, 0x00, 0x00)
	d.pdf.SetFillColor(0xff, 0xff, 0xff)
	minutes := d.wap.dayEnd.Sub(d.wap.dayStart).Minutes()
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
	checkPDFerror(err)
	err = d.pdf.CellWithOption(&headerRect, "Bemerkungen",
		gopdf.CellOption{
			Align: gopdf.Center | gopdf.Middle,
		})
	checkPDFerror(err)
	rectStart = d.toGridSystem(d.wap.dayStart, d.bigColumns-1)
	rectStart.X += d.padding
	rectStart.Y += d.padding
	colRect.H -= 2 * d.padding
	colRect.W -= 2 * d.padding
	for i, r := range d.wap.Remarks[weekIdx] {
		if r != "" {
			d.wap.Remarks[weekIdx][i] = "- " + r
		}
	}
	d.drawMultiLineText(d.wap.Remarks[weekIdx], rectStart, colRect)

	// Add a signature block
	signatureStart := d.toGridSystem(d.wap.dayEnd, d.bigColumns-1)
	signatureRect := gopdf.Rect{W: d.colWidth, H: 150 * d.minuteHeight}
	d.pdf.SetStrokeColor(0x00, 0x00, 0x00)
	d.pdf.SetFillColor(0xff, 0xff, 0xff)
	drawRect(d.pdf, signatureStart, signatureRect)
}

func (d *PDFDrawer) setupPage() {
	opt := gopdf.PageOption{
		PageSize: d.pageSize,
	}
	d.pdf.AddPageWithOption(opt)
	d.pdf.SetStrokeColor(0x80, 0x80, 0x80)
	d.pdf.SetLineWidth(.2)
	quarters := (d.wap.dayEnd.Sub(d.wap.dayStart)).Minutes() / 15.0
	drawHorizontalLines(d.pdf, d.p1, d.wapBox, int(quarters))
	// The Big Grid
	d.pdf.SetStrokeColor(0, 0, 0)
	d.pdf.SetLineWidth(.5)
	o1 := (60.0 - float64(d.wap.dayStart.Minute())) * d.minuteHeight
	if d.wap.dayStart.Minute() == 0 {
		o1 = 0.0
	}
	o2 := float64(d.wap.dayEnd.Minute()) * d.minuteHeight
	hourBox := d.wapBox
	hourBox.H -= o1 + o2
	hourPoint := d.p1
	hourPoint.Y += o1
	// round up to the next full hour
	tStart := d.wap.dayStart.Truncate(time.Hour)
	if tStart.Before(d.wap.dayStart) {
		tStart = tStart.Add(time.Hour)
	}
	// round dayEnd down to the next full hour
	tEnd := d.wap.dayEnd.Truncate(time.Hour)
	fullHours := tEnd.Sub(tStart).Hours()
	drawGrid(d.pdf, hourPoint, hourBox, int(fullHours), d.bigColumns)

	// Add time scale (mark all hours)
	err := d.pdf.SetFontSize(8)
	checkPDFerror(err)
	d.pdf.SetFillColor(0x00, 0x00, 0x00)
	d.pdf.SetStrokeColor(0x00, 0x00, 0x00)
	for hour := tStart.Hour(); hour <= tEnd.Hour(); hour += 1 {
		p := Add(d.toGridSystem(DayTime(hour, 0), 0), gopdf.Point{X: -20, Y: -d.smallFontSize})
		d.pdf.SetXY(p.X, p.Y)
		// convert to military time format
		err := d.pdf.Cell(nil, fmt.Sprintf("%02d00", hour))
		checkPDFerror(err)
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
	dayInWeek := totalDayOffset % daysInWeek
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
	checkPDFerror(err)
	d.pdf.SetTextColor(0x00, 0x00, 0x00)
	dayName := d.wap.dayNames[totalDayOffset]
	err = d.pdf.CellWithOption(&rect, dayName,
		gopdf.CellOption{
			Align: gopdf.Center | gopdf.Middle,
		})
	checkPDFerror(err)
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
		checkPDFerror(err)
		err = d.pdf.CellWithOption(&rect, colName,
			gopdf.CellOption{
				Align: gopdf.Center | gopdf.Middle,
			})
		checkPDFerror(err)
		d.pdf.RotateReset()
	}
}

func (d *PDFDrawer) drawEvent(elem EventPosition) {
	event := elem.Event
	if c, ok := d.wap.categories[event.Category]; ok {
		d.pdf.SetFillColor(c.bgColor.R, c.bgColor.G, c.bgColor.B)
		d.pdf.SetTextColor(c.textColor.R, c.textColor.G, c.textColor.B)
	}
	if elem.Event.OpenEnd {
		drawOpenEndRect(d.pdf, elem.P, elem.R)
	} else {
		drawRect(d.pdf, elem.P, elem.R)
	}
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
	err := d.pdf.SetFont("bold", "", titleFontSize)
	checkPDFerror(err)
	ok, heightNeeded, _ := d.pdf.IsFitMultiCell(&rect, title)
	if !ok {
		log.Println("WARNING", "title does not fit in rectangle for event: ", event)
	}
	err = d.pdf.MultiCellWithOption(&rect, title,
		gopdf.CellOption{
			Align:       gopdf.Center,
			BreakOption: d.breakOption,
		})
	checkPDFerror(err)
	d.pdf.SetXY(elem.P.X, elem.P.Y+heightNeeded)
	err = d.pdf.SetFont("regular", "", math.Min(d.smallFontSize, titleFontSize))
	checkPDFerror(err)
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
		checkPDFerror(err)
	}
}

type columnInfo struct {
	// Offset from the x of the day
	Offset float64
	// Width of the column
	W float64
}

// TODO(refactor) this is called from too many places
func (d *PDFDrawer) assignColumnLocations(columns []string, width float64) map[string]columnInfo {
	m := make(map[string]columnInfo)
	// divide evently
	if len(columns) == 0 {
		return m
	}
	// Treat Beso specially: otherwise it looks bad
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
	// Initialize event positions with starting point and rectangles
	pre := []EventPosition{}
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

		pre = append(pre, EventPosition{
			dayOffset: event.DayOffset,
			P:         RectStart,
			R:         gopdf.Rect{W: d.colWidth, H: height},
			Event:     &event})
	}
	// First pass: preprocess overlapping events
	for i, ev := range pre {
		if ev.Event.Repeats {
			continue
		}
		// Assumption: this can only happen for events in a single columns
		if len(ev.Event.AppearsIn) != 1 {
			continue
		}
		col := ev.Event.AppearsIn[0]
		// For example events ev1 and ev2 that overlap in time will have
		//	-----
		// | ev1 |-----|
		// |	 | ev2 |
		// -------------
		// ev1.parallelCols = ev2.parallelCols = 1	(the number of other events in this column)
		// ev1.parallelIdx = 1 and ev2.parallelIdx = 2
		overlapping := ev.parallelIdx
		// Search forward in time until we found all overlapping
		for j := i + 1; j < len(pre); j += 1 {
			next := pre[j]
			if next.Event.DayOffset != ev.Event.DayOffset {
				break
			}
			if ev.Event.End.Compare(next.Event.Start) <= 0 {
				break
			}
			if next.Event.Repeats {
				continue
			}
			if slices.Contains(next.Event.AppearsIn, col) {
				if len(next.Event.AppearsIn) != 1 {
					log.Printf("WARNING overlapping events %v and %v. \nSupported only in single columns, but got %v and %v", col, ev, next.Event, next.Event.AppearsIn)
					continue
				}
				pre[j].parallelIdx++
				overlapping++
			}
		}
		pre[i].parallelCols = overlapping
	}
	// if an event appears in multiple consecutive columns they can be merged
	// for ev1 that appears in columns A and B:
	// Merge the event over the columns:
	// | A | B |
	// |  ev1  |
	// if the columns are not adjacent, the event is printed in multiple ones
	// for ev1 that appears in columns A and C:
	// Copy the event:
	// | A   | B   | C   |
	// | ev1 | ... | ev2 |
	mergeAndCopy := func(elem EventPosition) (out []EventPosition) {
		event := elem.Event
		columnLocation := d.assignColumnLocations(d.wap.columns[elem.dayOffset], d.colWidth)
		// Consider the event to appear in all columns implicitly
		cols := event.AppearsIn
		if len(cols) == 0 {
			cols = slices.Clone(d.wap.columns[elem.dayOffset])
		}
		widthAccumulator := 0.0
		originalX := elem.P.X
		active := false // true if we are expanding a column
		for _, c := range d.wap.columns[elem.dayOffset] {
			if slices.Contains(cols, c) {
				widthAccumulator += columnLocation[c].W
				if !active {
					elem.P.X = originalX + columnLocation[c].Offset
				}
				active = true
			} else if active {
				elem.R.W = widthAccumulator
				out = append(out, elem)
				widthAccumulator = 0.0
				active = false
			}
		}
		if active {
			elem.R.W = widthAccumulator
			out = append(out, elem)
		}
		return
	}
	// Second pass
	for _, elem := range pre {
		event := elem.Event
		columnLocation := d.assignColumnLocations(d.wap.columns[event.DayOffset], d.colWidth)
		// Case 1: repeating events
		if event.Repeats {
			// Add all future instance of the repeating event
			// Later added events can drawn over them
			for day := range d.wap.Days {
				if event.DayOffset <= day {
					elem.P = d.toGridSystem(event.Start, day%7)
					elem.dayOffset = day
					res = append(res, mergeAndCopy(elem)...)
				}
			}
			continue
		}
		// Case 2: overlapping events
		if elem.parallelCols > 0 {
			// Example:
			// |   Det         |
			// | ev1 | ev2 |ev3|
			// We assumed len(elem.AppearsIn) = 1
			col := event.AppearsIn[0]
			width := columnLocation[col].W
			eventWidth := width / float64(elem.parallelCols+1)
			// Outer offset (e.g. for the column Det)
			for _, c := range d.wap.columns[event.DayOffset] {
				if col != c {
					elem.P.X += columnLocation[c].W
				} else {
					break
				}
			}
			// Inner offset (e.g. for ev2 the width of ev1 column)
			elem.P.X += float64(elem.parallelIdx) * eventWidth
			elem.R.W = eventWidth
			res = append(res, elem)
			continue
		}
		// Case 3: any other event
		res = append(res, mergeAndCopy(elem)...)
	}
	return
}
