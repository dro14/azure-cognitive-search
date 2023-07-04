package types

type Activity struct {
	MessageID    int    `json:"message_id"`
	Message      string `json:"message"`
	Date         int    `json:"date"`
	UserID       int64  `json:"user_id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
	IsPremium    string `json:"is_premium"`
}
