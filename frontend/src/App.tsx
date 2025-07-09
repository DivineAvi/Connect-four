import { useEffect, useState } from "react";
import { GameManager } from "./scripts/GameManager";
import Lobby from "./pages/Lobby";
import Room from "./pages/Room";
import axios from "axios";
export default function App() {
  const [roomId, setRoomId] = useState<string | null>(null);
  const [gameStarted, setGameStarted] = useState<boolean>(false);
  const [searching, setSearching] = useState<boolean>(false);
  const [reconnecting, setReconnecting] = useState<boolean>(false);
  const [hasSavedGame, setHasSavedGame] = useState<boolean>(false);
  
const serverUrl = import.meta.env.VITE_SERVER_URL;
  const wsUrl = `${import.meta.env.VITE_WS_URL}/ws`;
    
  const gameManager = GameManager.getInstance(wsUrl);
  gameManager.SetGameStarted = setGameStarted;
  gameManager.SetReconnecting = setReconnecting;

  useEffect(() => {
    const checkRoomValidity = async () => {
    const savedState = localStorage.getItem('connect4GameState');
    if (savedState) {
      try {
        const gameState = JSON.parse(savedState);
        const isValid = await axios.get(serverUrl+'/join?roomId='+gameState.roomId+'&username='+gameState.username)
        if (isValid.status === 200) 
        if (gameState.roomId) {
          setRoomId(gameState.roomId);
          setHasSavedGame(true);
        }
      } catch (e) {
        console.error('Failed to parse saved game state', e);
      }
    }
  }
    checkRoomValidity();

    return () => {
      gameManager.socketManager.disconnect();
    };
  }, []);

  const handleReconnectGame = async () => {
    const savedState = localStorage.getItem('connect4GameState');
    if (!savedState) return;
    
    try {
      const gameState = JSON.parse(savedState);
      if (gameState.roomId && gameState.username) {
        setReconnecting(true);
        await gameManager.reconnectToGame(gameState.username, gameState.roomId);
      }
    } catch (error) {
      console.error("Failed to reconnect:", error);
      alert("Failed to reconnect to your game. Please try starting a new game.");
      setReconnecting(false);
      localStorage.removeItem('connect4GameState');
      setHasSavedGame(false);
    }
  };

  async function handleNewGame() {
    const username = (document.querySelector("input[name='username']") as HTMLInputElement)?.value;
    if (!username) {
      alert("Please enter a username");
      return;
    }
    
    try {
      localStorage.removeItem('connect4GameState');
      setHasSavedGame(false);
      
      await gameManager.new_game_request_handler(username);
      setSearching(true);
    } catch (error) {
      console.error("Failed to start new game:", error);
      alert("Failed to connect to the game server. Please try again.");
      setSearching(false);
    }
  }

  return (
    <div>
      {gameStarted || reconnecting ? (
        <Room />
      ) : (
        <Lobby 
          handleNewGame={handleNewGame} 
          handleReconnectGame={handleReconnectGame}
          roomId={roomId} 
          searching={searching}
          hasSavedGame={hasSavedGame}
        />
      )}
    </div>
  );
}