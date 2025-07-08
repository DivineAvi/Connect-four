export type DiscColorType = "red" | "blue" | "neutral"
export type OpponentType ="human" | "bot"
export type ColorDiscFunctionType = (cIdx: number, rIdx: number, color: DiscColorType) => void
export type RoomIdType = string | null | undefined
