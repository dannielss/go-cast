package models

type StreamResponse struct {
	StreamID string `json:"streamId"`
	Viewers  int    `json:"viewers"`
}
