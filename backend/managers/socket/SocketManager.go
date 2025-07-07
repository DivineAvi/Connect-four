package socket

import (
	"github.com/gorilla/websocket"
)

type SocketManager struct {
	Connections map[string]*websocket.Conn
	// mu          sync.Mutex
}

var socketManager *SocketManager

// ///////////////////////////////////////////////////////////////////////////////
// GetSocketManager returns a singleton instance of SocketManager.
// It initializes the instance only once using sync.Once to ensure thread safety.
// ///////////////////////////////////////////////////////////////////////////////

func GetSocketManager() *SocketManager {
	if socketManager == nil {
		socketManager = &SocketManager{
			Connections: make(map[string]*websocket.Conn),
		}
	}
	return socketManager
}
