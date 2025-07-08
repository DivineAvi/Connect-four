import type { OpponentType } from "./GameTypes";
export interface SocketClientMessageType {
    type: "new_game" | "join_game" | "game_update" | "game_over" | "connection_ack";
    room_id?: string | null;
    socket_id?: string | null;
    player_id?: string | null;
}
export interface SocketServerMessageType {
    type: "new_game" | "join_game" | "game_update" | "game_over" | "connection_ack";
    success?: Boolean
    room_id?: string | null;
    socket_id?: string | null;
    player_id?: string | null;
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