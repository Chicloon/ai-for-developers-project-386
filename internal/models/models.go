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
	ID        string `json:"id"`
	Date      string `json:"date"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	IsBooked  bool   `json:"isBooked"`
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

// User represents a registered user
type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	PasswordHash string `json:"-"` // never expose in JSON
	CreatedAt    string `json:"createdAt,omitempty"`
	UpdatedAt    string `json:"updatedAt,omitempty"`
}

// AuthRequest for login/register
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse returned after successful auth
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// Schedule represents user's availability (replaces AvailabilityRule)
type Schedule struct {
	ID        string  `json:"id"`
	UserID    string  `json:"userId"`
	Type      string  `json:"type"`
	DayOfWeek *int32  `json:"dayOfWeek,omitempty"`
	Date      *string `json:"date,omitempty"`
	StartTime string  `json:"startTime"`
	EndTime   string  `json:"endTime"`
	IsBlocked bool    `json:"isBlocked"`
	CreatedAt string  `json:"createdAt,omitempty"`
}

type CreateScheduleRequest struct {
	Type      string  `json:"type"`
	DayOfWeek *int32  `json:"dayOfWeek,omitempty"`
	Date      *string `json:"date,omitempty"`
	StartTime string  `json:"startTime"`
	EndTime   string  `json:"endTime"`
	IsBlocked bool    `json:"isBlocked"`
}

// VisibilityGroup for access control
type VisibilityGroup struct {
	ID              string `json:"id"`
	OwnerID         string `json:"ownerId"`
	Name            string `json:"name"`
	VisibilityLevel string `json:"visibilityLevel"`
	CreatedAt       string `json:"createdAt,omitempty"`
}

type CreateGroupRequest struct {
	Name            string `json:"name"`
	VisibilityLevel string `json:"visibilityLevel"`
}

type AddMemberRequest struct {
	Email  *string `json:"email,omitempty"`
	UserID *string `json:"userId,omitempty"`
}

// GroupMember with user info
type GroupMember struct {
	ID      string `json:"id"`
	GroupID string `json:"groupId"`
	Member  User   `json:"member"`
	AddedBy string `json:"addedBy"`
	AddedAt string `json:"addedAt"`
}

// Booking with user info
type BookingWithUser struct {
	ID          string  `json:"id"`
	ScheduleID  string  `json:"scheduleId"`
	Booker      User    `json:"booker"`
	Owner       User    `json:"owner"`
	Date        string  `json:"date"`
	StartTime   string  `json:"startTime"`
	EndTime     string  `json:"endTime"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"createdAt,omitempty"`
	CancelledAt *string `json:"cancelledAt,omitempty"`
}

type CreateBookingRequest struct {
	OwnerID    string `json:"ownerId"`
	ScheduleID string `json:"scheduleId"`
}

// Slot for public display
type PublicSlot struct {
	ID        string `json:"id"`
	Date      string `json:"date"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	IsBooked  bool   `json:"isBooked"`
}
