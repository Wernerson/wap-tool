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
	padding := mmToPx(1.5)
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
	unit := ""
	if wap.data.Meta.Unit != nil {
		unit = *wap.data.Meta.Unit
	}
	version := time.Now().Format(time.DateOnly)
	if wap.data.Meta.Version != nil {
		version = *wap.data.Meta.Version
	}
	author := wap.data.Meta.Author
	title := wap.data.Meta.Title
	producer := "WAP-tool " + VERSION
	d.pdf.SetInfo(gopdf.PdfInfo{
		Author:   author,
		Title:    title,
		Producer: producer,
	})
	d.drawHeaderAndFooter(
		unit,
		title,
		version,
		"",
		"made with "+producer,
		author)

	columnOptions := make([]map[string]columnInfo, wap.Days)
	for i := range wap.data.Days {
		weekday := i % 7
		if weekday == 0 {
			// start a new week
			d.setupPage()
		}
		// Draw the column header
		columnInfos := d.assignColumnLocations(wap.columns[i], d.colWidth)
		columnOptions[i] = columnInfos
		d.drawColumnHeader(i, columnInfos)
		d.drawRemarks()
		// draw repeating events
		for _, event := range wap.events {
			if event.repeats && event.dayOffset <= i {
				event.dayOffset = i % 7
				d.drawEvent(event, 0, d.colWidth)
			}
		}
		// draw events for this day
		for _, event := range wap.events {
			if event.dayOffset != i {
				continue
			}
			eventWidth := 0.0
			offset := -1.0
			appears := event.json.AppearsIn
			// TODO handle the special case where columns = [A, B, C] and appearsIn = [A, C]
			for _, c := range d.wap.columns[event.dayOffset] {
				if slices.Contains(appears, c) {
					eventWidth += columnOptions[event.dayOffset][c].W
					// ugly, just find the first column
					if offset < 0.0 {
						offset = columnOptions[event.dayOffset][c].Offset
					}
				}
			}
			if event.parallelCols > 0 {
				// Assumption: this can only happen for events in a single columns
				// Example:
				// |   Det         |
				// | ev1 | ev2 |ev3|
				eventWidth = eventWidth / float64(event.parallelCols+1)
				offset += float64(event.parallelIdx) * eventWidth
			}
			if eventWidth > 0.0 {
				d.drawEvent(event, offset, eventWidth)
			} else {
				log.Println("WARNING appearsIn is empty. Will print the event full width: ", event)
				d.drawEvent(event, offset, d.colWidth)
			}
		}
	}
	log.Println("INFO writing pdf to ", outputPath)
	d.pdf.WritePdf(outputPath)
	return nil
}

func (d *PDFDrawer) drawRemarks() {
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
	for _, remark := range d.wap.data.Remarks {
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
func (d *PDFDrawer) drawColumnHeader(dayOffset int, ci map[string]columnInfo) {
	day := dayOffset % 7
	detHeightMin := 90.0
	dayHeightMin := 20.0
	// Box for the week
	d.pdf.SetStrokeColor(0x00, 0x00, 0x00)
	d.pdf.SetFillColor(0xf0, 0xf0, 0xf0)
	RectStart := Add(d.toGridSystem(d.wap.dayStart, day),
		gopdf.Point{X: 0, Y: -(detHeightMin + dayHeightMin) * d.minuteHeight})
	rect := gopdf.Rect{W: d.colWidth, H: dayHeightMin * d.minuteHeight}
	drawRect(d.pdf, RectStart, rect)
	d.pdf.SetXY(RectStart.X, RectStart.Y)
	d.pdf.SetFont("bold", "", 6)
	d.pdf.SetTextColor(0x00, 0x00, 0x00)
	dayName := d.wap.dayNames[dayOffset]
	d.pdf.CellWithOption(&rect, dayName,
		gopdf.CellOption{
			Align: gopdf.Center | gopdf.Middle,
		})

	d.pdf.SetStrokeColor(0x00, 0x00, 0x00)
	d.pdf.SetFillColor(0xf0, 0xf0, 0xf0)
	// empty box if no columns are defined
	if len(ci) == 0 {
		RectStart := Add(d.toGridSystem(d.wap.dayStart, day),
			gopdf.Point{X: 0, Y: -detHeightMin * d.minuteHeight})
		rect := gopdf.Rect{W: d.colWidth, H: detHeightMin * d.minuteHeight}
		drawRect(d.pdf, RectStart, rect)
	}
	for colName, opts := range ci {
		RectStart := Add(d.toGridSystem(d.wap.dayStart, day),
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

func (d *PDFDrawer) drawEvent(event Event, offset, width float64) {
	cat := event.json.Category
	if cat == nil {
		d.pdf.SetFillColor(127, 127, 127)
	} else if c, ok := d.wap.colors[*cat]; ok {
		d.pdf.SetFillColor(c.R, c.G, c.B)
	} else {
		d.pdf.SetFillColor(127, 127, 127)
	}
	RectStart := d.toGridSystem(event.start, event.dayOffset%7)
	RectStart.X += offset
	minutes := event.end.Sub(event.start).Minutes()
	if minutes < 0 {
		return
	}
	rect := gopdf.Rect{W: width, H: minutes * d.minuteHeight}
	drawRect(d.pdf, RectStart, rect)

	d.pdf.SetXY(RectStart.X, RectStart.Y)
	d.pdf.SetTextColor(0x00, 0x00, 0x00)
	title := event.json.Title

	titleFontSize := 7
	// Limit the size for the title
	// to avoid making it too large
	rect.H = Min(rect.H, float64(titleFontSize)*2)
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
		log.Println("WARNING", "title does not fit in rectangle:", event.json.Title)
	}
	err := d.pdf.MultiCellWithOption(&rect, title,
		gopdf.CellOption{
			Align: gopdf.Center,
		})
	check(err)
	description := ""
	d.pdf.SetXY(RectStart.X, RectStart.Y+heightNeeded)
	if event.json.Location != nil {
		description += *event.json.Location
	}
	if event.json.Responsible != nil {
		description += ", " + *event.json.Responsible
	}

	d.pdf.SetFont("regular", "", 6)
	ok, _, _ = d.pdf.IsFitMultiCell(&rect, description)
	if !ok {
		log.Println("WARNING description does not fit: ", description)
	} else {
		d.pdf.MultiCellWithOption(&gopdf.Rect{W: width, H: minutes*d.minuteHeight - heightNeeded}, description,
			gopdf.CellOption{
				Align: gopdf.Center,
			})
	}
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
