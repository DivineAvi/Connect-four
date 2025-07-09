/////////////////////////////////////////////////////////////
//  SocketManager.ts
//  This file manages the WebSocket connection for the game.
/////////////////////////////////////////////////////////////

import type { SocketClientMessageType } from "../types/SocketMessageTypes";

export class SocketManager {
    ////////////////////
    // Variables
    ///////////////////
    public wsClient: WebSocket | null = null;
    public isConnected: Boolean = false;
    private reconnectAttempts: number = 0;
    private maxReconnectAttempts: number = 5;
    private reconnectTimeout: number | null = null;
    
    ////////////////////////////////////////////
    // Connect to the WebSocket server
    // @param url - The WebSocket server URL
    ////////////////////////////////////////////

    public async connect(url: string): Promise<void> {
        console.log(`Connecting to WebSocket at ${url}`);
        
        // Clear any existing connection
        if (this.wsClient) {
            this.disconnect();
        }
        
        return new Promise((resolve, reject) => {
            try {
                this.wsClient = new WebSocket(url);
                
                this.wsClient.onerror = (error) => {
                    console.error("WebSocket connection error:", error);
                    this.isConnected = false;
                    reject(new Error("Failed to connect to WebSocket server"));
                };
                
                this.wsClient.onopen = () => {
                    this.isConnected = true;
                    this.reconnectAttempts = 0; // Reset reconnect attempts on successful connection
                    console.log("WebSocket connection established");
                    resolve();
                };
                
                this.wsClient.onclose = (event) => {
                    this.isConnected = false;
                    console.log("WebSocket connection closed:", event.code, event.reason);
                    
                    // Don't auto-reconnect here - we'll handle reconnection in GameManager
                };
            } catch (error) {
                console.error("Error creating WebSocket:", error);
                this.isConnected = false;
                reject(error);
            }
        });
    }

    ////////////////////////////////////////////
    // Disconnect from the WebSocket server
    ////////////////////////////////////////////

    public disconnect(): void {
        // Clear any reconnection timeout
        if (this.reconnectTimeout !== null) {
            window.clearTimeout(this.reconnectTimeout);
            this.reconnectTimeout = null;
        }
        
        if (this.wsClient) {
            // Remove all event listeners to prevent memory leaks
            this.wsClient.onopen = null;
            this.wsClient.onclose = null;
            this.wsClient.onerror = null;
            this.wsClient.onmessage = null;
            
            // Close the connection
            if (this.wsClient.readyState === WebSocket.OPEN || 
                this.wsClient.readyState === WebSocket.CONNECTING) {
                this.wsClient.close();
            }
            
            this.wsClient = null;
            this.isConnected = false;
            console.log("WebSocket connection closed");
        } else {
            console.error("No WebSocket connection to close");
        }
    }

    ////////////////////////////////////////////
    // Send a message through the WebSocket
    // @param message - The message to send
    ////////////////////////////////////////////

    public sendMessage(message: SocketClientMessageType): void {
        if (this.wsClient && this.wsClient.readyState === WebSocket.OPEN) {
            this.wsClient.send(JSON.stringify(message));
        } else {
            console.error("WebSocket is not open. Unable to send message.");
            throw new Error("WebSocket is not open. Unable to send message.");
        }
    }
}