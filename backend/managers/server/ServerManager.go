package server

import (
	"backend/managers/client"
	"backend/managers/room"
	"backend/managers/socket"
	"backend/managers/types"
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

		conn, _ := sm.socketManager.Upgrade(w, r)
		sm.clientManager.AddClient(username, conn)
		println("Client added for handleSocket tracking")
		go handleSocket(sm, conn)

	})

	http.ListenAndServe(":8080", nil)
}

////////////////////////////////////////////////
// NEW GAME HANDLER
////////////////////////////////////////////////

func NewGameHandler(sm *ServerManager, conn *websocket.Conn, username string) {
	log.Println("New game request received for ", username)

	//////////////////////////////////////////////////////
	//MATCHMAKING : CHECKING IF THE USER IS ALREADY IN A GAME
	//////////////////////////////////////////////////////

	if roomId, exists := sm.clientManager.GetPlayingClient(username); exists {
		conn.WriteJSON(types.SocketServerMessageType{
			Type: "info",
			Data: map[string]any{
				"info": "Previous game has been terminated",
			},
		})
		sm.clientManager.RemovePlayingClient(username)
		r := room.GetRoomById(roomId)
		r.PickWinner()
	}
	var r *room.Room

	//////////////////////////////////////////////////////
	//MATCHMAKING : SEARCHING FOR A ROOM IN WAITING ROOMS
	//////////////////////////////////////////////////////

	for roomId := range room.GetRoomManager().WaitingRooms {
		r := room.GetRoomById(roomId)
		r.AddPlayer(username, conn)
		sm.clientManager.AddPlayingClient(username, r.ID)
		r.ConverToPlaying()
		return
	}

	///////////////////////////////////////////////////////////////////////////
	//MATCHMAKING :IF NOT FOUND IN WAITING ROOMS , CREATING A NEW ROOM AND WAIT
	///////////////////////////////////////////////////////////////////////////

	r = room.CreateRoom(username, conn)
	sm.roomManager.WaitingRooms[r.ID] = r
	sm.clientManager.AddPlayingClient(username, r.ID)

	conn.WriteJSON(types.SocketServerMessageType{
		Type: "new_game_response",
		Data: map[string]any{
			"room_id":       r.ID,
			"status":        "waiting",
			"current_turn":  r.CurrentTurn,
			"total_players": r.TotalPlayers,
			"players":       r.Players,
			"grid_data":     r.GridData,
		},
	})

	/////////////////////////////
	// TIMER FOR BOT JOINING
	/////////////////////////////
	// go func() {
	// 	time.Sleep(10 * time.Second)
	// 	if r.Status == "playing" || r.TotalPlayers == room.PlayersNeeded {
	// 		return
	// 	}
	// 	r.AddPlayer("opponent", conn)
	// 	r.Status = "playing"
	// 	r.CurrentTurn = "opponent"
	// 	r.TotalPlayers = 2
	// 	conn.WriteJSON(types.SocketServerMessageType{
	// 		Type: "info_client_about_bot_joining",
	// 		Data: map[string]any{
	// 			"info": "Opponent bot has joined the game",
	// 		},
	// 	})
	// }()

}

////////////////////////////////////////////////
// SOCKET HANDLER , HANDLES ALL SOCKET MESSAGES
////////////////////////////////////////////////

func handleSocket(sm *ServerManager, conn *websocket.Conn) {
	defer func() {
		println("client disconnected")
		sm.clientManager.RemoveClient("", conn)
		conn.Close()
	}()

	for {

		_, msg, err := conn.ReadMessage()

		if err != nil {
			println("error in read message", err)
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				fmt.Println("WebSocket closed normally:", err)
			} else {
				fmt.Println("Read error:", err)
			}
			break
		}
		println("message received", string(msg))
		var parsedMsg types.SocketClientMessageType
		if err := json.Unmarshal(msg, &parsedMsg); err != nil {
			log.Println("JSON Unmarshal Error:", err)
			continue
		}

		switch parsedMsg.Type {
		case "new_game":

			go NewGameHandler(sm, conn, parsedMsg.Username)
		default:
			log.Println("Unknown message type:", parsedMsg.Type)
		}
	}
}
