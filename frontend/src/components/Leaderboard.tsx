import { useEffect, useState } from 'react';

interface Player {
    id: number;
    username: string;
    wins: number;
    losses: number;
    draws: number;
    rating: number;
    created_at: string;
    updated_at: string;
}

interface LeaderboardProps {
    isOpen: boolean;
    onClose: () => void;
}

const Leaderboard = ({ isOpen, onClose }: LeaderboardProps) => {
    const [players, setPlayers] = useState<Player[]>([]);
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (isOpen) {
            fetchLeaderboard();
        }
    }, [isOpen]);

    const fetchLeaderboard = async () => {
        try {
            setLoading(true);
            setError(null);
            
            const apiUrl = window.location.hostname === 'localhost'
                ? 'http://localhost:8080/api/leaderboard'
                : `http://${window.location.hostname}:8080/api/leaderboard`;
            
            const response = await fetch(`${apiUrl}?limit=10`);
            
            if (!response.ok) {
                throw new Error(`Failed to fetch leaderboard: ${response.status} ${response.statusText}`);
            }
            
            const data = await response.json();
            setPlayers(data);
        } catch (err) {
            console.error('Error fetching leaderboard:', err);
            setError('Failed to load leaderboard. Please try again later.');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className={`fixed top-0 right-0 z-50 bg-black border-b md:border-l border-blue-500/30 transition-all duration-300  overflow-hidden md:w-[40vw] w-full ${isOpen ? 'max-h-[80vh]' : 'max-h-0'}`}>
            <div className="p-6 w-full max-w-3xl mx-auto">
                <h2 className="text-2xl font-bold text-center text-blue-400 mb-4">Leaderboard</h2>
                
                {loading ? (
                    <div className="flex justify-center py-8">
                        <div className="w-10 h-10 border-4 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
                    </div>
                ) : error ? (
                    <div className="text-red-400 text-center py-8">{error}</div>
                ) : players && players.length === 0 ? (
                    <div className="text-gray-400 text-center py-8">No players found. Be the first to join!</div>
                ) : (
                    <div
                        className="overflow-x-hidden max-h-[50vh] overflow-y-scroll"
                        style={{
                            WebkitOverflowScrolling: 'touch',
                            scrollbarWidth: 'thin',
                            scrollbarColor: '#60a5fa #1e293b4d', // blue-400 and blue-900/30
                        }}
                    >
                        {players && players.length > 0 ?(
                        <table className="w-full text-white ">
                            <thead>
                                <tr className="border-b border-gray-700">
                                    <th className="py-2 px-4 text-left">Rank</th>
                                    <th className="py-2 px-4 text-left">Player</th>
                                    <th className="py-2 px-4 text-center">W/L/D</th>
                                </tr>
                            </thead>
                            <tbody>
                                {players.map((player, index) => (
                                    <tr 
                                        key={player.id} 
                                        className={`border-b border-gray-700 ${index < 3 ? 'bg-black' : ''}`}
                                    >
                                        <td className="py-3 px-4">
                                            <div className="flex items-center">
                                                {index === 0 && (
                                                    <span className="text-yellow-400 mr-2">🏆</span>
                                                )}
                                                {index === 1 && (
                                                    <span className="text-gray-400 mr-2">🥈</span>
                                                )}
                                                {index === 2 && (
                                                    <span className="text-amber-700 mr-2">🥉</span>
                                                )}
                                                {index > 2 && (
                                                    <span className="mr-2">{index + 1}</span>
                                                )}
                                            </div>
                                        </td>
                                        <td className="py-3 px-4 font-medium">
                                            {player.username}
                                        </td>
                                     
                                        <td className="py-3 px-4 text-center">
                                            <span className="text-green-400">{player.wins}</span>/
                                            <span className="text-red-400">{player.losses}</span>/
                                            <span className="text-gray-400">{player.draws}</span>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                        ):
                        <div className="text-gray-400 text-center py-8">No players found. Be the first to join!</div>
                        }
                    </div>
                )}
                
                <div className="mt-6 text-center">
                    <button 
                        onClick={onClose}
                        className="bg-blue-500 hover:bg-blue-600 text-white py-2 px-6 rounded-lg transition-colors"
                    >
                        Close
                    </button>
                </div>
            </div>
        </div>
    );
};

export default Leaderboard; 