package main

import (
	"backend/db"
	"backend/managers/server"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	///////////////////////////////////////////
	// Load .env file & connect to the database
	///////////////////////////////////////////
	godotenv.Load()
	db.Connect(os.Getenv("DATABASE_URL"))

	var serverManager = server.GetServerManager()
	serverManager.StartServer()

	// Your server code here...
}
