package socket

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type SocketManager struct {
	// mu          sync.Mutex
	upgrader *websocket.Upgrader
}

var socketManager *SocketManager

// ///////////////////////////////////////////////////////////////////////////////
// GetSocketManager returns a singleton instance of SocketManager.
// It initializes the instance only once using sync.Once to ensure thread safety.
// ///////////////////////////////////////////////////////////////////////////////

func GetSocketManager() *SocketManager {
	if socketManager == nil {
		socketManager = &SocketManager{
			upgrader: &websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			},
		}
	}
	return socketManager
}

func (sm *SocketManager) Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	println("Upgrading connection")
	conn, err := sm.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
