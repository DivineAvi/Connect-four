package room

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Room struct {
	ID           string
	TotalPlayers int
	Players      map[string]*websocket.Conn // Maps player usernames to their presence in the room
	CurrentTurn  string                     // Username of the player whose turn it is
	GridData     [][]string                 // 2D slice representing the game board
	Status       string
	Winner       string
	Loser        string
	Draw         bool
}

type RoomManager struct {
	WaitingRooms map[string]*Room
	PlayingRooms map[string]*Room
}

var roomManagerInstance *RoomManager = nil

// GetRoomManager returns a singleton instance of RoomManager.
// It initializes the instance only once to ensure thread safety.
func GetRoomManager() *RoomManager {
	if roomManagerInstance == nil {
		roomManagerInstance = &RoomManager{
			WaitingRooms: make(map[string]*Room),
			PlayingRooms: make(map[string]*Room),
		}
	}
	return roomManagerInstance
}

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
	for i := range Room.GridData {
		Room.GridData[i] = make([]string, 7)
		for j := range Room.GridData[i] {
			Room.GridData[i][j] = "neutral"
		}
	}

	return Room
}
