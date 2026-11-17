package meta

import "fmt"

type difficultyData struct {
	PlayLevel int
	Bpm       int
	Note      int
}

func timeCalc(length float64) string {
	for i := range 100 {
		if length-60*float64(i) < 60 {
			return fmt.Sprintf("%d:%04.1f", i, length-60*float64(i))
		}
	}
	return "0:00.0"
}
