package pipeline

type Progress struct {
	Total     int
	Processed int
}

func (p Progress) Percent() float64 {
	if p.Total == 0 {
		return 0
	}
	return float64(p.Processed) / float64(p.Total) * 100
}
