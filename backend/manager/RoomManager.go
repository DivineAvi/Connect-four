package manager

type Room struct {
	ID          string
	Players     map[string]bool // Maps player usernames to their presence in the room
	CurrentTurn string          // Username of the player whose turn it is
	GameState   [][]string      // 2D slice representing the game board
	CreatedAt   string          // Timestamp of when the room was created
}

type RoomManager struct {
	Rooms map[string]*Room
}

var roomManagerInstance *RoomManager

// GetRoomManager returns a singleton instance of RoomManager.
// It initializes the instance only once to ensure thread safety.
func GetRoomManager() *RoomManager {
	if roomManagerInstance == nil {
		roomManagerInstance = &RoomManager{
			Rooms: make(map[string]*Room),
		}
	}
	return roomManagerInstance
}
