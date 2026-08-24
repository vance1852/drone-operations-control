package service

import (
	"errors"
	"testing"
	"time"

	"drone-operations-control/internal/domain"
)

func TestCreateDroneTasksBulkRejectsEmptyInput(t *testing.T) {
	svc := New(nil)
	if _, err := svc.CreateDroneTasksBulk(t.Context(), RequestMeta{}, nil); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("error=%v", err)
	}
}

func TestSearchRequestUsesStableDefaults(t *testing.T) {
	request := domain.SearchRequest{}
	request = request.Normalize()
	if request.Sort != domain.SortCreated || request.Limit != 50 || request.Offset != 0 {
		t.Fatalf("request=%+v", request)
	}
}

func TestValidateBulkForDroneUnitAcceptsSameDroneUnit(t *testing.T) {
	svc := New(nil).WithClock(func() time.Time { return time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC) })
	requests := []domain.DroneTaskRequest{{MissionID: "p", DroneUnitID: "s", TaskCode: "S-1", ExpiresAt: time.Date(2026, 8, 19, 9, 0, 0, 0, time.UTC)}}
	if err := svc.ValidateBulkForDroneUnit(requests, "s"); err != nil {
		t.Fatal(err)
	}
}
