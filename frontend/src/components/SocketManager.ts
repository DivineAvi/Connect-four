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
    public socketId: string | null = null;
    
    ////////////////////////////////////////////
    // Connect to the WebSocket server
    // @param url - The WebSocket server URL
    ////////////////////////////////////////////

    public connect(url: string): void {
        console.log(`Connecting to WebSocket at ${url}`);
        this.wsClient = new WebSocket(url);
        this.wsClient.onopen = () => {
            console.log("WebSocket connection established");
        };
    }
    
    ////////////////////////////////////////////
    // Disconnect from the WebSocket server
    ////////////////////////////////////////////

    public disconnect(): void {
        if (this.wsClient) {
            this.wsClient.close();
            this.wsClient = null;
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
        }
    }


}