import { useEffect, useState } from "react";
import { GameManager } from "./scripts/GameManager";
import Lobby from "./pages/Lobby";
import Room from "./pages/Room";
export default function App() {
  const [roomId,] = useState<string | null>(null);
  const [gameStarted,] = useState<boolean>(false);
  const [seaching, Setsearching] = useState<Boolean>(false)
  const gameManager = GameManager.getInstance("ws://localhost:8080/ws");


  useEffect(() => {

    return () => {
      gameManager.socketManager.disconnect();
    };

  }, []);

  async function handleNewGame() {
    const username = (document.querySelector("input[name='username']") as HTMLInputElement)?.value;
    await gameManager.new_game_request_handler(username);
    Setsearching(true)
  }

  return (
    <div>
      {gameStarted ? <Room /> : <Lobby handleNewGame={handleNewGame} roomId={roomId} searching={seaching} />

      }
    </div>
  );
}