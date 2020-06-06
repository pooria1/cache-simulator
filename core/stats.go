package core

import "math"

// Statistics represents cache statistics
type Statistics struct {
	Writes   uint64
	Reads    uint64
	Hits     uint64
	Misses   uint64
	Accesses uint64
	Replaces uint64
}

// MissRatio returns the miss ration of this set of stats, eg,
// the number of misses divided by the total number of hits and misses
func (s *Statistics) CalculateMissRate() float64 {
	if s.Accesses == 0 {
		return 0
	}
	return math.Round(float64(s.Misses)/float64(s.Hits+s.Misses)*10000) / 10000
}
