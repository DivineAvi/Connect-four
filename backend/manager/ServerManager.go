package manager

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type ServerManager struct {
	ClientManagerInstance *ClientManager
	RoomManagerInstance   *RoomManager
	SocketManagerInstance *SocketManager
}

var (
	serverManagerInstance *ServerManager
	once                  sync.Once
)

func GetServerManager() *ServerManager {
	once.Do(func() {
		serverManagerInstance = &ServerManager{
			ClientManagerInstance: GetClientManager(),
			RoomManagerInstance:   GetRoomManager(),
			SocketManagerInstance: GetSocketManager(),
		}
	})
	return serverManagerInstance
}

func StartServer() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}

		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "Username is required", http.StatusBadRequest)
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
			return
		}

		if _, b := clientManagerInstance.GetClient(username); b {
			http.Error(w, "Username already in use", http.StatusConflict)
			conn.Close()
			return
		}

		clientManagerInstance.AddClient(username, conn)
	})

	http.ListenAndServe(":8080", nil)
}
