import { useEffect, useState } from "react";
import { GameManager } from "./scripts/GameManager";
import Lobby from "./pages/Lobby";
import Room from "./pages/Room";

export default function App() {
  const [roomId, setRoomId] = useState<string | null>(null);
  const [gameStarted, setGameStarted] = useState<boolean>(false);
  const [searching, setSearching] = useState<boolean>(false);
  const [reconnecting, setReconnecting] = useState<boolean>(false);
  const [hasSavedGame, setHasSavedGame] = useState<boolean>(false);
  
  // Use the correct WebSocket URL
  const wsUrl = window.location.hostname === 'localhost' 
    ? "ws://localhost:8080/ws"
    : `ws://${window.location.hostname}:8080/ws`;
    
  const gameManager = GameManager.getInstance(wsUrl);
  gameManager.SetGameStarted = setGameStarted;
  gameManager.SetReconnecting = setReconnecting;

  useEffect(() => {
    // Check if we have a saved game to reconnect to
    const savedState = localStorage.getItem('connect4GameState');
    if (savedState) {
      try {
        const gameState = JSON.parse(savedState);
        if (gameState.roomId) {
          setRoomId(gameState.roomId);
          setHasSavedGame(true);
          // Don't set reconnecting to true automatically
        }
      } catch (e) {
        console.error('Failed to parse saved game state', e);
      }
    }

    return () => {
      gameManager.socketManager.disconnect();
    };
  }, []);

  // Function to handle reconnect game
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
      // Clear the saved game state so user can start fresh
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
      // Clear any existing saved game
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