package domain

import "time"

type ContestResult struct {
	Title     string
	Time      time.Time
	Rating    float64
	Ranking   int
	Attended  bool
	HasRating bool
}

type ContestHistory struct {
	HasRating  bool
	Rating     float64
	TopPercent float64
	Results    []ContestResult
}

func (h ContestHistory) AttendedCount() int {
	n := 0
	for _, r := range h.Results {
		if r.Attended {
			n++
		}
	}
	return n
}
