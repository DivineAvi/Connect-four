import { useEffect, useState } from "react";
import Leaderboard from "../components/Leaderboard";

interface LobbyPropsType {
    roomId: string | null;
    handleNewGame: () => void;
    handleReconnectGame: () => void;
    searching: boolean;
    hasSavedGame: boolean;
}

export default function Lobby(props: LobbyPropsType) {
    const [savedUsername, setSavedUsername] = useState<string>("");
    const [rejoinLoading, setRejoinLoading] = useState<boolean>(false);
    const [showLeaderboard, setShowLeaderboard] = useState<boolean>(false);
    
    useEffect(() => {
        const savedState = localStorage.getItem('connect4GameState');
        if (savedState) {
            try {
                const gameState = JSON.parse(savedState);
                if (gameState.roomId && gameState.username) {
                    setSavedUsername(gameState.username);
                    
                    const usernameInput = document.querySelector("input[name='username']") as HTMLInputElement;
                    if (usernameInput) {
                        usernameInput.value = gameState.username;
                    }
                }
            } catch (e) {
                console.error('Failed to parse saved game state', e);
            }
        }
    }, []);
    
    const handleRejoinGame = async () => {
        setRejoinLoading(true);
        try {
            await props.handleReconnectGame();
        } catch (error) {
            setRejoinLoading(false);
        }
    };

    const toggleLeaderboard = () => {
        setShowLeaderboard(!showLeaderboard);
    };
    
    return (
        <div className="w-full min-h-screen bg-black text-white flex flex-col">
 
            <div className="w-full bg-black border-b border-blue-500/30 p-4 flex justify-between items-center relative">
                <h1 className="text-2xl font-bold text-blue-400">Connect Four</h1>
                <button 
                    onClick={toggleLeaderboard}
                    className="bg-blue-500 hover:bg-blue-600 text-white py-2 px-4 rounded-lg transition-colors"
                >
                    {showLeaderboard ? 'Hide Leaderboard' : 'Show Leaderboard'}
                </button>
            </div>
            <Leaderboard isOpen={showLeaderboard} onClose={toggleLeaderboard} />
            
            <div className="flex-1 flex items-center justify-center p-2">
                <div className="flex flex-col gap-4 w-96 border-2 border-white/9 p-5 rounded-xl">
                    <h1 className="text-2xl font-bold text-center mb-4">Connect Four</h1>
             
                    {props.hasSavedGame && (
                        <div className="bg-blue-500/20 p-4 rounded-lg mb-4">
                            <p className="text-center mb-2">You have an ongoing game as <strong>{savedUsername}</strong></p>
                            <button 
                                onClick={handleRejoinGame} 
                                className="w-full bg-blue-500 hover:bg-blue-600 p-3 rounded-xl transition-all duration-300 cursor-pointer"
                                disabled={rejoinLoading}
                            >
                                {rejoinLoading ? (
                                    <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin mx-auto" />
                                ) : 'Rejoin Game'}
                            </button>
                        </div>
                    )}
                    
                    <label>Username</label>
                    <input name="username" type="text" className="border-b-1 outline-none p-3 text-white bg-black/50 " placeholder="Enter username" />
                    <div className="grid grid-cols-1 gap-4 w-fit m-auto">
                        <button 
                            onClick={props.handleNewGame} 
                            className={`${props.searching ? 'pointer-events-none' : ''} bg-white/4 hover:bg-white/10 p-3 active:bg-white/10 rounded-xl transition-[background] duration-300 cursor-pointer min-w-[88.11px]`}
                            disabled={props.searching}
                        >
                            {props.searching ? (
                                <div>
                                        Searching for a game...
                                <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin mx-auto flex items-center justify-center" >S</div>
                                </div>
                            ) : 'New Game'}
                        </button>
                    </div>
                    <div className="flex justify-center items-center text-white/20 text-sm">Leaderboard won't be counted for bots matches. <br/>
                    Server used is free tier, so it may be slow.</div>
                </div>
            </div>
        </div>
    )
}