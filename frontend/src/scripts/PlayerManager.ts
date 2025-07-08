import type { ColorDiscFunctionType, DiscColorType, OpponentType, RoomIdType} from "../types/GameTypes"

export class PlayerManager{
    public Username: string
    public Turn: Boolean
    public Opponent : OpponentType
    public OpponentUsername: string
    public DiscColor: DiscColorType
    public RoomId :RoomIdType
    public ColorDisc: ColorDiscFunctionType

    constructor(cb:ColorDiscFunctionType , DiscColor:DiscColorType, Opponent : OpponentType,RoomId:RoomIdType ,Username:string){
        this.Turn = false;
        this.ColorDisc = cb;
        this.DiscColor = DiscColor
        this.Opponent = Opponent;
        this.OpponentUsername = "";
        this.RoomId = RoomId;

        this.Username = Username;
    }

    public PlaceDisc(cIdx:number,rIdx:number){
        this.ColorDisc(cIdx,rIdx,this.DiscColor);
    }
}