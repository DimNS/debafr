package docker

import "github.com/docker/docker/pkg/jsonmessage"

// pullStats sums the layers of a pull up into the share of the image that is
// downloaded. Every layer is remembered, docker keeps sending the state of the
// layers it is working on and the bytes of a layer must be counted only once.
type pullStats struct {
	current map[string]int64
	total   map[string]int64
}

func newPullStats() *pullStats {
	return &pullStats{
		current: make(map[string]int64),
		total:   make(map[string]int64),
	}
}

// add takes the state of a layer as docker has just reported it and returns the
// state of the whole image.
func (s *pullStats) add(id string, p *jsonmessage.JSONProgress) PullProgress {
	// docker announces a layer with an empty progress and reports one again
	// when the layer is complete, the size shows up in between.
	if s.total[id] == 0 {
		s.total[id] = p.Total
	}
	s.current[id] = max(p.Current, s.current[id])

	var current, total int64
	for _, c := range s.current {
		current += c
	}
	for _, t := range s.total {
		total += t
	}

	return PullProgress{Current: current, Total: total}
}
