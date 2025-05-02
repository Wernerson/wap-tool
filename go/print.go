package main

import (
	"fmt"
	"log"
	"slices"
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
	pdf          *gopdf.GoPdf
	wap          *Wap
	pageSize     *gopdf.Rect
	p1           gopdf.Point
	wapBox       gopdf.Rect
	bigColumns   int
	hoursPerDay  int
	minuteHeight float64
	colWidth     float64
}

func NewPDFDrawer() *PDFDrawer {
	pdf := gopdf.GoPdf{}
	return &PDFDrawer{pdf: &pdf, pageSize: gopdf.PageSizeA4Landscape, bigColumns: 8}
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
	err = d.pdf.SetFont("regular", "", 14)
	if err != nil {
		return err
	}

	// Padding left, top, right, and bottom from the page to the WAP Grid
	PL := mmToPx(25)
	PT := mmToPx(20)
	PR := PL
	PB := mmToPx(30)
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
	headerFontSize := 12.0
	padding := mmToPx(4)
	d.pdf.AddHeader(func() {
		d.pdf.SetFontSize(headerFontSize)
		d.pdf.SetY(padding)
		d.pdf.SetX(padding)
		d.pdf.CellWithOption(nil, topLeft, gopdf.CellOption{Align: gopdf.Left})
		tmW, err := d.pdf.MeasureTextWidth(topMiddle)
		check(err)
		d.pdf.SetX(d.pageSize.W/2 - tmW/2)
		d.pdf.CellWithOption(nil, topMiddle, gopdf.CellOption{Align: gopdf.Center})
		trW, err := d.pdf.MeasureTextWidth(topRight)
		check(err)
		d.pdf.SetX(d.pageSize.W - trW - padding)
		d.pdf.CellWithOption(nil, topRight, gopdf.CellOption{Align: gopdf.Right})
	})
	d.pdf.AddFooter(func() {
		d.pdf.SetFontSize(headerFontSize)
		d.pdf.SetY(d.pageSize.H - padding - headerFontSize)
		d.pdf.SetX(padding)
		d.pdf.CellWithOption(nil, botLeft, gopdf.CellOption{Align: gopdf.Left})
		bmW, err := d.pdf.MeasureTextWidth(botMiddle)
		check(err)
		d.pdf.SetX(d.pageSize.W/2 - bmW/2)
		d.pdf.CellWithOption(nil, botMiddle, gopdf.CellOption{Align: gopdf.Center})
		brW, err := d.pdf.MeasureTextWidth(botRight)
		check(err)
		d.pdf.SetX(d.pageSize.W - brW - padding)
		d.pdf.CellWithOption(nil, botRight, gopdf.CellOption{Align: gopdf.Right})
	})
}

func (d *PDFDrawer) Draw(wap *Wap, outputPath string) (err error) {
	d.wap = wap
	d.setupDocument()
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
		d.drawRemarks(week)
		for day := range 7 {
			d.drawColumnHeader(week*7 + day)
		}
		// draw events for this day
		for _, elem := range layout {
			if week*7 <= elem.dayOffset && elem.dayOffset < (week+1)*7 {
				d.drawEvent(elem)
			}
		}
	}
	log.Println("INFO writing pdf to ", outputPath)
	d.pdf.WritePdf(outputPath)
	return nil
}

