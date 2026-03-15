package core

// UnifiedMessage represents the unified message format sent to the Buffy API.
type UnifiedMessage struct {
	UserID   string `json:"user_id"`
	Platform string `json:"platform"`
	Message  string `json:"message"`
}

// MessageReply is the reply object returned by the Buffy API.
type MessageReply struct {
	Reply string `json:"reply"`
}
