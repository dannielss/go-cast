package models

type ChatMessage struct {
	ClientID string `json:"clientId"`
	Text     string `json:"text"`
}
