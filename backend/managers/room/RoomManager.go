package room

import (
	"backend/managers/client"
	"backend/managers/types"
	"math/rand"
	"sync"
	"time"

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

// Initialize random seed for bot moves
func init() {
	rand.Seed(time.Now().UnixNano())
}

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
		GridData:     make([][]string, 7),
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
		Room.GridData[i] = make([]string, 6)
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

	println("Adding bot to room", r.ID)
	mu.Lock()
	r.OpponentType = "bot"
	r.TotalPlayers++
	r.Players["bot"] = nil
	mu.Unlock()
	if r.TotalPlayers == PlayersNeeded {
		println("Total players reached", r.TotalPlayers)
		r.ConverToPlaying()
	}
}

///////////////////////////////////////////////
// JOIN PLAYER TO ROOM
// JOIN A PLAYER TO THE ROOM
///////////////////////////////////////////////

////////////////////////////////////////////////
// ADDS A PLAYER TO THE ROOM
///////////////////////////////////////////////

func (r *Room) AddPlayer(username string, conn *websocket.Conn) {
	println("Adding player to room", username)
	mu.Lock()
	r.Players[username] = conn
	r.TotalPlayers++
	r.OpponentType = "human"
	mu.Unlock()
	if r.TotalPlayers == PlayersNeeded {
		r.ConverToPlaying()
	}
}

/////////////////////////////////////////////////////
//STARTS GAME AND INFORMS CLIENTS
/////////////////////////////////////////////////////

func (r *Room) StartGame() {

	println("Starting game for room", r.ID)

	// Get player usernames for the response
	playerNames := make([]string, 0, len(r.Players))
	for username := range r.Players {
		playerNames = append(playerNames, username)
	}

	// Assign colors to players
	var playerColor, botColor string

	// For bot games, assign colors differently
	if r.OpponentType == "bot" {
		// Find the human player
		var humanPlayer string
		for username := range r.Players {
			if username != "bot" {
				humanPlayer = username
				break
			}
		}

		// Assign colors - human player is red, bot is blue
		playerColor = "red"
		botColor = "blue"

		// Notify the human player
		conn := r.Players[humanPlayer]
		err := conn.WriteJSON(types.SocketServerMessageType{
			Type: "game_started",
			Data: map[string]interface{}{
				"room_id":           r.ID,
				"status":            r.Status,
				"opponent_type":     r.OpponentType,
				"current_turn":      r.CurrentTurn,
				"total_players":     r.TotalPlayers,
				"players":           playerNames,
				"grid_data":         r.GridData,
				"player_username":   humanPlayer,
				"player_color":      playerColor,
				"opponent_color":    botColor,
				"opponent_username": "bot",
			},
		})
		if err != nil {
			println("Error sending game started notification to", humanPlayer, ":", err.Error())
		}
	} else {
		// For human vs human games
		if len(playerNames) >= 2 {
			// First player is red, second is blue
			firstPlayerColor := "red"
			secondPlayerColor := "blue"

			// Notify all human players
			for username, conn := range r.Players {
				// Determine player color and opponent username
				var playerColor, opponentColor, opponentUsername string
				if username == playerNames[0] {
					playerColor = firstPlayerColor
					opponentColor = secondPlayerColor
					opponentUsername = playerNames[1]
				} else {
					playerColor = secondPlayerColor
					opponentColor = firstPlayerColor
					opponentUsername = playerNames[0]
				}

				err := conn.WriteJSON(types.SocketServerMessageType{
					Type: "game_started",
					Data: map[string]interface{}{
						"room_id":           r.ID,
						"status":            r.Status,
						"opponent_type":     r.OpponentType,
						"current_turn":      r.CurrentTurn,
						"total_players":     r.TotalPlayers,
						"players":           playerNames,
						"grid_data":         r.GridData,
						"player_username":   username,
						"player_color":      playerColor,
						"opponent_color":    opponentColor,
						"opponent_username": opponentUsername,
					},
				})
				if err != nil {
					println("Error sending game started notification to", username, ":", err.Error())
				}
			}
		}
	}

	// If bot is in the game and it's the bot's turn, make a move
	if r.OpponentType == "bot" && r.CurrentTurn == "bot" {
		go r.MakeBotMove()
	}
}

/////////////////////////////////////////////////////
// BOT MAKES A MOVE
/////////////////////////////////////////////////////

