package internal

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan MessagePayLoad)

type MessagePayLoad struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

func HandleConnections(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
		return
	}

	user, err := GetUserByUsername(username.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not upgrade connection"})
		return
	}
	defer conn.Close()
	// конец проверки
	clients[conn] = true

	history, err := GetMessageHistory()
	if err == nil {
		for _, msg := range history {
			err := conn.WriteJSON(msg)
			if err != nil {
				fmt.Println("Error sending message history:", err)
				conn.Close()
				delete(clients, conn)
				return
			}
		}
	}

	for {
		var msgPayload MessagePayLoad
		err := conn.ReadJSON(&msgPayload)
		if err != nil {
			fmt.Println("Error reading JSON:", err) // Отладка
			delete(clients, conn)
			return
		}

		//Сохранение в БД
		message := Message{
			UserID:    user.ID,
			RoomID:    1, // пока одна
			Message:   msgPayload.Message,
			Timestamp: time.Now(),
		}
		if err := SaveMessage(message); err != nil {
			fmt.Println("Error saving message:", err)
			continue
		}

		msgPayload.Username = username.(string)
		broadcast <- msgPayload
	}
}

func SaveMessage(message Message) error {
	result := DB.Create(&message)
	return result.Error
}

func GetMessageHistory() ([]Message, error) {
	var message []Message
	result := DB.Order("timestamp desc").Limit(10).Find(&message)
	return message, result.Error
}

// HandleMessages обрабатывает рассылку сообщений
func HandleMessages() {
	for {
		msgPayload := <-broadcast

		//formattedMessage := fmt.Sprintf("%s: %s", msgPayload.Username, msgPayload.Message)
		for client := range clients {
			err := client.WriteJSON(MessagePayLoad{
				Username: msgPayload.Username,
				Message:  msgPayload.Message,
			})
			if err != nil {
				fmt.Println("Error writing JSON:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
