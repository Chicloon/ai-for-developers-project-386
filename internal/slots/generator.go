package slots

import (
	"fmt"
	"time"

	"call-booking/internal/models"
)

func GenerateSlots(rules []models.AvailabilityRule, blockedDays []models.BlockedDay, bookings []models.Booking, date string) []models.Slot {
	// Check if day is blocked
	for _, bd := range blockedDays {
		if bd.Date == date {
			return []models.Slot{}
		}
	}

	// Parse the date to get day of week
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return []models.Slot{}
	}
	dayOfWeek := int32(t.Weekday()) // 0=Sunday

	// Collect matching rules
	var matchingRules []models.AvailabilityRule
	for _, rule := range rules {
		if rule.Type == "recurring" && rule.DayOfWeek != nil && *rule.DayOfWeek == dayOfWeek {
			matchingRules = append(matchingRules, rule)
		}
		if rule.Type == "one-time" && rule.Date != nil && *rule.Date == date {
			matchingRules = append(matchingRules, rule)
		}
	}

	// Default to 9:00-18:00 if no rules
	if len(matchingRules) == 0 {
		defaultRule := models.AvailabilityRule{
			TimeRanges: []models.TimeRange{
				{StartTime: "09:00", EndTime: "18:00"},
			},
		}
		matchingRules = append(matchingRules, defaultRule)
	}

	// Generate slots from rules
	var slots []models.Slot
	for _, rule := range matchingRules {
		slots = append(slots, generateSlotsFromRule(rule, date)...)
	}

	// Mark booked slots
	for i := range slots {
		for _, booking := range bookings {
			if booking.Status != "active" {
				continue
			}
			if isSlotBooked(slots[i], booking, date) {
				slots[i].IsBooked = true
				break
			}
		}
	}

	return slots
}

func generateSlotsFromRule(rule models.AvailabilityRule, date string) []models.Slot {
	var slots []models.Slot

	for _, tr := range rule.TimeRanges {
		current := tr.StartTime
		for current < tr.EndTime {
			end := addMinutes(current, 30)
			if end > tr.EndTime {
				break
			}
			slot := models.Slot{
				ID:        fmt.Sprintf("%s_%s", date, current),
				Date:      date,
				StartTime: current,
				EndTime:   end,
				IsBooked:  false,
			}
			slots = append(slots, slot)
			current = end
		}
	}

	return slots
}

func addMinutes(t string, minutes int) string {
	parsed, _ := time.Parse("15:04", t)
	result := parsed.Add(time.Duration(minutes) * time.Minute)
	return result.Format("15:04")
}

func isSlotBooked(slot models.Slot, booking models.Booking, date string) bool {
	if booking.SlotDate != date || booking.SlotStartTime != slot.StartTime {
		return false
	}

	rec := "none"
	if booking.Recurrence != nil {
		rec = *booking.Recurrence
	}

	switch rec {
	case "none":
		return true
	case "daily":
		if booking.EndDate == nil {
			return true
		}
		return date <= *booking.EndDate
	case "weekly":
		if booking.DayOfWeek == nil {
			return false
		}
		t, err := time.Parse("2006-01-02", date)
		if err != nil {
			return false
		}
		if int32(t.Weekday()) != *booking.DayOfWeek {
			return false
		}
		if booking.EndDate != nil && date > *booking.EndDate {
			return false
		}
		return true
	case "yearly":
		t, err := time.Parse("2006-01-02", date)
		if err != nil {
			return false
		}
		bookingDate, _ := time.Parse("2006-01-02", booking.SlotDate)
		if t.Month() != bookingDate.Month() || t.Day() != bookingDate.Day() {
			return false
		}
		if booking.EndDate != nil && date > *booking.EndDate {
			return false
		}
		return true
	default:
		return false
	}
}
