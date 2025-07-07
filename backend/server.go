package main

import (
	"backend/db"
	"backend/manager"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	///////////////////
	// Load .env file & connect to the database
	///////////////////
	godotenv.Load()
	db.Connect(os.Getenv("DATABASE_URL"))

	manager.StartServer() // Initialize the client manager

	// Your server code here...
}
