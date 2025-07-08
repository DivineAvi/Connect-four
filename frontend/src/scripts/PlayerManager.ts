import type { ColorDiscFunctionType, DiscColorType, OpponentType, RoomIdType} from "../types/GameTypes"

export class PlayerManager{
    public Turn: Boolean
    public Opponent : OpponentType
    public DiscColor: DiscColorType
    public RoomId :RoomIdType
    public ColorDisc: ColorDiscFunctionType

    constructor(cb:ColorDiscFunctionType , DiscColor:DiscColorType, Opponent : OpponentType,RoomId:RoomIdType){
        this.Turn = false;
        this.ColorDisc = cb;
        this.DiscColor = DiscColor
        this.Opponent = Opponent;
        this.RoomId = RoomId;
    }

    public PlaceDisc(cIdx:number,rIdx:number){
        this.ColorDisc(cIdx,rIdx,this.DiscColor);
    }
}