package server

type SocketClientMessageType struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}
