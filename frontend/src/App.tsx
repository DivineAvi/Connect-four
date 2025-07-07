import { useEffect,useState } from "react";
import { GameManager } from "./components/GameManager";
export default function App() {
  const [roomId,SetroomId] = useState<string | null>(null);
  const [gameStarted,SetgameStarted] = useState<boolean>(false);

  const gameManager = GameManager.getInstance("ws://localhost:8080/ws");

  useEffect(() => {

    return () => {
      gameManager.socketManager.disconnect();
    };

  }, []);


  async function handleNewGame() {
    const username = (document.querySelector("input[name='username']") as HTMLInputElement)?.value;
    await gameManager.new_game_request_handler(username);
  }

  return (
    <div className="w-full min-h-screen bg-black text-white flex items-center justify-center p-2">
      <div className="flex flex-col gap-4 w-96 border-2 border-white/9  p-5 rounded-xl">
        <label >Username</label>
        <input name="username" type="text" className=" border-b-1 outline-none p-3 " placeholder="Enter username" />
        <div className={`${roomId ? 'grid-cols-2' : 'grid-cols-1'} grid gap-4 w-fit m-auto`}>
          <button onClick={handleNewGame} className="bg-white/4 hover:bg-white/10 p-3 rounded-xl transition-[background] duration-300 cursor-pointer min-w-[88.11px] ">{roomId ? 'New' : 'Play'}</button>
          <div className={`${roomId ? 'flex' : 'hidden'} items-center justify-center`}>

            <button className=" bg-white/4 hover:bg-white/10 p-3 rounded-xl transition-[background] duration-300 cursor-pointer relative">Continue
              <div className="absolute  pointer-events-none  top-[100%] left-1/2 -translate-x-1/2 mt-2 text-orange-500 text-xs whitespace-nowrap bg-white p-2 rounded-lg shadow-lg transition-all duration-300 z-10 hover:border-orange hover:border-1 ">
                <div className="absolute -top-2 left-1/2 -translate-x-1/2 w-0 h-0 border-l-8 border-r-8 border-b-8 border-l-transparent border-r-transparent border-b-white"></div>
                Your match is pending.
              </div>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}