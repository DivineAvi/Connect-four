//////////////////////////////////////////////////////////////
//  GameManager.ts
//  This file manages the game state and WebSocket connection.
///////////////////////////////////////////////////////////////

import type { GameRejoinedMessageType, GameStartedServerMessageType, GameUpdateServerMessageType, NewGameServerMessageType, PlayerDisconnectedMessageType, PlayerRejoinedMessageType, SocketClientMessageType, SocketServerMessageType } from "../types/SocketMessageTypes";
import { SocketManager } from "./SocketManager";
import { PlayerManager } from "./PlayerManager";
import type { ColorDiscFunctionType, DiscColorType, OpponentType, RoomIdType } from "../types/GameTypes";

export class GameManager {
    ///////////////////////////////
    // Variables
    ///////////////////////////////

    public socketManager: SocketManager;
    public wsUrl: string | null
    private static instance: GameManager | null = null
    public hasGameStarted: boolean = false
    public Player: PlayerManager | null = null
    public ColorDiscFunction: ColorDiscFunctionType | null = null;
    public SetGameStarted: (value: boolean) => void = () => { }
    public SetGridData: (data: string[][]) => void = () => { }
    public SetCurrentTurn: (value: boolean) => void = () => { }
    public SetStatusMessage: (message: string) => void = () => { }
    public SetReconnecting: (value: boolean) => void = () => { }
    public SetCountdown: (value: number | undefined) => void = () => { }
    
    // Store game state for reconnection
    private lastKnownRoomId: string | null = null;
    private lastKnownUsername: string | null = null;
    private reconnectionTimer: number | null = null;
    private countdownInterval: ReturnType<typeof setInterval> | null = null;

    ///////////////////////////////////////
    // Singleton pattern to ensure only one instance of GameManager exists
    ///////////////////////////////////////

    public static getInstance(wsUrl: string | null = null): GameManager {
        if (GameManager.instance === null) {
            GameManager.instance = new GameManager(wsUrl);
        }
        return GameManager.instance;
    }

    ///////////////////////////////////////
    // Constructor
    // @param wsUrl - The WebSocket server Url
    ///////////////////////////////////////

    constructor(wsUrl: string | null = null) {
        this.wsUrl = wsUrl;
        this.socketManager = new SocketManager();
        
        // Try to load saved game state from localStorage
        this.loadGameState();
        
        // Setup reconnection handling
        window.addEventListener('online', this.handleReconnection.bind(this));
    }
    
    ///////////////////////////////////////
    // Save game state to localStorage
    ///////////////////////////////////////
    private saveGameState(): void {
        if (this.Player) {
            const gameState = {
                roomId: this.Player.RoomId as string,
                username: this.Player.Username
            };
            localStorage.setItem('connect4GameState', JSON.stringify(gameState));
            
            // Also store in memory for immediate access
            this.lastKnownRoomId = this.Player.RoomId as string;
            this.lastKnownUsername = this.Player.Username;
        }
    }
    
    ///////////////////////////////////////
    // Load game state from localStorage
    ///////////////////////////////////////
    private loadGameState(): void {
        const savedState = localStorage.getItem('connect4GameState');
        if (savedState) {
            try {
                const gameState = JSON.parse(savedState);
                this.lastKnownRoomId = gameState.roomId;
                this.lastKnownUsername = gameState.username;
            } catch (e) {
                console.error('Failed to parse saved game state', e);
            }
        }
    }
    
    ///////////////////////////////////////
    // Clear saved game state
    ///////////////////////////////////////
    public clearGameState(): void {
        localStorage.removeItem('connect4GameState');
        this.lastKnownRoomId = null;
        this.lastKnownUsername = null;
    }

    public SetUpPlayer(ColorDiscFunction: ColorDiscFunctionType, DiscColor: DiscColorType, Opponent: OpponentType, RoomId: RoomIdType, Username: string) {
        console.log("Setting up player", ColorDiscFunction, DiscColor, Opponent, RoomId, Username)
        this.Player = new PlayerManager(ColorDiscFunction, DiscColor, Opponent, RoomId, Username)
        
        // Save game state for potential reconnection
        this.saveGameState();
    }

    ///////////////////////////////////////
    // Place Disc
    // This method sends a message to the server to place a disc
    ///////////////////////////////////////
    public place_disc(colIdx: number, rowIdx: number) {
        if (!this.Player || !this.Player.Turn) {
            console.log("Not your turn or player not initialized");
            return;
        }
        
        console.log("Placing disc at", colIdx, rowIdx);
        this.socketManager.sendMessage({
            type: "game_update",
            username: this.Player.Username,
            data: {
                "action": "place_disc",
                "column": colIdx,
                "row": rowIdx,
                "room_id": this.Player.RoomId,
                "player_color": this.Player.DiscColor
            }
        });
        
        // Update local state
        this.Player.PlaceDisc(colIdx, rowIdx);
        this.Player.Turn = false;
    }

