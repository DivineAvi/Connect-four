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
    success: Boolean
    Opponent: string
    OpponentType : OpponentType
}