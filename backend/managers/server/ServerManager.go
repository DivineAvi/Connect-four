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
	"time"

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
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		println("Hello World")
		w.Write([]byte("Hello World"))
	})
	http.HandleFunc("/join", func(w http.ResponseWriter, r *http.Request) {
		println("Join request received")
		username := r.URL.Query().Get("username")
		roomId := r.URL.Query().Get("roomId")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if username == "" {
			http.Error(w, "Username is required", http.StatusBadRequest)
			return
		}

		if roomId == "" {
			http.Error(w, "Room ID is required", http.StatusBadRequest)
			return
		}

		roomie := room.GetRoomById(roomId)
		if roomie == nil {
			println("Room not found")
			http.Error(w, "Room not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Room is valid"))

	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		println("New connection")
		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "Username is required", http.StatusBadRequest)
			return
		}

		if roomId, b := sm.clientManager.GetPlayingClient(username); b {
			previousRoom, itexists := sm.roomManager.PlayingRooms[roomId]
			if itexists {
				if _, exists := previousRoom.DisconnectedPlayers[username]; !exists {

					log.Println("Username in use ❌")
					http.Error(w, "Username already in use", http.StatusConflict)
					return
				}
			}
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
	//////////////////////////////////////////////////////
	//MATCHMAKING : CHECKING IF THE USER IS ALREADY IN A GAME
	//////////////////////////////////////////////////////

	if roomId, exists := sm.clientManager.GetPlayingClient(username); exists {
		if sm.roomManager.PlayingRooms[roomId] != nil {
			conn.WriteJSON(types.SocketServerMessageType{
				Type: "info",
				Data: map[string]any{
					"info": "Previous game has been terminated",
				},
			})
			sm.clientManager.RemovePlayingClient(username)

			r := room.GetRoomById(roomId)
			r.PickWinner()
		} else {
			conn.WriteJSON(types.SocketServerMessageType{
				Type: "info",
				Data: map[string]any{
					"info": "Previous game was closed by the server",
				},
			})
		}

	}
	var r *room.Room

	//////////////////////////////////////////////////////
	//MATCHMAKING : SEARCHING FOR A ROOM IN WAITING ROOMS
	//////////////////////////////////////////////////////

	for roomId := range room.GetRoomManager().WaitingRooms {
		r := room.GetRoomById(roomId)
		r.AddPlayer(username, conn)
		sm.clientManager.AddPlayingClient(username, r.ID)
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

	go func() {
		time.Sleep(10 * time.Second)
		if r.Status == "playing" || r.TotalPlayers == room.PlayersNeeded {
			return
		}
		if sm.roomManager.WaitingRooms[r.ID] == nil {
			println("Room not found in waiting rooms")
			return
		}
		r.AddBot()
	}()

}

////////////////////////////////////////////////
// GAME UPDATE HANDLER
// Handles game updates like placing discs
////////////////////////////////////////////////

func GameUpdateHandler(sm *ServerManager, conn *websocket.Conn, username string, data map[string]any) {
	// Get room ID from the message
	roomId, ok := data["room_id"].(string)
	if !ok {
		conn.WriteJSON(types.SocketServerMessageType{
			Type: "error",
			Data: map[string]any{
				"error": "Invalid room ID",
			},
		})
		return
	}

	r := room.GetRoomById(roomId)
	if r == nil {
		conn.WriteJSON(types.SocketServerMessageType{
			Type: "error",
			Data: map[string]any{
				"error": "Room not found",
			},
		})
		return
	}

	if r.CurrentTurn != username {
		conn.WriteJSON(types.SocketServerMessageType{
			Type: "error",
			Data: map[string]any{
				"error": "Not your turn",
			},
		})
		return
	}

	action, ok := data["action"].(string)
	if !ok {
		conn.WriteJSON(types.SocketServerMessageType{
			Type: "error",
			Data: map[string]any{
				"error": "Invalid action",
			},
		})
		return
	}

	switch action {
	case "place_disc":
		column, okCol := data["column"].(float64)
		row, okRow := data["row"].(float64)
		playerColor, okColor := data["player_color"].(string)

		if !okCol || !okRow || !okColor {
			conn.WriteJSON(types.SocketServerMessageType{
				Type: "error",
				Data: map[string]any{
					"error": "Invalid column or row or color",
				},
			})
			return
		}

		r.GridData[int(column)][int(row)] = playerColor

		for playerName := range r.Players {
			if playerName != username {
				r.CurrentTurn = playerName
				break
			}
		}

		winner := checkForWin(r.GridData, playerColor)

		if winner != "" {
			println("Game won by", username, "with color", playerColor)
			r.Status = "finished"
			r.Winner = username

			if username != "bot" {
				println("Calling UpdatePlayerStats for winner:", username)
				r.UpdatePlayerStats(username)
			} else {
				println("Bot won, not updating player stats")
			}
		}

		// Notify all players about the update
		for playerName, playerConn := range r.Players {
			if playerName == "bot" {
				// Handle bot logic if needed
				continue
			}

			updateMsg := types.SocketServerMessageType{
				Type: "game_update",
				Data: map[string]any{
					"room_id":      r.ID,
					"status":       r.Status,
					"current_turn": r.CurrentTurn,
					"grid_data":    r.GridData,
				},
			}

			if r.Status == "finished" {
				updateMsg.Data["winner"] = r.Winner
			}

			err := playerConn.WriteJSON(updateMsg)
			if err != nil {
				println("Error sending game update to", playerName, ":", err.Error())
			}
		}

		if r.Status == "finished" {
			go func() {
				time.Sleep(5 * time.Second)
				for playerName := range r.Players {
					sm.clientManager.RemovePlayingClient(playerName)
				}
				r.DeleteRoom()
			}()
		} else if r.OpponentType == "bot" && r.CurrentTurn == "bot" {
			go r.MakeBotMove()
		}
	}
}

////////////////////////////////////////////////
// CHECK FOR WIN CONDITION COPY PASTED FROM ROOM MANAGER
////////////////////////////////////////////////

func checkForWin(grid [][]string, color string) string {
	for col := 0; col < len(grid); col++ {
		for row := 0; row < len(grid[col])-3; row++ {
			if grid[col][row] == color &&
				grid[col][row+1] == color &&
				grid[col][row+2] == color &&
				grid[col][row+3] == color {
				return color
			}
		}
	}

	for col := 0; col < len(grid)-3; col++ {
		for row := 0; row < len(grid[col]); row++ {
			if grid[col][row] == color &&
				grid[col+1][row] == color &&
				grid[col+2][row] == color &&
				grid[col+3][row] == color {
				return color
			}
		}
	}

	for col := 0; col < len(grid)-3; col++ {
		for row := 0; row < len(grid[col])-3; row++ {
			if grid[col][row] == color &&
				grid[col+1][row+1] == color &&
				grid[col+2][row+2] == color &&
				grid[col+3][row+3] == color {
				return color
			}
		}
	}

	for col := 0; col < len(grid)-3; col++ {
		for row := 3; row < len(grid[col]); row++ {
			if grid[col][row] == color &&
				grid[col+1][row-1] == color &&
				grid[col+2][row-2] == color &&
				grid[col+3][row-3] == color {
				return color
			}
		}
	}

	return ""
}

////////////////////////////////////////////////
// SOCKET HANDLER , HANDLES ALL SOCKET MESSAGES
////////////////////////////////////////////////

func handleSocket(sm *ServerManager, conn *websocket.Conn) {
	defer func() {
		username, exists := sm.clientManager.GetConnectionToUsername(conn)
		if !exists {
			println("Username not found for disconnected client")
			sm.clientManager.RemoveClient("", conn)
			conn.Close()
			return
		}

		if roomId, exists := sm.clientManager.GetPlayingClient(username); exists {
			println("Client disconnection starting for ", username, " in the room ", roomId)

			r := room.GetRoomById(roomId)
			if r != nil {
				r.DisconnectPlayer(username)
			}
		}

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
		case "game_update":
			go GameUpdateHandler(sm, conn, parsedMsg.Username, parsedMsg.Data)
		case "reconnect":
			go ReconnectHandler(sm, conn, parsedMsg.Username, parsedMsg.Data)
		default:
			log.Println("Unknown message type:", parsedMsg.Type)
		}
	}
}

