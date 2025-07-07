//////////////////////////////////////////////////////////////
//  GameManager.ts
//  This file manages the game state and WebSocket connection.
///////////////////////////////////////////////////////////////

import type { SocketClientMessageType, SocketServerMessageType } from "../types/SocketMessageTypes";
import { SocketManager } from "./SocketManager";

export class GameManager {
    ///////////////////////////////
    // Variables
    ///////////////////////////////

    public roomId: string | null | undefined = null;
    public socketManager: SocketManager;
    public wsUrl: string | null = null;
    private static instance: GameManager | null = null;
    public hasGameStarted: boolean = false;

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
        this.roomId = null;
    }

    ///////////////////////////////////////
    // Create a new game
    // This method sends a message to the server to create a new game
    ///////////////////////////////////////

    public async new_game_request_handler(username: string) {

        if (!this.socketManager.isConnected) {
            await this.socketManager.connect(this.wsUrl + "?username=" + encodeURIComponent(username));
            this.listen_server_for_messages();
        }
        if (this.socketManager.isConnected) {
            console.log("Requesting new game...");

            this.socketManager.sendMessage({
                type: "new_game",
                socket_id: this.socketManager.socketId,
            } as SocketClientMessageType);
        }
    }
    ///////////////////////////////////////
    // Handle new game response
    // This method handles the response from the server when a new game is created
    ///////////////////////////////////////

    public new_game_response_handler(message: SocketServerMessageType) {
        this.roomId = message.room_id || null;
        if (this.roomId) {
            this.hasGameStarted = true;
            localStorage.setItem("room_id", this.roomId);
        }
    }

    ///////////////////////////////////////
    // Listen for messages from the server
    // This method sets up the WebSocket to listen for messages from the server
    ///////////////////////////////////////

    public listen_server_for_messages() {
        if (this.socketManager.wsClient) {
            this.socketManager.wsClient.onmessage = (event) => {
                const message: SocketServerMessageType = JSON.parse(event.data);
                console.log("Received message from server:", message);

                switch (message.type) {
                    case "new_game":
                        this.new_game_response_handler(message);
                        break;
                    case "join_game":
                        break;
                    case "game_update":
                        break;
                    case "game_over":

                        break;
                    case "connection_ack":
                        break;
                    default:
                        console.warn("Unknown message type:", message.type);
                }
            };
        }
    }

}