package manager

import (
	"sync"

	"github.com/gorilla/websocket"
)

type ClientManager struct {
	Clients map[string]*websocket.Conn
	mu      sync.Mutex
}

var (
	clientManagerInstance *ClientManager
)

////////////////////////////////
// GetClientManager returns a singleton instance of ClientManager.
// It initializes the instance only once using sync.Once to ensure thread safety.
////////////////////////////////

func GetClientManager() *ClientManager {
	once.Do(func() {
		clientManagerInstance = &ClientManager{
			Clients: make(map[string]*websocket.Conn),
		}
	})
	return clientManagerInstance
}

////////////////////////////////
// AddClient adds a new client connection to the manager.
// It uses a mutex to ensure thread-safe access to the Clients map.
////////////////////////////////

func (cm *ClientManager) AddClient(username string, conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.Clients[username] = conn
}

///////////////////////////////
// RemoveClient removes a client from the manager by username.
// It closes the connection if it exists and deletes the entry from the Clients map.
///////////////////////////////

func (cm *ClientManager) RemoveClient(username string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if conn, exists := cm.Clients[username]; exists {
		conn.Close()
		delete(cm.Clients, username)
	}
}

///////////////////////////////
// GetClient retrieves a client connection by username.
// It returns the connection and a boolean indicating if the client exists.
///////////////////////////////

func (cm *ClientManager) GetClient(username string) (*websocket.Conn, bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	conn, exists := cm.Clients[username]
	return conn, exists
}