    ///////////////////////////////////////
    // Create a new game
    // This method sends a message to the server to create a new game
    ///////////////////////////////////////

    public async new_game_request_handler(username: string) {
        console.log("new_game_request_handler", username)
        if (!this.socketManager.isConnected) {
            await this.socketManager.connect(this.wsUrl + "?username=" + encodeURIComponent(username));
            this.listen_server_for_messages();
        }
        if (this.socketManager.isConnected) {
            console.log("Requesting new game...");
            this.socketManager.sendMessage({
                type: "new_game",
                username: username,
                data: {}
            } as SocketClientMessageType);
        }
    }
    
    ///////////////////////////////////////
    // Handle reconnection when connection is lost
    ///////////////////////////////////////
    public handleReconnection(): void {
        // Clear any existing timers
        if (this.reconnectionTimer !== null) {
            window.clearTimeout(this.reconnectionTimer);
        }
        if (this.countdownInterval !== null) {
            window.clearInterval(this.countdownInterval);
            this.SetCountdown(undefined);
        }
        
        // If we have a saved game state, try to reconnect
        if (this.lastKnownRoomId && this.lastKnownUsername) {
            this.SetReconnecting(true);
            this.SetStatusMessage("Attempting to reconnect...");
            
            // Try to reconnect
            this.reconnectToGame(this.lastKnownUsername, this.lastKnownRoomId);
        }
    }
    
    ///////////////////////////////////////
    // Reconnect to an existing game
    ///////////////////////////////////////
    public async reconnectToGame(username: string, roomId: string): Promise<void> {
        try {
            // Connect to the websocket
            if (!this.socketManager.isConnected) {
                await this.socketManager.connect(this.wsUrl + "?username=" + encodeURIComponent(username));
                this.listen_server_for_messages();
            }
            
            if (this.socketManager.isConnected) {
                console.log("Attempting to reconnect to game...");
                
                // Send reconnect message
                this.socketManager.sendMessage({
                    type: "reconnect",
                    username: username,
                    data: {
                        room_id: roomId
                    }
                } as SocketClientMessageType);
            }
        } catch (error) {
            console.error("Failed to reconnect:", error);
            this.SetStatusMessage("Failed to reconnect. Trying again in 5 seconds...");
            
            // Try again in 5 seconds
            this.reconnectionTimer = window.setTimeout(() => {
                this.reconnectToGame(username, roomId);
            }, 5000);
        }
    }
    
    ///////////////////////////////////////
    // Handle player disconnection
    ///////////////////////////////////////
    public player_disconnected_handler(message: PlayerDisconnectedMessageType): void {
        console.log("Player disconnected:", message);
        
        // Show message
        this.SetStatusMessage(message.data.message);
        
        // Start countdown timer
        let countdown = 30;
        this.SetCountdown(countdown);
        
        this.countdownInterval = window.setInterval(() => {
            countdown--;
            this.SetCountdown(countdown);
            
            if (countdown <= 0) {
                if (this.countdownInterval !== null) {
                    window.clearInterval(this.countdownInterval);
                    this.countdownInterval = null;
                }
                this.SetCountdown(undefined);
            }
        }, 1000);
    }
    
    ///////////////////////////////////////
    // Handle player rejoining
    ///////////////////////////////////////
    public player_rejoined_handler(message: PlayerRejoinedMessageType): void {
        console.log("Player rejoined:", message);
        
        // Clear countdown
        if (this.countdownInterval !== null) {
            window.clearInterval(this.countdownInterval);
            this.countdownInterval = null;
            this.SetCountdown(undefined);
        }
        
        // Show message
        this.SetStatusMessage(`${message.data.username} has rejoined the game!`);
        
        // Clear message after 5 seconds
        setTimeout(() => {
            this.SetStatusMessage("");
        }, 5000);
    }
    
    ///////////////////////////////////////
    // Handle game rejoined
    ///////////////////////////////////////
    public game_rejoined_handler(message: GameRejoinedMessageType): void {
        console.log("Game rejoined:", message);
        
        this.SetReconnecting(false);
        this.SetGameStarted(true);
        
        // Set up player
        if (this.ColorDiscFunction) {
            this.SetUpPlayer(
                this.ColorDiscFunction,
                message.data.player_color,
                message.data.opponent_type,
                message.data.room_id,
                message.data.player_username
            );
            
            if (this.Player) {
                this.Player.Opponent = message.data.opponent_type;
                this.Player.OpponentUsername = message.data.opponent_username;
                this.Player.RoomId = message.data.room_id;
                this.Player.Username = message.data.player_username;
                this.Player.Turn = message.data.current_turn === message.data.player_username;
                this.SetGridData(message.data.grid_data);
                this.SetCurrentTurn(this.Player.Turn as boolean);
            }
        }
        
        // Show message
        this.SetStatusMessage("Successfully reconnected to the game!");
        
        // Clear message after 5 seconds
        setTimeout(() => {
            this.SetStatusMessage("");
        }, 5000);
    }
    
