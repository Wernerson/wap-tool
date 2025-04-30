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
	hoursPerDay  int
	minuteHeight float64
	colWidth     float64
}

func NewPDFDrawer() *PDFDrawer {
	pdf := gopdf.GoPdf{}
	return &PDFDrawer{pdf: &pdf, pageSize: gopdf.PageSizeA4Landscape}
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
	DAYS := 7
	duration := d.wap.dayEnd.Sub(d.wap.dayStart)
	d.hoursPerDay = int(duration.Hours())
	d.colWidth = d.wapBox.W / float64(DAYS)
	d.minuteHeight = d.wapBox.H / duration.Minutes()
	return nil
}

func (d *PDFDrawer) drawHeaderAndFooter(
	topLeft, topMiddle, topRight string,
	botLeft, botMiddle, botRight string,
) {
	headerFontSize := 12.0
	d.pdf.SetFontSize(headerFontSize)
	padding := mmToPx(1.5)
	d.pdf.AddHeader(func() {
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
	unit := ""
	if wap.data.Meta.Unit != nil {
		unit = *wap.data.Meta.Unit
	}
	version := time.Now().Format(time.DateOnly)
	if wap.data.Meta.Version != nil {
		version = *wap.data.Meta.Version
	}
	d.drawHeaderAndFooter(
		unit,
		wap.data.Meta.Title,
		version,
		"",
		"made with WAP-tool v0.1",
		"")

	// TODO() add more pages if there are more days
	// columns issue for repeating tasks: just draw it full
	d.setupPage()

	// TODO(refactor): make this a layouting subroutine
	columnOptions := make([]map[string]columnInfo, wap.Days)
	for i := range wap.data.Days {
		// TODO the indexing could panic
		columnInfos := d.assignColumnLocations(wap.columns[i], d.colWidth)
		columnOptions[i] = columnInfos
		// draw the column header
		for colName, opts := range columnInfos {
			heightInMinutes := 90.0
			RectStart := Add(d.toGridSystem(wap.dayStart, i),
				gopdf.Point{X: opts.Offset, Y: -heightInMinutes * d.minuteHeight})
			rect := gopdf.Rect{W: opts.W, H: heightInMinutes * d.minuteHeight}
			d.pdf.SetXY(RectStart.X, RectStart.Y)
			d.pdf.SetStrokeColor(0x00, 0x00, 0x00)
			d.pdf.SetFillColor(0xf0, 0xf0, 0xf0)
			drawRect(d.pdf, RectStart, rect)
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
	for _, event := range wap.repeating {
		for idx := event.dayOffset; idx < wap.Days; idx += 1 {
			event.dayOffset = idx
			// Special case. Repeating tasks could be defined on other days with different columns
			// Just print them full width.
			d.drawEvent(event, 0, d.colWidth)
		}
	}
	for _, event := range wap.events {
		width := 0.0
		offset := -1.0
		// TODO handle the special case where columns = [A, B, C] and appearsIn = [A, C]
		// TODO(refactor): make this a layouting subroutine
		appears := event.json.AppearsIn
		for _, c := range d.wap.columns[event.dayOffset] {
			log.Println(event, c)
			if slices.Contains(appears, c) {
				width += columnOptions[event.dayOffset][c].W
				// ugly hack
				if offset < 0.0 {
					offset = columnOptions[event.dayOffset][c].Offset
				}
			}
		}
		if width > 0.0 {
			d.drawEvent(event, offset, width)
		} else {
			log.Println("WARNING has no columns (appears in no columns): ", event)
		}
	}

	// possibly add more pages
	d.pdf.WritePdf(outputPath)
	return nil
}

func (d *PDFDrawer) setupPage() {
	opt := gopdf.PageOption{
		PageSize: d.pageSize,
	}
	DAYS := 7
	d.pdf.AddPageWithOption(opt)
	// The Big Grid
	d.pdf.SetStrokeColor(0, 0, 0)
	d.pdf.SetLineWidth(1)
	drawGrid(d.pdf, d.p1, d.wapBox, d.hoursPerDay, DAYS)
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

func (d *PDFDrawer) drawEvent(event Event, offset, width float64) {
	cat := event.json.Category
	if cat == nil {
		d.pdf.SetFillColor(127, 127, 127)
	} else if c, ok := d.wap.colors[*cat]; ok {
		d.pdf.SetFillColor(c.R, c.G, c.B)
	} else {
		d.pdf.SetFillColor(127, 127, 127)
	}
	RectStart := d.toGridSystem(event.start, event.dayOffset)
	RectStart.X += offset
	minutes := event.end.Sub(event.start).Minutes()
	rect := gopdf.Rect{W: width, H: minutes * d.minuteHeight}
	drawRect(d.pdf, RectStart, rect)
	smallFont := 6
	d.pdf.SetXY(RectStart.X, RectStart.Y-1)
	d.pdf.SetTextColor(0x00, 0x00, 0x00)
	d.pdf.SetFont("bold", "", smallFont)
	title := event.json.Title
	ok, heightNeeded, _ := d.pdf.IsFitMultiCell(&rect, title)
	if !ok {
		log.Println("WARNING", "event title does not fit in rectangle!")
	}
	err := d.pdf.MultiCellWithOption(&rect, title,
		gopdf.CellOption{
			Align: gopdf.Center,
		})
	check(err)
	description := ""
	d.pdf.SetXY(RectStart.X, RectStart.Y+heightNeeded-3)
	if event.json.Location != nil {
		description += *event.json.Location
	}
	if event.json.Responsible != nil {
		description += ", " + *event.json.Responsible
	}
	d.pdf.SetFont("regular", "", smallFont)
	err = d.pdf.MultiCellWithOption(&gopdf.Rect{W: width, H: minutes*d.minuteHeight - heightNeeded}, description,
		gopdf.CellOption{
			Align: gopdf.Center,
		})
	check(err)
}

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
