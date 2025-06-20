// JavaScript version using jsPDF
import jsPDF from "jspdf";
import openSansRegular from "./ttf/OpenSans-Regular-normal.ts"
import openSansItalic from "./ttf/OpenSans-Italic-normal.ts"
import openSansBold from "./ttf/OpenSans-Bold-normal.ts"
import type {WAP} from "@/wap";

// page paddings
const P = {
  L: 20, T: 15, R: 15, B: 15
}

const Font = {
  small: 4,
  medium: 6,
  large: 8
}

// offsets
const O = {
  Header: {x: 5, y: 8},
  Footer: {x: 5, y: 5},
}

// sizes
const S = {
  Note: 40,
  Day: 3,
  Det: 10,
  Beso: 5
}
// How much larger a line (including spacing) is compared to the text
const textSpacingFactor = 1.3;

type BoxOpts = {
  x: number, y: number, w: number, h: number,
  fillColor?: string,
  openEnd?: boolean
}

type TextBoxOpts = {
  textColor?: string,
  text?: string,
  subtext?: string,
  vertical?: boolean,
} & BoxOpts

export const printPdf = (wap: WAP) => {
  const pdf = new jsPDF({
    orientation: "landscape",
    unit: "mm",
    format: "a4"
  })

  // add fonts
  pdf.addFileToVFS('OpenSans-Regular-normal.ttf', openSansRegular);
  pdf.addFont('OpenSans-Regular-normal.ttf', 'OpenSans-Regular', 'normal');
  pdf.addFileToVFS('OpenSans-Regular-normal.ttf', openSansItalic);
  pdf.addFont('OpenSans-Regular-normal.ttf', 'OpenSans-Regular', 'italic');
  pdf.addFileToVFS('OpenSans-Regular-normal.ttf', openSansBold);
  pdf.addFont('OpenSans-Regular-normal.ttf', 'OpenSans-Regular', 'bold');
  pdf.setFont("OpenSans-Regular")

  // page dimensions
  const W = pdf.internal.pageSize.getWidth()
  const H = pdf.internal.pageSize.getHeight()

  // utility functions
  const normal = () => pdf.setFont("OpenSans-Regular", "normal")
  const bold = () => pdf.setFont("OpenSans-Regular", "bold")
  const small = () => pdf.setFontSize(Font.small)
  const medium = () => pdf.setFontSize(Font.medium)
  const large = () => pdf.setFontSize(Font.large)

  // x,y are center of first line of text
  const verticalText = (x: number, y: number, text: string, opts: any) => {
    const lineHeight = pdf.getLineHeightFactor() * textSpacingFactor;
    const textLines = opts.maxWidth ? pdf.splitTextToSize(text, opts.maxWidth) : [text];
    // align doesnt work together with angle, so do it all by hand
    for (let i = 0; i < textLines.length; i++) {
      const line = textLines[i];
      const textWidth = pdf.getTextWidth(line);
      pdf.text(line, x + (i)*lineHeight, y+textWidth/2, {
        baseline: "miiddle",
        angle: 90,
        ...opts
      })
    }
  }

  const horizontalText = (x: number, y: number, text: string, opts: any) => {
    pdf.text(text, x, y, {align: "center", baseline: "middle", ...opts})
  }

  const box = ({
                 x, y, h, w,
                 fillColor = "#F0F0F0",
                 openEnd = false
               }: BoxOpts) => {
    pdf.setFillColor(fillColor)
    if (openEnd) {
      pdf.moveTo(x, y)
      pdf.lineTo(x, y + h)
      const offset = Math.min(h, w) * 0.5
      pdf.curveTo(x + w / 2, y + h - offset, x + w / 2, y + h + offset, x + w, y + h)
      pdf.lineTo(x + w, y)
      pdf.close()
      pdf.fillStroke()
    } else pdf.rect(x, y, w, h, "FD")
  }

  const textbox = (x: number, y: number, w: number, h: number, texts: string[]) => {
    normal()
    box({x, y, h, w, fillColor: "#FFFFFF"})
    const textHeight = pdf.getLineHeightFactor()
    pdf.text(texts, x + 0.5 * textHeight, y + 0.5 * textHeight, {
      maxWidth: colWidth - 0.5 * textHeight,
      baseline: "top"
    })
  }

  const titlebox = (x: number, y: number, w: number, h: number, title: string, vertical: boolean) => {
    box({x, y, h, w, fillColor: "#F0F0F0"})
    if (vertical) verticalText(x + w / 2, y + h / 2, title, {maxWidth: h})
    else horizontalText(x + w / 2, y + h / 2, title, {maxWidth: w})
  }

  const activity = ({
                      x, y, h, w,
                      fillColor = "#F0F0F0",
                      textColor = "#000000",
                      text, subtext,
                      vertical = false,
                      openEnd = false
                    }: TextBoxOpts) => {
    box({x, y, h, w, openEnd, fillColor})
    if (!text) return

    pdf.setTextColor(textColor);
    const textHeight = pdf.getLineHeightFactor() * textSpacingFactor;
    bold()

    const allowedTextWidth = (vertical ? h : w) - 1.5
    const titleLines = pdf.splitTextToSize(text, allowedTextWidth).length;
    const subLines = subtext ? pdf.splitTextToSize(subtext, allowedTextWidth).length : 0;
    const totalLines = titleLines + subLines;
    const titleOffset = (totalLines - 1) * textHeight / 2;
    const subtextOffset = titleOffset - titleLines * textHeight;

    const opts = {maxWidth: allowedTextWidth};
    if (vertical) {
      verticalText(x + w / 2 - titleOffset, y + h / 2, text, opts)
      if (subtext) {
        normal()
        verticalText(x + w / 2 - subtextOffset, y + h / 2, subtext, opts)
      }
    } else {
      horizontalText(x + w / 2, y + h / 2 - titleOffset, text, opts)
      if (subtext) {
        normal()
        horizontalText(x + w / 2, y + h / 2 - subtextOffset, subtext, opts)
      }
    }
  }

  const getTime = (text: string): number => {
    const splits = text.split(":")
    const hour = Number(splits[0])
    const minutes = Number(splits[1]) / 60
    return hour + minutes
  }

  // calculate base measures
  const colWidth = (W - P.L - P.R) / 8
  const timeStart = getTime(wap.meta?.startTime || "06:00")
  const timeEnd = getTime(wap.meta?.endTime || "23:30")
  const schedHeight = H - (P.T + S.Day + S.Det) - P.B - S.Note
  const lines = Math.round((timeEnd - timeStart) * 4)
  const gap = schedHeight / lines
  const firstDayString = wap.meta?.firstDay
  if (!firstDayString) throw Error("First day not set!")
  const firstDay = new Date(firstDayString)

  const getDateText = (week: number, day: number) => {
    const date = new Date()
    date.setDate(firstDay.getDate() + 7 * week + day)
    return date.toLocaleDateString("de-CH", {
      weekday: "long",
      year: "numeric",
      month: "long",
      day: "numeric"
    })
  }

  const categories: Map<string, {
    color: string,
    textColor?: string
  }> = (wap.categories || []).reduceRight(
    (acc, curr) => {
      acc.set(curr.identifier, {color: curr.color, textColor: curr.textColor})
      return acc
    }, new Map())

  // draw wap
  const pages = (wap.weeks?.length || -1)
  for (let p = 0; p < pages; ++p) {

    // get current week
    if (!wap.weeks || !wap.weeks[p]) throw Error(`Week ${p} does not exist!`)
    const week = wap.weeks[p]

    // header
    normal()
    large()
    const title = wap.meta?.title ? `${wap.meta?.title} - Woche ${p + 1}` : `Woche ${p + 1}`
    if (wap.meta?.unit) pdf.text(wap.meta?.unit, O.Header.x, O.Header.y)
    pdf.text(title, W / 2, O.Header.y, {align: "center"})
    if (wap.meta?.version) pdf.text(wap.meta?.version, W - O.Header.x, O.Header.y, {align: "right"})

    // footer
    small()
    normal()
    pdf.text("made with WAUI (WAP UI)", W / 2, H - O.Footer.y, {align: "center"})
    if (wap.meta?.author) pdf.text(wap.meta?.author, W - O.Footer.x, H - O.Footer.y, {align: "right"})

    // draw times and lines
    normal()
    large()
    for (let i = 0; i < lines; ++i) {
      const x = P.L
      const y = P.T + S.Day + S.Det + i * gap
      const fullHour = Number.isInteger(timeStart + i * 0.25)
      if (fullHour) {
        const lineHeight = pdf.getLineHeightFactor()
        const hour = timeStart + i * 0.25
        const time = `${hour >= 10 ? '' : '0'}${hour}00`
        pdf.text(time, x - 0.5 * lineHeight, y, {
          align: "right",
          baseline: "middle"
        })
        pdf.setLineWidth(0.5)
      } else pdf.setLineWidth(0.2)
      pdf.line(x, y, x + 7 * colWidth, y)
    }

    // fill all day columns
    small()
    for (let d = 0; d < 7; ++d) {
      // draw day separator line
      const xDay = P.L + (d * colWidth)
      bold()
      pdf.setLineWidth(0.5)
      pdf.line(xDay, P.T + S.Day + S.Det, xDay, P.T + S.Day + S.Det + schedHeight)

      // day title
      const dateText = getDateText(p, d)
      titlebox(P.L + d * colWidth, P.T, colWidth, S.Day, dateText, false)

      // det title
      const cols = week.days[d].columns || []
      const detWidth = (colWidth - S.Beso) / Math.max(cols.length, 1)

      for (let j = 0; j < cols.length; ++j) {
        const x = P.L + d * colWidth + j * detWidth
        const y = P.T + S.Day
        titlebox(x, y, detWidth, S.Det, cols[j], true)
      }
      titlebox(P.L + d * colWidth + Math.max(cols.length, 1) * detWidth, P.T + S.Day, S.Beso, S.Det, "Beso", true)

      // schedule
      const day = week.days[d]
      const items = (day.events || []).sort((a, b) => (a.zIndex || 0) - (b.zIndex || 0))
      for (let item of items) {
        small()
        bold()

        const evCols = item.appearsIn || cols
        const itemStart = Math.max(getTime(item.start), timeStart)
        const y = P.T + S.Day + S.Det + (itemStart - timeStart) * 4 * gap
        const itemEnd = Math.min(getTime(item.end), timeEnd)
        const h = (itemEnd - itemStart) * 4 * gap

        const category = item.category ? categories.get(item.category) : undefined
        const printBox = (x: number, w: number) => activity({
          x, y, h, w,
          fillColor: category?.color,
          textColor: category?.textColor,
          text: item.title,
          subtext: item.description,
          openEnd: item.openEnd,
          vertical: item.forceHorizontalText ? false : h > 2 * w
        })

        let span = 0
        for (let i = 0; i < cols.length; ++i) {
          if (evCols.includes(cols[i])) ++span
          else if (span > 0) {
            printBox(xDay + (i - span) * detWidth, span * detWidth)
            span = 0
          }
        }
        if (span > 0) printBox(xDay + (cols.length - span) * detWidth, span * detWidth)
      }

      // daily remarks
      const texts = (day.remarks || [])
        .filter(it => !it.start)
        .map(it => it.title)

      const besos = (day.remarks || [])
        .filter(it => it.start)
        .sort((a, b) => getTime(a.start!!) - getTime(b.start!!))

      const base = Math.ceil(Math.log10(Math.max(2, besos.length)))
      for(let i = 0; i < besos.length; ++i) {
        small()
        bold()

        const beso = besos[i]
        const itemStart = Math.max(getTime(beso.start!!), timeStart)
        const y = P.T + 13 + (itemStart - timeStart) * 4 * gap
        const itemEnd = Math.min(getTime(beso.end!!), timeEnd)
        const h = (itemEnd - itemStart) * 4 * gap
        const footnote = (d+1) * Math.pow(10, base) + i

        activity({
          x: xDay + colWidth - 5, y, h, w: S.Beso,
          fillColor: "#E0EF04",
          text: `${footnote}`,
        })
        const time = beso.start!!.split(":").join("")
        texts.push(`${footnote} ${time} ${beso.title}`)
      }
      textbox(P.L + d * colWidth, H - P.B - S.Note, colWidth, S.Note, texts)
    }

    // weekly remarks
    const weeklyRemarks = (week.remarks || []).map(it => {
      if (it.trim() === "") return ""
      else return `- ${it}`
    })
    titlebox(P.L + 7 * colWidth, P.T + S.Day, colWidth, S.Det, "Bemerkungen", false)
    textbox(P.L + 7 * colWidth, P.T + S.Day + S.Det, colWidth, schedHeight, weeklyRemarks)
    textbox(P.L + 7 * colWidth, H - P.B - S.Note, colWidth, S.Note / 2, [])

    // add signer
    const lineHeight = pdf.getLineHeight()
    bold()
    medium()
    if(wap.meta?.signerText) pdf.text(wap.meta?.signerText, P.L + 7 * colWidth + 0.5 * lineHeight, H - P.B - S.Note / 2 + 0.5 * lineHeight)

    if (p + 1 < pages) pdf.addPage()
  }

// safe the pdf
  pdf.save("WAP.pdf");
}
