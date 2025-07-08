import { useEffect, useState } from "react"
import type { ColorDiscFunctionType, DiscColorType } from "../types/GameTypes";
import { GameManager } from "../scripts/GameManager";

export default function Room() {

    const [gridData, setGridData] = useState<Array<Array<DiscColorType>>>(
        Array.from({ length: 7 }, () => Array(6).fill("neutral" as DiscColorType))
    )
    const gameManager = GameManager.getInstance("ws://localhost:8080/ws");

    function colorDisc(colIdx: number, rowIdx: number, DiscColor: DiscColorType) {
        setGridData(prevGrid => {
            const newGrid = prevGrid.map(col => [...col]);
            newGrid[colIdx][rowIdx] = DiscColor; // or use a variable for player (e.g., 0/1)
            console.log("Disc is placed at")
            return newGrid;
        })
    }
    function PlaceDisc(cIdx:number,rIdx:number){
        
    }
    useEffect(() => {
        gameManager.ColorDiscFunction = colorDisc;
    }, [])
    return (
        <div className="w-full min-h-screen bg-black text-black flex flex-col text-white items-center justify-center p-2">
            <div>Opponent</div>
            <div
                className="grid grid-cols-7  w-fit"
            >
                {gridData.map((_, cIdx) => (
                    <div key={cIdx} id={`${cIdx + 1}`} className="w-full flex flex-col">
                        {
                            gridData[cIdx].map((_, rIdx) => {
                                const isPlacableTile = (rIdx < 5 && gridData[cIdx][rIdx + 1] != "neutral") && gridData[cIdx][rIdx] == "neutral" || (rIdx == 5 && gridData[cIdx][rIdx] == "neutral");
                                return (

                                    <div
                                        key={rIdx}
                                        onClick={() => {
                                            if (isPlacableTile) {
                                                PlaceDisc(cIdx, rIdx)
                                            }
                                        }}
                                        className={`${rIdx ? '' : ' border-t '} ${cIdx ? '' : ' border-l '} ${isPlacableTile ? ' bg-white/10 active:bg-blue-400/20 sm:hover:bg-blue-400/20 ' : ''}` + " w-12 aspect-square flex items-center justify-center border-b border-r border-white/40 p-1"}
                                    >
                                        {gridData[cIdx][rIdx] != "neutral" ?
                                            <div className={`${gridData[cIdx][rIdx] == "blue" ? ' bg-blue-400 ' : ' bg-red-400 '} rounded-full w-full h-full flex `}>

                                            </div>
                                            : ''}
                                    </div>

                                )
                            })
                        }
                    </div>

                ))}
            </div>
        </div>
    )
}