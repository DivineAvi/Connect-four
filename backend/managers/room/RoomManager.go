package room

import (
	"backend/managers/client"
	"backend/managers/types"
	"log"
	"math/rand"
	"sync"
	"time"

	"backend/db"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

////////////////////////////////////
//   STRUCTURE AND VARIABLES
////////////////////////////////////

type Room struct {
	ID                  string
	OpponentType        string
	TotalPlayers        int
	Players             map[string]*websocket.Conn // Maps player usernames to their presence in the room
	DisconnectedPlayers map[string]time.Time       // Maps disconnected player usernames to their disconnect time
	CurrentTurn         string                     // Username of the player whose turn it is
	GridData            [][]string                 // 2D slice representing the game board
	Status              string                     // waiting, playing, finished
	Winner              string
	Loser               string
	Draw                bool
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
		ID:                  RoomId,
		GridData:            make([][]string, 7),
		Players:             make(map[string]*websocket.Conn),
		DisconnectedPlayers: make(map[string]time.Time),
		Status:              "waiting",
		CurrentTurn:         username,
		TotalPlayers:        1,
		Winner:              "",
		Loser:               "",
		Draw:                false,
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

func (r *Room) JoinPlayer(username string, conn *websocket.Conn) {
	println("Player rejoining room:", username)

	mu.Lock()
	defer mu.Unlock()

	if disconnectTime, exists := r.DisconnectedPlayers[username]; exists {
		if time.Since(disconnectTime) <= 30*time.Second {
			delete(r.DisconnectedPlayers, username)

			r.Players[username] = conn

			playerNames := make([]string, 0, len(r.Players))
			for playerName := range r.Players {
				playerNames = append(playerNames, playerName)
			}

			var playerColor, opponentColor, opponentUsername string
			if len(playerNames) >= 2 {
				if r.OpponentType == "bot" {
					playerColor = "red"
					opponentColor = "blue"
					opponentUsername = "bot"
				} else {
					if username == playerNames[0] {
						playerColor = "red"
						opponentColor = "blue"
						opponentUsername = playerNames[1]
					} else {
						playerColor = "blue"
						opponentColor = "red"
						opponentUsername = playerNames[0]
					}
				}
			}

			conn.WriteJSON(types.SocketServerMessageType{
				Type: "game_rejoined",
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

			for playerName, playerConn := range r.Players {
				if playerName != username && playerName != "bot" {
					playerConn.WriteJSON(types.SocketServerMessageType{
						Type: "player_rejoined",
						Data: map[string]interface{}{
							"username": username,
						},
					})
				}
			}

			println("Player successfully rejoined:", username)
			return
		} else {
			println("Rejoin time expired for player:", username)

			delete(r.DisconnectedPlayers, username)

			for playerName := range r.Players {
				if playerName != "bot" {
					r.Winner = playerName
					r.Status = "finished"

					r.Players[playerName].WriteJSON(types.SocketServerMessageType{
						Type: "game_update",
						Data: map[string]interface{}{
							"room_id": r.ID,
							"status":  "finished",
							"winner":  playerName,
							"message": "Opponent failed to reconnect in time",
						},
					})
					break
				}
			}

			conn.WriteJSON(types.SocketServerMessageType{
				Type: "error",
				Data: map[string]interface{}{
					"error": "You failed to reconnect within the time limit. The game is over.",
				},
			})

			go func() {
				time.Sleep(5 * time.Second)
				for playerName := range r.Players {
					client.GetClientManager().RemovePlayingClient(playerName)
				}
				r.DeleteRoom()
			}()

			return
		}
	}

	r.Players[username] = conn
	r.TotalPlayers++

	playerNames := make([]string, 0, len(r.Players))
	for playerName := range r.Players {
		playerNames = append(playerNames, playerName)
	}

	conn.WriteJSON(types.SocketServerMessageType{
		Type: "game_joined",
		Data: map[string]interface{}{
			"room_id":       r.ID,
			"status":        r.Status,
			"current_turn":  r.CurrentTurn,
			"total_players": r.TotalPlayers,
			"players":       playerNames,
			"grid_data":     r.GridData,
		},
	})
}

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

	playerNames := make([]string, 0, len(r.Players))
	for username := range r.Players {
		playerNames = append(playerNames, username)
	}

	var playerColor, botColor string

	if r.OpponentType == "bot" {
		var humanPlayer string
		for username := range r.Players {
			if username != "bot" {
				humanPlayer = username
				break
			}
		}

		playerColor = "red"
		botColor = "blue"

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
		if len(playerNames) >= 2 {
			firstPlayerColor := "red"
			secondPlayerColor := "blue"

			for username, conn := range r.Players {
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

	if r.OpponentType == "bot" && r.CurrentTurn == "bot" {
		go r.MakeBotMove()
	}
}

/////////////////////////////////////////////////////
// BOT MAKES A MOVE
/////////////////////////////////////////////////////

func (r *Room) MakeBotMove() {
	time.Sleep(1 * time.Second)

	if r.Status != "playing" || r.CurrentTurn != "bot" {

		return
	}

	column, row := r.findBotMove()

	if r.Status != "playing" || r.CurrentTurn != "bot" {

		return
	}

	botColor := "blue"

	r.GridData[column][row] = botColor

	for username := range r.Players {
		if username != "bot" {
			r.CurrentTurn = username
			break
		}
	}

	winner := r.checkForWin(r.GridData, botColor)

	if winner != "" {
		r.Status = "finished"
		r.Winner = "bot"
	}

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
			r.GridData[col][row] = "blue"
			if r.checkForWin(r.GridData, "blue") != "" {
				r.GridData[col][row] = "neutral"
				return col, row
			}
			r.GridData[col][row] = "neutral"
		}
	}

	// Then try to block player's winning move
	for col := 0; col < len(r.GridData); col++ {
		row := r.getLowestEmptyRow(col)
		if row != -1 {
			r.GridData[col][row] = "red"
			if r.checkForWin(r.GridData, "red") != "" {
				r.GridData[col][row] = "neutral"
				return col, row
			}
			r.GridData[col][row] = "neutral"
		}
	}

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

	return 0, 0
}

/////////////////////////////////////////////////////
// GET THE LOWEST EMPTY ROW IN A COLUMN
/////////////////////////////////////////////////////

func (r *Room) getLowestEmptyRow(col int) int {
	for row := len(r.GridData[col]) - 1; row >= 0; row-- {
		if r.GridData[col][row] == "neutral" {
			return row
		}
	}
	return -1
}

/////////////////////////////////////////////////////
// CHECK FOR WIN CONDITION
/////////////////////////////////////////////////////

func (r *Room) checkForWin(grid [][]string, color string) string {

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

/////////////////////////////////////////////////////
//REMOVE PLAYER FROM ROOM FUNCTION AND NOTIFIES EVERYONE
/////////////////////////////////////////////////////

func (r *Room) DisconnectPlayer(username string) {
	println("Disconnecting player from room:", username)

	if r.Status == "playing" {
		mu.Lock()

		r.DisconnectedPlayers[username] = time.Now()

		players := make(map[string]*websocket.Conn)
		for playerName, conn := range r.Players {
			players[playerName] = conn
		}

		mu.Unlock()

		// Notify remaining players about the disconnection
		for playerName, conn := range players {
			println("Notifying player ", playerName, " about disconnection of ", username)
			if playerName != "bot" && playerName != username {
				conn.WriteJSON(types.SocketServerMessageType{
					Type: "player_disconnected",
					Data: map[string]interface{}{
						"username": username,
						"message":  "Player disconnected. They have 30 seconds to reconnect.",
					},
				})
			}
		}
		println("Players Notified about disconnection of ", username)
		// Start a timer to check if the player reconnects within 30 seconds
		go func(disconnectedUsername string) {
			time.Sleep(30 * time.Second)

			mu.Lock()
			defer mu.Unlock()

			// Check if the player is still disconnected
			if _, stillDisconnected := r.DisconnectedPlayers[disconnectedUsername]; stillDisconnected {

				r.PickWinner()

			}
		}(username)

	} else if r.Status == "waiting" {
		client.GetClientManager().RemovePlayingClient(username)
		mu.Lock()
		r.DeleteRoom()
		mu.Unlock()
	}

}

/////////////////////////////////////////////////////
//PICK WINNER AFTER PLAYER MISSING FROM ROOM
/////////////////////////////////////////////////////

func (r *Room) PickWinner() {
	println("Picking winner")
	if len(r.DisconnectedPlayers) == 2 {
		println("Both players disconnected")
		r.DeleteRoom()
		return
	}
	if r.Status == "finished" {
		return
	}

	for username := range r.Players {
		if _, exists := r.DisconnectedPlayers[username]; !exists {
			r.Winner = username
			r.Status = "finished"
			println("Winner is", r.Winner)
			break
		}
	}
	if r.Winner != "bot" {
		r.UpdatePlayerStats(r.Winner)

		playerConn := r.Players[r.Winner]
		playerName := r.Winner
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

///////////////////////////////////////////
//UPDATE PLAYER STATS FUNCTION
//UPDATES THE PLAYER STATS IN THE DATABASE
///////////////////////////////////////////

func (r *Room) UpdatePlayerStats(winner string) {

	println("Updating player stats for winner:", winner)
	dbInstance, err := db.NewDB()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return
	}
	defer dbInstance.Close()

	if winner != "" {
		var loser string
		for playerName := range r.Players {
			println("Player name:", playerName)
			if playerName != winner && playerName != "bot" {
				loser = playerName
				break
			}
		}

		if loser == "" {
			for playerName := range r.DisconnectedPlayers {
				if playerName != winner && playerName != "bot" {
					loser = playerName
					break
				}
			}
		}

		println("Winner:", winner, "Loser:", loser)

		if loser != "" && winner != "bot" {
			println("Updating database with winner:", winner, "and loser:", loser)

			_, err := dbInstance.CreateOrUpdatePlayer(winner)
			if err != nil {
				log.Printf("Failed to create/update winner entry: %v", err)
			}

			_, err = dbInstance.CreateOrUpdatePlayer(loser)
			if err != nil {
				log.Printf("Failed to create/update loser entry: %v", err)
			}

			err = dbInstance.UpdateGameResult(winner, loser)
			if err != nil {
				log.Printf("Failed to update game result: %v", err)
			} else {
				println("Successfully updated game result in database")

				winnerPlayer, err := dbInstance.GetPlayerByUsername(winner)
				if err != nil {
					log.Printf("Failed to retrieve winner data: %v", err)
				} else {
					println("Winner stats - Wins:", winnerPlayer.Wins, "Losses:", winnerPlayer.Losses, "Rating:", winnerPlayer.Rating)
				}

				loserPlayer, err := dbInstance.GetPlayerByUsername(loser)
				if err != nil {
					log.Printf("Failed to retrieve loser data: %v", err)
				} else {
					println("Loser stats - Wins:", loserPlayer.Wins, "Losses:", loserPlayer.Losses, "Rating:", loserPlayer.Rating)
				}
			}
		} else {
			println("Cannot update stats: invalid winner or loser")
			if winner == "bot" {
				println("Winner is a bot, not updating stats")
			}
			if loser == "" {
				println("No loser found to update stats")
			}
		}
	} else {
		var humanPlayers []string
		for playerName := range r.Players {
			if playerName != "bot" {
				humanPlayers = append(humanPlayers, playerName)
			}
		}

		println("Draw detected with", len(humanPlayers), "human players")

		if len(humanPlayers) >= 2 {
			println("Updating draw result for", humanPlayers[0], "and", humanPlayers[1])

			_, err := dbInstance.CreateOrUpdatePlayer(humanPlayers[0])
			if err != nil {
				log.Printf("Failed to create/update first player entry: %v", err)
			}

			_, err = dbInstance.CreateOrUpdatePlayer(humanPlayers[1])
			if err != nil {
				log.Printf("Failed to create/update second player entry: %v", err)
			}

			err = dbInstance.UpdateDraw(humanPlayers[0], humanPlayers[1])
			if err != nil {
				log.Printf("Failed to update draw result: %v", err)
			} else {
				println("Successfully updated draw result in database")
			}
		} else {
			println("Not enough human players for a draw update")
		}
	}
}
