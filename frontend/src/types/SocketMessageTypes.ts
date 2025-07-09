import type { DiscColorType, OpponentType } from "./GameTypes";
export interface SocketClientMessageType {
    type: "new_game" | "join_game" | "game_update" | "game_over" | "connection_ack" | "reconnect";
    username: string;
    data: any;
}
export interface SocketServerMessageType {
    type: "new_game" | "join_game" | "game_update" | "game_over" | "connection_ack" | "error" | "info" | "player_disconnected" | "player_rejoined" | "game_rejoined";
    success?: Boolean
    room_id?: string | null;
    socket_id?: string | null;
    player_id?: string | null;
    data?: any;
}

export interface NewGameServerMessageType {
    type: "new_game_response"
    data: {
        room_id: string;
        status: string;
        current_turn: string;
        total_players: number;
        players: string[];
        grid_data: string[][];
    }
}

export interface GameStartedServerMessageType {
    type: "game_started"
    data: {
        player_username: string;
        player_color: DiscColorType;
        opponent_color : DiscColorType;
        opponent_username: string;
        opponent_type: OpponentType;
        room_id: string;
        status: string;
        current_turn: string;
        total_players: number;
        players: string[];
        grid_data: string[][];
    }
}

export interface GameUpdateServerMessageType {
    type: "game_update"
    data: {
        room_id: string;
        status: string;
        current_turn: string;
        grid_data: string[][];
        winner?: string;
        message?: string;
    }
}

export interface PlayerDisconnectedMessageType {
    type: "player_disconnected"
    data: {
        username: string;
        message: string;
    }
}

export interface PlayerRejoinedMessageType {
    type: "player_rejoined"
    data: {
        username: string;
    }
}

export interface GameRejoinedMessageType {
    type: "game_rejoined"
    data: {
        player_username: string;
        player_color: DiscColorType;
        opponent_color: DiscColorType;
        opponent_username: string;
        opponent_type: OpponentType;
        room_id: string;
        status: string;
        current_turn: string;
        total_players: number;
        players: string[];
        grid_data: string[][];
    }
}