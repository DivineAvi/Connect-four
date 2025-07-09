# Connect Four

A full-stack, real-time multiplayer Connect Four game built with a Go backend and a React + TypeScript frontend.

## Features

- Real-time multiplayer gameplay with WebSocket support
- Persistent player stats and leaderboard
- Resume interrupted games
- Modern, responsive UI with Tailwind CSS
- REST API for player and leaderboard data

---

## Project Structure

```
Connect-four/
  backend/    # Go server, API, WebSocket, and database logic
  frontend/   # React + TypeScript client
```

---

## Getting Started

### Prerequisites

- Go 1.24+
- Node.js 18+
- PostgreSQL (or compatible database)
### DESIGN & WORKFLOW

1. **MATCHMAKING WORKFLOW**
  ![image](https://github.com/user-attachments/assets/9fa4e994-6ea2-43d3-b603-0e5d5dfbcfb0)

2. **NEW CONNECTION WORKFLOW **
  ![image](https://github.com/user-attachments/assets/c210fd41-54f2-4331-89f8-c2c3c1e0fa06)

### Backend Setup

1. **Configure Environment Variables**

   Create a `.env` file in `backend/` with:

   ```env
   DATABASE_URL=postgres://user:password@localhost:5432/connectfour
   PORT=8080
   ```

2. **Install Go dependencies**

   ```sh
   cd backend
   go mod tidy
   ```

3. **Run the server**

   ```sh
   go run server.go
   ```

   The backend will start on `localhost:8080`.

### Frontend Setup

1. **Install dependencies**

   ```sh
   cd frontend
   npm install
   ```

2. **Configure environment variables**

   Create a `.env` file in `frontend/` with:

   ```env
   VITE_SERVER_URL=http://localhost:8080/api
   VITE_WS_URL=ws://localhost:8080
   ```

3. **Start the development server**

   ```sh
   npm run dev
   ```

   The frontend will start on `localhost:5173` (default Vite port).

---

## API Endpoints

- `GET /api/leaderboard?limit=10`  
  Returns the top players.

- `GET /api/player?username=USERNAME`  
  Returns (or creates) a player.

- `GET /api/test/update-stats?winner=WINNER&loser=LOSER`  
  Updates stats for test purposes.

---

## Gameplay

- Enter a username and start a new game or rejoin an existing one.
- The game board updates in real time for both players.
- Player stats and leaderboard are updated after each game.

---

## Technologies Used

- **Backend:** Go, Gorilla WebSocket, PostgreSQL, godotenv
- **Frontend:** React, TypeScript, Vite, Tailwind CSS, Axios

---

## License

MIT 
