package slots

import (
	"testing"

	"call-booking/internal/models"
)

func TestGenerateSlotsFromRule(t *testing.T) {
	rule := models.AvailabilityRule{
		Type:      "recurring",
		DayOfWeek: ptrInt32(1), // Monday
		TimeRanges: []models.TimeRange{
			{StartTime: "10:00", EndTime: "11:00"},
		},
	}

	slots := generateSlotsFromRule(rule, "2026-04-06") // Monday

	if len(slots) != 2 {
		t.Fatalf("expected 2 slots, got %d", len(slots))
	}
	if slots[0].StartTime != "10:00" || slots[0].EndTime != "10:30" {
		t.Fatalf("unexpected first slot: %+v", slots[0])
	}
	if slots[1].StartTime != "10:30" || slots[1].EndTime != "11:00" {
		t.Fatalf("unexpected second slot: %+v", slots[1])
	}
}

func TestGenerateSlotsFromRuleMultipleRanges(t *testing.T) {
	rule := models.AvailabilityRule{
		Type: "one-time",
		Date: strPtr("2026-04-10"),
		TimeRanges: []models.TimeRange{
			{StartTime: "10:00", EndTime: "11:00"},
			{StartTime: "14:00", EndTime: "15:00"},
		},
	}

	slots := generateSlotsFromRule(rule, "2026-04-10")

	if len(slots) != 4 {
		t.Fatalf("expected 4 slots, got %d", len(slots))
	}
}

func TestGenerateSlotsEmptyWhenNoRules(t *testing.T) {
	slots := GenerateSlots(nil, nil, nil, "2026-04-06")
	// Default behavior: 9:00-18:00 = 18 slots (30 min intervals)
	if len(slots) != 18 {
		t.Fatalf("expected 18 slots (default 9-18), got %d", len(slots))
	}
}

func TestGenerateSlotsExcludesBlockedDay(t *testing.T) {
	rule := models.AvailabilityRule{
		Type:      "recurring",
		DayOfWeek: ptrInt32(1),
		TimeRanges: []models.TimeRange{
			{StartTime: "10:00", EndTime: "11:00"},
		},
	}
	blocked := []models.BlockedDay{{Date: "2026-04-06"}}

	slots := GenerateSlots([]models.AvailabilityRule{rule}, blocked, nil, "2026-04-06")
	if len(slots) != 0 {
		t.Fatalf("expected 0 slots on blocked day, got %d", len(slots))
	}
}

func TestAddMinutes(t *testing.T) {
	result := addMinutes("10:00", 30)
	if result != "10:30" {
		t.Fatalf("expected 10:30, got %s", result)
	}

	result = addMinutes("10:30", 30)
	if result != "11:00" {
		t.Fatalf("expected 11:00, got %s", result)
	}

	result = addMinutes("23:30", 30)
	if result != "00:00" {
		t.Fatalf("expected 00:00, got %s", result)
	}
}

func ptrInt32(v int32) *int32 { return &v }
func strPtr(v string) *string  { return &v }
