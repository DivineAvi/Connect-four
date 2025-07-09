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
	http.HandleFunc("/join", func(w http.ResponseWriter, r *http.Request) {
		username := r.URL.Query().Get("username")
		roomId := r.URL.Query().Get("roomId")
		if username == "" {
			http.Error(w, "Username is required", http.StatusBadRequest)
			return
		}

		// Check if room exists
		if roomId == "" {
			http.Error(w, "Room ID is required", http.StatusBadRequest)
			return
		}

		roomie := room.GetRoomById(roomId)
		if roomie == nil {
			http.Error(w, "Room not found", http.StatusNotFound)
			return
		}

		conn, err := sm.socketManager.Upgrade(w, r)
		if err != nil {
			http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
			return
		}

		sm.clientManager.AddClient(username, conn)
		println("Client added for handleSocket tracking")
		go handleSocket(sm, conn)

		// Check if the player is rejoining a game
		_, wasDisconnected := roomie.DisconnectedPlayers[username]

		// If the room is playing and the player was disconnected, handle rejoin
		if roomie.Status == "playing" && wasDisconnected {
			println("Player is rejoining a game in progress:", username)

			// Add player back to the room
			roomie.JoinPlayer(username, conn)

			// Add to playing clients
			sm.clientManager.AddPlayingClient(username, roomId)
			return
		}

		// If the room is playing but player wasn't disconnected, they can't join
		if roomie.Status == "playing" && !wasDisconnected {
			http.Error(w, "Game already in progress", http.StatusBadRequest)
			return
		}

		// If the room is waiting and not full, player can join
		if roomie.Status == "waiting" && roomie.TotalPlayers < room.PlayersNeeded {
			println("Player is joining a waiting room:", username)
			roomie.AddPlayer(username, conn)
			sm.clientManager.AddPlayingClient(username, roomId)
			return
		}

		// If the room is full
		if roomie.TotalPlayers == room.PlayersNeeded {
			http.Error(w, "Room is full", http.StatusBadRequest)
			return
		}
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		println("New connection")
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

	// Get the room
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

	// Check if it's the player's turn
	if r.CurrentTurn != username {
		conn.WriteJSON(types.SocketServerMessageType{
			Type: "error",
			Data: map[string]any{
				"error": "Not your turn",
			},
		})
		return
	}

	// Handle the action
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
		// Get column and row
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

		// Update the grid
		r.GridData[int(column)][int(row)] = playerColor

		// Change turn
		for playerName := range r.Players {
			if playerName != username {
				r.CurrentTurn = playerName
				break
			}
		}

		// Check for win condition (simplified)
		winner := checkForWin(r.GridData, playerColor)

		// Update game status if there's a winner
		if winner != "" {
			r.Status = "finished"
			r.Winner = username
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

		// If game is over, clean up
		if r.Status == "finished" {
			// Wait a bit to ensure messages are sent before cleanup
			go func() {
				time.Sleep(5 * time.Second)
				for playerName := range r.Players {
					sm.clientManager.RemovePlayingClient(playerName)
				}
				r.DeleteRoom()
			}()
		} else if r.OpponentType == "bot" && r.CurrentTurn == "bot" {
			// If it's the bot's turn, make a move
			go r.MakeBotMove()
		}
	}
}

// Simple function to check for a win condition
func checkForWin(grid [][]string, color string) string {
	// Check horizontally
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

	// Check vertically
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

	// Check diagonally (down-right)
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

	// Check diagonally (up-right)
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
		println("client disconnected")
		username, exists := sm.clientManager.GetConnectionToUsername(conn)
		if !exists {
			println("Username not found for disconnected client")
			sm.clientManager.RemoveClient("", conn)
			conn.Close()
			return
		}

		println("Client disconnected:", username)

		// Check if the client is in a game
		if roomId, exists := sm.clientManager.GetPlayingClient(username); exists {
			println("Player was in a game:", username, "in room:", roomId)

			// Get the room
			r := room.GetRoomById(roomId)
			if r != nil {
				// Handle player removal (which will trigger the reconnection timer)
				r.RemovePlayer(username)
			}
		}

		// Remove client from client manager
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

	// Get the room
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

	// Check if player was disconnected from this room
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

	// Rejoin the player to the room
	r.JoinPlayer(username, conn)

	// Add to playing clients
	sm.clientManager.AddPlayingClient(username, roomId)
}
