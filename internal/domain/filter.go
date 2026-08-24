package domain

import "strings"

type DroneTaskFilter struct {
	MissionID string
	Status    DroneTaskStatus
	Search    string
}

func (f DroneTaskFilter) Normalize() DroneTaskFilter {
	f.MissionID = strings.TrimSpace(f.MissionID)
	f.Search = strings.TrimSpace(strings.ToLower(f.Search))
	return f
}

func (f DroneTaskFilter) Matches(s DroneTask) bool {
	f = f.Normalize()
	if f.MissionID != "" && s.MissionID != f.MissionID {
		return false
	}
	if f.Status != "" && s.Status != f.Status {
		return false
	}
	if f.Search != "" && !strings.Contains(strings.ToLower(s.TaskCode), f.Search) {
		return false
	}
	return true
}