////////////////////////////////////////////////
// RECONNECT HANDLER
// Handles player reconnection to a game
////////////////////////////////////////////////

func ReconnectHandler(sm *ServerManager, conn *websocket.Conn, username string, data map[string]any) {
	roomId, ok := data["room_id"].(string)
	if !ok {
		conn.WriteJSON(types.SocketServerMessageType{
			Type: "error",
			Data: map[string]any{
				"error": "Invalid room ID",
			},
		})
		return
	}

	r := room.GetRoomById(roomId)
	if r == nil {
		conn.WriteJSON(types.SocketServerMessageType{
			Type: "error",
			Data: map[string]any{
				"error": "Room not found",
			},
		})
		return
	}

	if r.Status == "finished" {
		winnerMsg := "The game has ended."
		if r.Winner != "" {
			if r.Winner == username {
				winnerMsg = "You won the game!"
			} else {
				winnerMsg = "You lost the game."
			}
		}

		conn.WriteJSON(types.SocketServerMessageType{
			Type: "game_update",
			Data: map[string]any{
				"room_id": r.ID,
				"status":  "finished",
				"winner":  r.Winner,
				"message": winnerMsg,
			},
		})
		return
	}

	_, wasDisconnected := r.DisconnectedPlayers[username]
	if !wasDisconnected {
		conn.WriteJSON(types.SocketServerMessageType{
			Type: "error",
			Data: map[string]any{
				"error": "You were not disconnected from this room",
			},
		})
		return
	}

	r.JoinPlayer(username, conn)

	sm.clientManager.AddPlayingClient(username, roomId)
}
