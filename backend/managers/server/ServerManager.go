package server

import (
	"backend/managers/client"
	"backend/managers/room"
	"backend/managers/socket"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

///////////////////////////////////////////////
//STRUCTS AND VARIABLES DEFINATION
//////////////////////////////////////////////

type ServerManager struct {
	clientManager *client.ClientManager
	roomManager   *room.RoomManager
	socketManager *socket.SocketManager
}

var (
	serverManager *ServerManager
	once          sync.Once
)

//////////////////////////////////////////////
//Singleton ServerManager
//////////////////////////////////////////////

func GetServerManager() *ServerManager {
	once.Do(func() {
		serverManager = &ServerManager{
			clientManager: client.GetClientManager(),
			roomManager:   room.GetRoomManager(),
			socketManager: socket.GetSocketManager(),
		}
	})
	return serverManager
}

/////////////////////////////////////////////
// STARTS WEBSOCKET SERVER
/////////////////////////////////////////////

func (sm *ServerManager) StartServer() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "Username is required", http.StatusBadRequest)
			return
		}

		if _, b := sm.clientManager.GetClient(username); b {
			log.Println("Username in use ❌")
			http.Error(w, "Username already in use", http.StatusConflict)
			return
		}

		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
			return
		}

		sm.clientManager.AddClient(username, conn)

		handleSocket(sm, conn)

	})

	http.ListenAndServe(":8080", nil)
}

////////////////////////////////////////////////
// SOCKET HANDLER
////////////////////////////////////////////////

func handleSocket(sm *ServerManager, conn *websocket.Conn) {
	defer func() {
		sm.clientManager.RemoveClient("", conn)
		conn.Close()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				fmt.Println("WebSocket closed normally:", err)
			} else {
				fmt.Println("Read error:", err)
			}
			break
		}

		var parsedMsg SocketClientMessageType
		if err := json.Unmarshal(msg, &parsedMsg); err != nil {
			log.Println("JSON Unmarshal Error:", err)
			continue
		}

		switch parsedMsg.Type {
		case "new_game":
			log.Println("Chat message received:", parsedMsg.Data)

		default:
			log.Println("Unknown message type:", parsedMsg.Type)
		}
	}
}
