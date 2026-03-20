package coverage

import "time"

// Session exercises fields with simple stdlib types (single-segment import
// path like "time").
type Session struct {
	ID        string
	CreatedAt time.Time
	TTL       time.Duration
}

// SessionContainers exercises container fields (maps, slices, arrays) whose
// element or key types are simple stdlib types.
type SessionContainers struct {
	ID        string
	Deadlines map[string]time.Time
	Limits    map[time.Duration]string
	Intervals []time.Duration
	Samples   [3]time.Duration
}
