package internal

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)

// type Message struct {
// 	Username string `json:"username"`
// 	Message  string `json:"message"`
// }

func HandleConnections(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
		return
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil) // Изменяем здесь
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not upgrade connection"})
		return
	}
	defer conn.Close()

	clients[conn] = true

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("Error reading JSON:", err) // Отладка
			delete(clients, conn)
			return
		}

		msg.Username = username.(string)
		broadcast <- msg
	}
}

// HandleMessages обрабатывает рассылку сообщений
func HandleMessages() {
	for {
		msg := <-broadcast

		formattedMessage := fmt.Sprintf("%s: %s", msg.Username, msg.Message)
		for client := range clients {
			err := client.WriteJSON(Message{
				Username: msg.Username,
				Message:  formattedMessage,
			})
			if err != nil {
				fmt.Println("Error writing JSON:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
