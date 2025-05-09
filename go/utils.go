package main

import (
	"fmt"
	"strconv"
	"time"
)

type RGBColor struct {
	R uint8 // Red component (0-255)
	G uint8 // Green component (0-255)
	B uint8 // Blue component (0-255)
}

func (c RGBColor) String() string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

// Compare compares two RGBColor structs.
// Returns -1 if the current color is less than the other,
// 0 if they are equal, and 1 if the current color is greater.
func (c RGBColor) Compare(other RGBColor) int {
	if c.R < other.R {
		return -1
	} else if c.R > other.R {
		return 1
	}
	if c.G < other.G {
		return -1
	} else if c.G > other.G {
		return 1
	}
	if c.B < other.B {
		return -1
	} else if c.B > other.B {
		return 1
	}
	return 0
}

// parseColor takes a hexadecimal color string (e.g., "#RRGGBB") and converts it into an RGBColor struct.
func parseColor(s string) (c RGBColor, err error) {
	if len(s) < 7 {
		return RGBColor{}, fmt.Errorf("ERROR Use the format #RRGGBB for colors. Invalid color string: %s", s)
	}
	rUint, err := strconv.ParseUint(s[1:3], 16, 8)
	if err != nil {
		return RGBColor{}, err
	}
	gUint, err := strconv.ParseUint(s[3:5], 16, 8)
	if err != nil {
		return RGBColor{}, err
	}
	bUint, err := strconv.ParseUint(s[5:7], 16, 8)
	if err != nil {
		return RGBColor{}, err
	}
	return RGBColor{uint8(rUint), uint8(gUint), uint8(bUint)}, nil
}

func parseDayTime(s string) (t time.Time, err error) {
	return time.Parse("15:04", s)
}

func DayTime(hour, minute int) time.Time {
	return time.Date(0, 1, 1, hour, minute, 0, 0, time.UTC)
}

func MilitaryTime(t time.Time) string {
	return fmt.Sprintf("%02d%02d", t.Hour(), t.Minute())
}

// 2024-04-23 -> 23.04.2024
// Format DD.MM.YYYY
func SwissDate(t time.Time) string {
	return fmt.Sprintf("%02d.%02d.%d", t.Day(), t.Month(), t.Year())
}

func RoundToQuarterHour(t time.Time) time.Time {
	minutes := t.Minute()
	roundedMinutes := ((minutes + 7) / 15) * 15
	if roundedMinutes == 60 {
		return t.Add(time.Hour).Truncate(time.Hour)
	}
	return t.Truncate(time.Hour).Add(time.Duration(roundedMinutes) * time.Minute)
}

func TranslateWeekDay(t time.Weekday) string {
	switch t {
	case time.Monday:
		return "Montag"
	case time.Tuesday:
		return "Dienstag"
	case time.Wednesday:
		return "Mittwoch"
	case time.Thursday:
		return "Donnerstag"
	case time.Friday:
		return "Freitag"
	case time.Saturday:
		return "Samstag"
	case time.Sunday:
		return "Sonntag"
	default:
		return "Unknown day"
	}
}
