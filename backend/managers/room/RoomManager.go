package room

import (
	"backend/managers/types"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

////////////////////////////////////
//   STRUCTURE AND VARIABLES
////////////////////////////////////

type Room struct {
	ID           string
	OpponentType string
	TotalPlayers int
	Players      map[string]*websocket.Conn // Maps player usernames to their presence in the room
	CurrentTurn  string                     // Username of the player whose turn it is
	GridData     [][]string                 // 2D slice representing the game board
	Status       string                     // waiting, playing, finished
	Winner       string
	Loser        string
	Draw         bool
}

type RoomManager struct {
	WaitingRooms map[string]*Room
	PlayingRooms map[string]*Room
	roomIdToRoom map[string]*Room
}

var roomManagerInstance *RoomManager = nil
var PlayersNeeded int = 2

var mu sync.Mutex

// ////////////////////////////////////////////////
// SINGLETON INSTACNE OF ROOM MANAGER
// ////////////////////////////////////////////////
func GetRoomManager() *RoomManager {
	if roomManagerInstance == nil {
		roomManagerInstance = &RoomManager{
			WaitingRooms: make(map[string]*Room),
			PlayingRooms: make(map[string]*Room),
			roomIdToRoom: make(map[string]*Room),
		}
	}
	return roomManagerInstance
}

//////////////////////////////////////////////
// CREATE ROOM FUNCTION
// CREATES A NEW ROOM AND RETURNS IT
//////////////////////////////////////////////

func CreateRoom(username string, conn *websocket.Conn) *Room {
	RoomId := uuid.New().String()
	Room := &Room{
		ID:           RoomId,
		GridData:     make([][]string, 6),
		Players:      make(map[string]*websocket.Conn),
		Status:       "waiting",
		CurrentTurn:  username,
		TotalPlayers: 1,
		Winner:       "",
		Loser:        "",
		Draw:         false,
	}
	Room.Players[username] = conn
	roomManagerInstance.roomIdToRoom[RoomId] = Room
	for i := range Room.GridData {
		Room.GridData[i] = make([]string, 7)
		for j := range Room.GridData[i] {
			Room.GridData[i][j] = "neutral"
		}
	}

	return Room
}

/////////////////////////////////////////////////
// ADDS A BOT TO THE ROOM
////////////////////////////////////////////////

func (r *Room) AddBot() {
	mu.Lock()
	defer mu.Unlock()
	println("Adding bot to room", r.ID)
	r.OpponentType = "bot"
	r.TotalPlayers++
	r.Players["bot"] = nil
	if r.TotalPlayers == PlayersNeeded {
		println("Total players reached", r.TotalPlayers)
		go r.ConverToPlaying()
	}

}

////////////////////////////////////////////////
// ADDS A PLAYER TO THE ROOM
///////////////////////////////////////////////

func (r *Room) AddPlayer(username string, conn *websocket.Conn) {
	mu.Lock()
	defer mu.Unlock()
	println("Adding player to room", username)
	r.Players[username] = conn
	r.TotalPlayers++
	r.OpponentType = "human"
	if r.TotalPlayers == PlayersNeeded {
		go r.ConverToPlaying()
	}
}

/////////////////////////////////////////////////////
//STARTS GAME AND INFORMS CLIENTS
/////////////////////////////////////////////////////

func (r *Room) StartGame() {
	mu.Lock()
	defer mu.Unlock()
	println("Starting game for room", r.ID)

	// Get player usernames for the response
	playerNames := make([]string, 0, len(r.Players))
	for username := range r.Players {
		playerNames = append(playerNames, username)
	}

	// Notify all players that the game has started
	for username, conn := range r.Players {
		if username == "bot" && r.OpponentType == "bot" {
			continue
		}
		err := conn.WriteJSON(types.SocketServerMessageType{
			Type: "game_started",
			Data: map[string]interface{}{
				"room_id":       r.ID,
				"status":        r.Status,
				"opponent_type": r.OpponentType,
				"current_turn":  r.CurrentTurn,
				"total_players": r.TotalPlayers,
				"players":       playerNames,
				"grid_data":     r.GridData,
			},
		})
		if err != nil {
			println("Error sending game started notification to", username, ":", err.Error())
		}
	}
}

/////////////////////////////////////////////////////
//REMOVE PLAYER FROM ROOM FUNCTION
/////////////////////////////////////////////////////

func (r *Room) RemovePlayer(username string) {
	mu.Lock()
	defer mu.Unlock()
	delete(r.Players, username)
	r.TotalPlayers--
	if r.TotalPlayers == 0 {
		r.deleteRoom()
	}

}

/////////////////////////////////////////////////////
//PICK WINNER AFTER PLAYER MISSING FROM ROOM
/////////////////////////////////////////////////////

func (r *Room) PickWinner() {
	println("Picking winner")
	mu.Lock()
	defer mu.Unlock()
	if r.TotalPlayers == 1 {
		for username := range r.Players {

			r.Winner = username
			println("Winner is", r.Winner)
			break
		}
	}

	r.deleteRoom()

}

////////////////////////////////////////////////////
//DELETE ROOM FUNCTION
//DELETES THE ROOM FROM THE WaitingRooms , PlayingRooms , roomIdToRoom
////////////////////////////////////////////////////

func (r *Room) deleteRoom() {

	mu.Lock()
	defer mu.Unlock()
	if r.Status == "waiting" {
		delete(roomManagerInstance.WaitingRooms, r.ID)
	} else {

		delete(roomManagerInstance.PlayingRooms, r.ID)
	}
	delete(roomManagerInstance.roomIdToRoom, r.ID)

}

////////////////////////////////////////////////////
// GET ROOM BY ID FUNCTION
// RETURNS THE ROOM BY ID
////////////////////////////////////////////////////

func GetRoomById(id string) *Room {
	mu.Lock()
	defer mu.Unlock()
	return roomManagerInstance.roomIdToRoom[id]
}

////////////////////////////////////////////////////
// CONVERT WAITING ROOM TO PLAYING ROOM
////////////////////////////////////////////////////

func (r *Room) ConverToPlaying() {
	mu.Lock()
	defer mu.Unlock()
	println("Converting room to playing")
	r.Status = "playing"
	roomManagerInstance.PlayingRooms[r.ID] = r
	delete(roomManagerInstance.WaitingRooms, r.ID)
	println("Starting playing game")
	go r.StartGame()
}
