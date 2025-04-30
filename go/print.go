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
	pdf      *gopdf.GoPdf
	pageSize *gopdf.Rect
}

func NewPDFDrawer() *PDFDrawer {
	pdf := gopdf.GoPdf{}
	return &PDFDrawer{pdf: &pdf, pageSize: gopdf.PageSizeA4Landscape}
}

func (d *PDFDrawer) setupPage() (err error) {
	mm6ToPx := mmToPx(6)
	trimbox := gopdf.Box{Left: mm6ToPx, Top: mm6ToPx, Right: d.pageSize.W - mm6ToPx, Bottom: d.pageSize.H - mm6ToPx}
	d.pdf.Start(gopdf.Config{
		PageSize: *d.pageSize,
		TrimBox:  trimbox,
	})
	d.pdf.AddPage()
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
	return nil
}

func (d *PDFDrawer) Draw(wap *Wap, outputPath string) (err error) {
	d.setupPage()
	unit := ""
	if wap.data.Meta.Unit != nil {
		unit = *wap.data.Meta.Unit
	}
	version := time.Now().Format(time.DateOnly)
	if wap.data.Meta.Version != nil {
		version = *wap.data.Meta.Version
	}
	padding := mmToPx(1.5)
	d.pdf.AddHeader(func() {
		d.pdf.SetY(padding)
		d.pdf.SetX(padding)
		d.pdf.Cell(nil, unit)
		tm := wap.data.Meta.Title
		tmW, err := d.pdf.MeasureTextWidth(tm)
		check(err)
		d.pdf.SetX(d.pageSize.W/2 - tmW/2)
		d.pdf.CellWithOption(nil, tm, gopdf.CellOption{Align: gopdf.Center})
		tr := version
		trW, err := d.pdf.MeasureTextWidth(tr)
		check(err)
		d.pdf.SetX(d.pageSize.W - trW - padding)
		d.pdf.CellWithOption(nil, tr, gopdf.CellOption{Align: gopdf.Right})
	})
	d.pdf.AddFooter(func() {
		footerText := "footer"
		ftH, err := d.pdf.MeasureCellHeightByText(footerText)
		check(err)
		d.pdf.SetY(d.pageSize.H - padding - ftH)
		d.pdf.Cell(nil, "footer")
	})
	// Page trim-box
	opt := gopdf.PageOption{
		PageSize: d.pageSize,
	}
	d.pdf.AddPageWithOption(opt)

	d.pdf.SetStrokeColor(255, 0, 0)
	d.pdf.SetLineWidth(1)
	PL := mmToPx(25)
	PR := PL
	PT := mmToPx(20)
	PB := mmToPx(30)
	P1 := gopdf.Point{X: PL, Y: PL}
	wapBox := gopdf.Rect{W: d.pageSize.W - PL - PR, H: d.pageSize.H - PT - PB}
	// [ ] Find the correct Start and bounds
	DAYS := 7
	duration := wap.dayEnd.Sub(wap.dayStart)
	HOURS := int(duration.Hours())
	SMALL_COLS := 5
	colWidth := wapBox.W / float64(DAYS)
	minuteHeight := wapBox.H / duration.Minutes()
	ToGridSystem := func(t time.Time, dayIndex int) gopdf.Point {
		deltaX := float64(dayIndex) * colWidth
		deltaY := t.Sub(wap.dayStart).Minutes() * minuteHeight
		return Add(P1, gopdf.Point{X: deltaX, Y: deltaY})
	}
	Grid(d.pdf, P1, wapBox, HOURS, DAYS)
	d.pdf.SetStrokeColor(0x80, 0x80, 0x80)
	d.pdf.SetLineWidth(0.5)
	Grid(d.pdf, P1, wapBox, HOURS*2, DAYS*SMALL_COLS)
	// Add time scale (mark all hours)
	d.pdf.SetFontSize(8)
	d.pdf.SetFillColor(0x00, 0x00, 0x00)
	d.pdf.SetStrokeColor(0x00, 0x00, 0x00)
	for hour := wap.dayStart.Hour(); hour <= wap.dayEnd.Hour(); hour += 1 {
		p := Add(ToGridSystem(DayTime(hour, 0), 0), gopdf.Point{X: -20, Y: -6})
		d.pdf.SetXY(p.X, p.Y)
		// convert to military time format
		d.pdf.Cell(nil, fmt.Sprintf("%02d00", hour))
	}
	columnOptions := make([]map[string]columnInfo, wap.Days)
	for i := range wap.Days {
		columnInfos := AssignColumns(wap.columns[i], colWidth)
		columnOptions[i] = columnInfos
		// draw the column header
		for colName, opts := range columnInfos {
			heightInMinutes := 90.0
			RectStart := Add(ToGridSystem(wap.dayStart, i),
				gopdf.Point{X: opts.Offset, Y: -heightInMinutes * minuteHeight})
			rect := gopdf.Rect{W: opts.W, H: heightInMinutes * minuteHeight}
			d.pdf.SetXY(RectStart.X, RectStart.Y)
			d.pdf.SetStrokeColor(0x00, 0x00, 0x00)
			d.pdf.SetFillColor(0xf0, 0xf0, 0xf0)
			PrintRect(d.pdf, RectStart, rect)
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

	drawEvent := func(event Event) {
		cat := event.json.Category
		if cat == nil {
			d.pdf.SetFillColor(127, 127, 127)
		} else if c, ok := wap.colors[*cat]; ok {
			d.pdf.SetFillColor(c.R, c.G, c.B)
		} else {
			d.pdf.SetFillColor(127, 127, 127)
		}
		// Adjust because of columns
		width := 0.0
		offset := -1.0
		appears := event.json.AppearsIn
		for _, c := range wap.columns[event.dayOffset] {
			if slices.Contains(appears, c) {
				width += columnOptions[event.dayOffset][c].W
				// ugly hack
				if offset < 0.0 {
					offset = columnOptions[event.dayOffset][c].Offset
				}
			}
		}
		// Special case. Repeating tasks could be defined on other days with different columns
		// Just print them full width.
		if event.repeats {
			width = colWidth
		}
		// TODO handle the case where columns = [A, B, C] and appearsIn = [A, C]
		RectStart := ToGridSystem(event.start, event.dayOffset)
		RectStart.X += offset
		minutes := event.end.Sub(event.start).Minutes()
		rect := gopdf.Rect{W: width, H: minutes * minuteHeight}
		PrintRect(d.pdf, RectStart, rect)
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
		err = d.pdf.MultiCellWithOption(&gopdf.Rect{W: width, H: minutes*minuteHeight - heightNeeded}, description,
			gopdf.CellOption{
				Align: gopdf.Center,
			})
		check(err)
	}
	// TODO handle columns issue for repeating tasks: just draw it full?
	for _, event := range wap.repeating {
		for idx := event.dayOffset; idx < wap.Days; idx += 1 {
			event.dayOffset = idx
			drawEvent(event)
		}
	}
	// TODO columns
	for _, event := range wap.events {
		drawEvent(event)
	}

	// possibly add more pages
	d.pdf.WritePdf(outputPath)
	return nil
}

type columnInfo struct {
	// Offset from the x of the day
	Offset float64
	// Width of the column
	W float64
}

func AssignColumns(columns []string, width float64) map[string]columnInfo {
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

func PrintRect(pdf *gopdf.GoPdf, p gopdf.Point, rect gopdf.Rect) {
	err := pdf.Rectangle(p.X, p.Y, p.X+rect.W, p.Y+rect.H, "DF", 0, 0)
	check(err)
}

func Add(p1, p2 gopdf.Point) gopdf.Point {
	return gopdf.Point{X: p1.X + p2.X, Y: p1.Y + p2.Y}
}

// Draw a grid on pdf where start is the top left corner and bounds is the size of the grid.
// rows and columns define the number of rows and columns.
func Grid(pdf *gopdf.GoPdf, start gopdf.Point, bounds gopdf.Rect, rows, columns int) {
	rowHeight := bounds.H / float64(rows)
	colWidth := bounds.W / float64(columns)
	// Draw horizontal lines ---
	for h := range rows + 1 {
		y := start.Y + rowHeight*float64(h)
		pdf.Line(start.X, y, start.X+bounds.W, y)
	}
	// Draw vertical lines |
	for w := range columns + 1 {
		x := start.X + colWidth*float64(w)
		pdf.Line(x, start.Y, x, start.Y+bounds.H)
	}
}