    ///////////////////////////////////////
    // Handle new game response
    // This method handles the response from the server when a new game is created
    ///////////////////////////////////////

    public new_game_response_handler(message: NewGameServerMessageType) {
        console.log("Room ID : ", message.data.room_id)
    }
    
    ////////////////////////////////////////
    // Game started handler
    // This method handles the response from the server when a game is started
    ////////////////////////////////////////

    public game_started_handler(message: GameStartedServerMessageType) {
        console.log("Game started", message)
        this.SetGameStarted(true)
        setTimeout(() => {
        if (this.ColorDiscFunction) {
            this.SetUpPlayer(this.ColorDiscFunction, message.data.player_color, message.data.opponent_type, message.data.room_id, message.data.player_username)
            if (this.Player) {
                this.Player.Opponent = message.data.opponent_type
                this.Player.OpponentUsername = message.data.opponent_username
                this.Player.RoomId = message.data.room_id
                this.Player.Username = message.data.player_username
                this.Player.Turn = message.data.current_turn == message.data.player_username
                this.SetGridData(message.data.grid_data)
                this.SetCurrentTurn(this.Player.Turn as boolean)
            }
        }
    }, 1000)
        console.log("Player", this.Player)
    }

    ////////////////////////////////////////
    // Game update handler
    // This method handles updates from the server about the game state
    ////////////////////////////////////////
    public game_update_handler(message: GameUpdateServerMessageType) {
        console.log("Game update received", message);
        
        if (this.Player) {
            // Update grid data
            this.SetGridData(message.data.grid_data);
            
            // Update turn
            const isMyTurn = message.data.current_turn === this.Player.Username;
            this.Player.Turn = isMyTurn;
            this.SetCurrentTurn(isMyTurn);
            
            // If game is over
            if (message.data.status === "finished") {
                // Show message if provided
                if (message.data.message) {
                    alert(message.data.message);
                } else {
                    alert(message.data.winner === this.Player.Username ? "You won!" : "You lost!");
                }
                
                // Clear saved game state
                this.clearGameState();
            }
        }
    }

    ////////////////////////////////////////
    // Game over handler
    // This method handles the game over message from the server
    ////////////////////////////////////////
    public game_over_handler(message: SocketServerMessageType) {
        console.log("Game over", message);
        if (this.Player) {
            const winner = message.data.winner as string;
            if (winner === this.Player.Username) {
                alert("You won!");
            } else if (winner === "draw") {
                alert("It's a draw!");
            } else {
                alert("You lost!");
            }
            
            // Clear saved game state
            this.clearGameState();
        }
    }

    ///////////////////////////////////////
    // Listen for messages from the server
    // This method sets up the WebSocket to listen for messages from the server
    ///////////////////////////////////////

    public listen_server_for_messages() {
        if (this.socketManager.wsClient) {
            this.socketManager.wsClient.onmessage = (event) => {
                const message = JSON.parse(event.data);

                switch (message.type) {
                    case "new_game_response":
                        this.new_game_response_handler(message);
                        break;
                    case "game_started":
                        this.game_started_handler(message);
                        break;
                    case "join_game":
                        break;
                    case "game_update":
                        this.game_update_handler(message);
                        break;
                    case "game_over":
                        this.game_over_handler(message);
                        break;
                    case "player_disconnected":
                        this.player_disconnected_handler(message);
                        break;
                    case "player_rejoined":
                        this.player_rejoined_handler(message);
                        break;
                    case "game_rejoined":
                        this.game_rejoined_handler(message);
                        break;
                    case "error":
                        console.error("Error:", message.data.error);
                        alert("Error: " + message.data.error);
                        break;
                    case "info":
                        console.log("Info:", message.data.info);
                        break;
                    case "connection_ack":
                        break;
                    default:
                        console.warn("Unknown message type:", message.type);
                }
            };
            
            // Handle connection close
            this.socketManager.wsClient.onclose = () => {
                console.log("WebSocket connection closed");
                
                // If we have game state, start reconnection process
                if (this.Player && this.Player.RoomId && this.Player.Username) {
                    this.SetStatusMessage("Connection lost. Attempting to reconnect...");
                    this.SetReconnecting(true);
                    
                    // Try to reconnect after a short delay
                    this.reconnectionTimer = window.setTimeout(() => {
                        if (this.Player && this.Player.RoomId && this.Player.Username) {
                            this.reconnectToGame(this.Player.Username, this.Player.RoomId as string);
                        }
                    }, 2000);
                }
            };
        }
    }
}