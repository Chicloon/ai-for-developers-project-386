package models

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type TimeRange struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

type AvailabilityRule struct {
	ID         string      `json:"id"`
	Type       string      `json:"type"`
	DayOfWeek  *int32      `json:"dayOfWeek,omitempty"`
	Date       *string     `json:"date,omitempty"`
	TimeRanges []TimeRange `json:"timeRanges"`
}

type CreateAvailabilityRule struct {
	Type       string      `json:"type"`
	DayOfWeek  *int32      `json:"dayOfWeek,omitempty"`
	Date       *string     `json:"date,omitempty"`
	TimeRanges []TimeRange `json:"timeRanges"`
}

type BlockedDay struct {
	ID   string `json:"id"`
	Date string `json:"date"`
}

type CreateBlockedDay struct {
	Date string `json:"date"`
}

type Slot struct {
	ID         string `json:"id"`
	Date       string `json:"date"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
	IsBooked   bool   `json:"isBooked"`
}

type Booking struct {
	ID            string  `json:"id"`
	SlotDate      string  `json:"slotDate"`
	SlotStartTime string  `json:"slotStartTime"`
	Name          string  `json:"name"`
	Email         string  `json:"email"`
	Status        string  `json:"status"`
	Recurrence    *string `json:"recurrence,omitempty"`
	DayOfWeek     *int32  `json:"dayOfWeek,omitempty"`
	EndDate       *string `json:"endDate,omitempty"`
}

type CreateBooking struct {
	SlotDate      string  `json:"slotDate"`
	SlotStartTime string  `json:"slotStartTime"`
	Name          string  `json:"name"`
	Email         string  `json:"email"`
	Recurrence    *string `json:"recurrence,omitempty"`
	DayOfWeek     *int32  `json:"dayOfWeek,omitempty"`
	EndDate       *string `json:"endDate,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

var _ = time.Time{}
var _ = pgtype.Text{}
