package main

import (
	"fmt"
	"sort"
	"time"
)

type Meeting struct {
	start time.Time
	end   time.Time
}

func parse(t string) time.Time {
	v, _ := time.Parse("15:04", t)
	return v
}

func main() {

	startLimit := parse("08:00")
	endLimit := parse("20:00")

	data := [][]string{
		{"7:00", "7:25"},
		{"8:00", "8:10"},
		{"7:45", "8:15"},
		{"9:30", "9:45"},
		{"15:10", "15:15"},
		{"17:40", "20:10"},
		{"21:00", "22:00"},
	}

	var meetings []Meeting

	for _, m := range data {
		s := parse(m[0])
		e := parse(m[1])

		if !s.Before(startLimit) && !e.After(endLimit) {
			meetings = append(meetings, Meeting{s, e})
		}
	}

	sort.Slice(meetings, func(i, j int) bool {
		return meetings[i].end.Before(meetings[j].end)
	})

	count := 0
	lastEnd := startLimit

	for _, m := range meetings {
		if !m.start.Before(lastEnd) {
			count++
			lastEnd = m.end
		}
	}

	fmt.Println("Max meetings:", count)
}