func (r *Room) MakeBotMove() {
	// Wait a bit to simulate "thinking"
	time.Sleep(1 * time.Second)

	// Check if the game is still active
	if r.Status != "playing" || r.CurrentTurn != "bot" {

		return
	}

	// Find a valid move
	column, row := r.findBotMove()

	// Apply the move

	// Double check game is still active
	if r.Status != "playing" || r.CurrentTurn != "bot" {

		return
	}

	// Bot always uses blue color
	botColor := "blue"

	// Update the grid
	r.GridData[column][row] = botColor

	// Change turn to the player
	for username := range r.Players {
		if username != "bot" {
			r.CurrentTurn = username
			break
		}
	}

	// Check for win condition
	winner := r.checkForWin(r.GridData, botColor)

	// Update game status if there's a winner
	if winner != "" {
		r.Status = "finished"
		r.Winner = "bot"
	}

	// Notify player about the update
	for username, conn := range r.Players {
		if username != "bot" {
			updateMsg := types.SocketServerMessageType{
				Type: "game_update",
				Data: map[string]interface{}{
					"room_id":      r.ID,
					"status":       r.Status,
					"current_turn": r.CurrentTurn,
					"grid_data":    r.GridData,
				},
			}

			if r.Status == "finished" {
				updateMsg.Data["winner"] = r.Winner
			}

			err := conn.WriteJSON(updateMsg)
			if err != nil {
				println("Error sending game update to", username, ":", err.Error())
			}
		}
	}
}

/////////////////////////////////////////////////////
// FIND A VALID MOVE FOR THE BOT
/////////////////////////////////////////////////////

func (r *Room) findBotMove() (int, int) {
	// First try to find a winning move
	for col := 0; col < len(r.GridData); col++ {
		row := r.getLowestEmptyRow(col)
		if row != -1 {
			// Try this move
			r.GridData[col][row] = "blue" // Bot is always blue
			if r.checkForWin(r.GridData, "blue") != "" {
				r.GridData[col][row] = "neutral" // Reset
				return col, row
			}
			r.GridData[col][row] = "neutral" // Reset
		}
	}

	// Then try to block player's winning move
	for col := 0; col < len(r.GridData); col++ {
		row := r.getLowestEmptyRow(col)
		if row != -1 {
			// Try this move for the player
			r.GridData[col][row] = "red" // Player is always red
			if r.checkForWin(r.GridData, "red") != "" {
				r.GridData[col][row] = "neutral" // Reset
				return col, row                  // Block this move
			}
			r.GridData[col][row] = "neutral" // Reset
		}
	}

	// Otherwise, make a random valid move
	validMoves := []struct {
		col int
		row int
	}{}

	for col := 0; col < len(r.GridData); col++ {
		row := r.getLowestEmptyRow(col)
		if row != -1 {
			validMoves = append(validMoves, struct {
				col int
				row int
			}{col, row})
		}
	}

	if len(validMoves) > 0 {
		randomMove := validMoves[rand.Intn(len(validMoves))]
		return randomMove.col, randomMove.row
	}

	// Fallback (should never happen in a valid game)
	return 0, 0
}

/////////////////////////////////////////////////////
// GET THE LOWEST EMPTY ROW IN A COLUMN
/////////////////////////////////////////////////////

func (r *Room) getLowestEmptyRow(col int) int {
	// Start from the bottom of the column and go up
	for row := len(r.GridData[col]) - 1; row >= 0; row-- {
		if r.GridData[col][row] == "neutral" {
			return row
		}
	}
	return -1 // Column is full
}

/////////////////////////////////////////////////////
// CHECK FOR WIN CONDITION
/////////////////////////////////////////////////////

func (r *Room) checkForWin(grid [][]string, color string) string {
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

/////////////////////////////////////////////////////
//REMOVE PLAYER FROM ROOM FUNCTION
/////////////////////////////////////////////////////

func (r *Room) RemovePlayer(username string) {

	if r.Status == "playing" {
		mu.Lock()
		delete(r.Players, username)
		r.TotalPlayers--
	}
	if r.Status == "waiting" {
		client.GetClientManager().RemovePlayingClient(username)
		mu.Lock()
		delete(r.Players, username)
		mu.Unlock()
		r.TotalPlayers--
	}

	if r.TotalPlayers == 0 {
		r.DeleteRoom()
	}

}

/////////////////////////////////////////////////////
//PICK WINNER AFTER PLAYER MISSING FROM ROOM
/////////////////////////////////////////////////////

func (r *Room) PickWinner() {
	println("Picking winner")

	if r.TotalPlayers == 1 {
		for username := range r.Players {

			r.Winner = username
			println("Winner is", r.Winner)
			break
		}
	}

	r.DeleteRoom()

}

////////////////////////////////////////////////////
//DELETE ROOM FUNCTION
//DELETES THE ROOM FROM THE WaitingRooms , PlayingRooms , roomIdToRoom
////////////////////////////////////////////////////

func (r *Room) DeleteRoom() {

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

	return roomManagerInstance.roomIdToRoom[id]
}

////////////////////////////////////////////////////
// CONVERT WAITING ROOM TO PLAYING ROOM
////////////////////////////////////////////////////

func (r *Room) ConverToPlaying() {
	println("Converting room to playing")
	mu.Lock()
	r.Status = "playing"
	roomManagerInstance.PlayingRooms[r.ID] = r
	delete(roomManagerInstance.WaitingRooms, r.ID)
	mu.Unlock()
	println("Starting playing game")
	r.StartGame()
}
