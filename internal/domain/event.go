package domain

import "time"

type Event struct {
	EventId     int64     `json:"event_id"`
	UserId      int64     `json:"user_id"`
	Date        time.Time `json:"date"`
	IsArchived  bool      `json:"is_archived"`
	Description string    `json:"description"`
}
