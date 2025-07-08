package types

type SocketClientMessageType struct {
	Type     string         `json:"type"`
	Username string         `json:"username"`
	Data     map[string]any `json:"data"`
}

type SocketServerMessageType struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}