func (d *PDFDrawer) drawRemarks(weekIdx int) {
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
	d.pdf.SetFont("bold", "", 6)
	d.pdf.CellWithOption(&headerRect, "Bemerkungen",
		gopdf.CellOption{
			Align: gopdf.Center | gopdf.Middle,
		})

	rectStart = d.toGridSystem(d.wap.dayStart, d.bigColumns-1)
	d.pdf.SetXY(rectStart.X, rectStart.Y)
	d.pdf.SetFont("regular", "", 6)
	for _, remark := range d.wap.Remarks[weekIdx] {
		txt := " - " + remark
		_, h, _ := d.pdf.IsFitMultiCell(&colRect, txt)
		d.pdf.MultiCell(&colRect, txt)
		rectStart.Y += h
		colRect.H -= h
		d.pdf.SetXY(rectStart.X, rectStart.Y)
	}
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
	d.pdf.SetFontSize(8)
	d.pdf.SetFillColor(0x00, 0x00, 0x00)
	d.pdf.SetStrokeColor(0x00, 0x00, 0x00)
	for hour := d.wap.dayStart.Hour(); hour <= d.wap.dayEnd.Hour(); hour += 1 {
		p := Add(d.toGridSystem(DayTime(hour, 0), 0), gopdf.Point{X: -20, Y: -6})
		d.pdf.SetXY(p.X, p.Y)
		// convert to military time format
		d.pdf.Cell(nil, fmt.Sprintf("%02d00", hour))
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
	d.pdf.SetFont("bold", "", 6)
	d.pdf.SetTextColor(0x00, 0x00, 0x00)
	dayName := d.wap.dayNames[totalDayOffset]
	d.pdf.CellWithOption(&rect, dayName,
		gopdf.CellOption{
			Align: gopdf.Center | gopdf.Middle,
		})

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
		d.pdf.SetFont("bold", "", 6)
		d.pdf.CellWithOption(&rect, colName,
			gopdf.CellOption{
				Align: gopdf.Center | gopdf.Middle,
			})
		d.pdf.RotateReset()
	}
}

func (d *PDFDrawer) drawEvent(elem EventPosition) {
	event := elem.Event
	if c, ok := d.wap.categories[event.Category]; ok {
		d.pdf.SetFillColor(c.R, c.G, c.B)
	}
	drawRect(d.pdf, elem.P, elem.R)
	d.pdf.SetXY(elem.P.X, elem.P.Y)
	d.pdf.SetTextColor(0x00, 0x00, 0x00)
	title := event.Title
	titleFontSize := 7
	// Limit the size for the title
	// to avoid making it too large
	rect := gopdf.Rect{
		W: elem.R.W,
		H: Min(elem.R.H, float64(titleFontSize)*2),
	}
	// Dynamically decrease font-size until it fits
	for ; titleFontSize >= 4; titleFontSize -= 1 {
		d.pdf.SetFont("bold", "", titleFontSize)
		ok, _, _ := d.pdf.IsFitMultiCell(&rect, title)
		if ok {
			break
		}
	}
	ok, heightNeeded, _ := d.pdf.IsFitMultiCell(&rect, title)
	if !ok {
		log.Println("WARNING", "title does not fit in rectangle:", event.Title)
	}
	err := d.pdf.MultiCellWithOption(&rect, title,
		gopdf.CellOption{
			Align: gopdf.Center,
		})
	check(err)
	d.pdf.SetXY(elem.P.X, elem.P.Y+heightNeeded)

	d.pdf.SetFont("regular", "", 6)
	ok, _, _ = d.pdf.IsFitMultiCell(&rect, event.Description)
	if !ok {
		log.Println("WARNING description does not fit: ", event.Description)
	} else {
		descriptionRect := gopdf.Rect{
			W: elem.R.W,
			H: elem.R.W - heightNeeded,
		}
		d.pdf.MultiCellWithOption(&descriptionRect, event.Description,
			gopdf.CellOption{
				Align: gopdf.Center,
			})
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
	columnWidth := width / float64(len(columns))
	accumulator := 0.0
	for _, c := range columns {
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
		// - Check for overlap
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
			// Assumption: this can only happen for events in a single columns
			// Example:
			// |   Det         |
			// | ev1 | ev2 |ev3|
			eventWidth := d.colWidth / float64(elem.parallelCols+1)
			offset := float64(elem.parallelIdx) * eventWidth
			res[i].P.X += offset
			res[i].R.W = eventWidth
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
		columnLocation := d.assignColumnLocations(d.wap.columns[event.DayOffset], d.colWidth)
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
	for _, e := range newPositions {
		res = append(res, e)
	}
	return
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
