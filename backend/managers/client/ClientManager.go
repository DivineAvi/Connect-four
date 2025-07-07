package client

import (
	"sync"

	"github.com/gorilla/websocket"
)

type ClientManager struct {
	clients      map[string]*websocket.Conn
	mu           sync.Mutex
	connToclient map[*websocket.Conn]string
}

var (
	clientManager *ClientManager
	once          sync.Once
)

////////////////////////////////
// GetClientManager returns a singleton instance of ClientManager.
// It initializes the instance only once using sync.Once to ensure thread safety.
////////////////////////////////

func GetClientManager() *ClientManager {
	once.Do(func() {
		clientManager = &ClientManager{
			clients:      make(map[string]*websocket.Conn),
			connToclient: make(map[*websocket.Conn]string),
		}
	})
	return clientManager
}

////////////////////////////////
// AddClient adds a new client connection to the manager.
// It uses a mutex to ensure thread-safe access to the clients map.
////////////////////////////////

func (cm *ClientManager) AddClient(username string, conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.clients[username] = conn
	cm.connToclient[conn] = username
	println(cm.connToclient[conn], " Added")
	println(cm.clients[username], " Added")
}

///////////////////////////////
// RemoveClient removes a client from the manager by username.
// It closes the connection if it exists and deletes the entry from the clients map.
///////////////////////////////

func (cm *ClientManager) RemoveClient(username string, conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	println("Connection ", conn)
	if username != "" {
		if conn, exists := cm.clients[username]; exists {
			println("Deleting Enteries for ", username)
			delete(cm.connToclient, conn)
			delete(cm.clients, username)
			conn.Close()
		}
	} else {
		if username, exists := cm.connToclient[conn]; exists {
			println("Deleting Enteries for ", username)
			delete(cm.connToclient, conn)
			delete(cm.clients, username)
			conn.Close()

		}
	}
}

///////////////////////////////
// GetClient retrieves a client connection by username.
// It returns the connection and a boolean indicating if the client exists.
///////////////////////////////

func (cm *ClientManager) GetClient(username string) (*websocket.Conn, bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	conn, exists := cm.clients[username]
	return conn, exists
}
