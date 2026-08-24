package domain

import "time"

type ComplianceReport struct {
	MissionID         string
	GeneratedAt       time.Time
	Progress          MissionProgress
	Expiring          []DroneTask
	OpenInterventions int
}

func (r ComplianceReport) AtRisk() bool {
	return len(r.Expiring) > 0 || r.OpenInterventions > 0 || r.Progress.Rejected > 0
}

func (r ComplianceReport) Status() string {
	if r.AtRisk() {
		return "attention_required"
	}
	if r.Progress.Complete() {
		return "complete"
	}
	return "in_progress"
}
