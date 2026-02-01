package types

// CreateEventRequest запрос на создание события (user_id, date YYYY-MM-DD, event — текст)
type CreateEventRequest struct {
	UserID      int64  `json:"user_id" form:"user_id"`
	Date        string `json:"date" form:"date"` // YYYY-MM-DD
	Event       string `json:"event" form:"event"`
}

// UpdateEventRequest запрос на обновление события
type UpdateEventRequest struct {
	EventID     int64  `json:"event_id" form:"event_id"`
	UserID      int64  `json:"user_id" form:"user_id"`
	Date        string `json:"date" form:"date"`
	Event       string `json:"event" form:"event"`
}

// DeleteEventRequest запрос на удаление события
type DeleteEventRequest struct {
	EventID int64 `json:"event_id" form:"event_id"`
}
