package manager

import (
	"github.com/gorilla/websocket"
)

type SocketManager struct {
	Connections map[string]*websocket.Conn
	// mu          sync.Mutex
}

var socketManagerInstance *SocketManager

// ///////////////////////////
// GetSocketManager returns a singleton instance of SocketManager.
// It initializes the instance only once using sync.Once to ensure thread safety.
// ///////////////////////////
func GetSocketManager() *SocketManager {
	if socketManagerInstance == nil {
		socketManagerInstance = &SocketManager{
			Connections: make(map[string]*websocket.Conn),
		}
	}
	return socketManagerInstance
}
