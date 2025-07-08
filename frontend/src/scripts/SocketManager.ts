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
    ////////////////////////////////////////////
    // Connect to the WebSocket server
    // @param url - The WebSocket server URL
    ////////////////////////////////////////////

    public async connect(url: string): Promise<void> {
        console.log(`Connecting to WebSocket at ${url}`);
        return new Promise((resolve, reject) => {
            this.wsClient = new WebSocket(url);
            this.wsClient.onerror = () => {
                reject();
            }
            this.wsClient.onopen = () => {
                this.isConnected = true
                console.log("WebSocket connection established");
                resolve();
            };
        })

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