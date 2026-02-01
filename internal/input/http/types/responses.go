package types

// APIResponse успешный ответ: {"result": "..."}
type APIResponse struct {
	Result string `json:"result"`
}

// ErrorResponse ответ с ошибкой: {"error": "описание ошибки"}
type ErrorResponse struct {
	Error string `json:"error"`
}

// EventsResponse ответ со списком событий (result может содержать JSON массива)
type EventsResponse struct {
	Result interface{} `json:"result"`
}
