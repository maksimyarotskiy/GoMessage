package internal

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	Conn   *websocket.Conn
	RoomID uint
}

var clients = make(map[*Client]bool)

// var broadcast = make(chan MessagePayLoad)
var roomBroadcast = make(map[uint]chan MessagePayLoad)

type MessagePayLoad struct {
	Username string `json:"username"`
	Message  string `json:"message"`
	RoomID   uint   `json:"room_id"`
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

	roomID, err := getRoomIDFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ivalid room ID"})
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not upgrade connection"})
		return
	}
	defer conn.Close()

	//clients[conn] = true
	client := &Client{Conn: conn, RoomID: roomID}
	clients[client] = true

	history, err := GetRoomMesageHistory(roomID)
	if err == nil {
		for _, msg := range history {
			err := conn.WriteJSON(MessagePayLoad{
				Username: msg.Username,
				Message:  msg.Message,
				RoomID:   roomID,
			})
			if err != nil {
				fmt.Println("Error sending message history:", err)
				conn.Close()
				delete(clients, client)
				return
			}
		}
	}

	for {
		var msgPayload MessagePayLoad
		err := conn.ReadJSON(&msgPayload)
		if err != nil {
			fmt.Println("Error reading JSON:", err) // Отладка
			delete(clients, client)
			return
		}

		//Сохранение в БД
		message := Message{
			UserID:    user.ID,
			Username:  user.Username,
			RoomID:    roomID,
			Message:   msgPayload.Message,
			Timestamp: time.Now(),
		}
		if err := SaveMessage(message); err != nil {
			fmt.Println("Error saving message:", err)
			continue
		}

		if roomBroadcast[roomID] == nil {
			roomBroadcast[roomID] = make(chan MessagePayLoad)
			go HandleRoomMessages(roomID)
		}
		roomBroadcast[roomID] <- msgPayload

		//msgPayload.Username = username.(string)
		//broadcast <- msgPayload
	}
}

func SaveMessage(message Message) error {
	result := DB.Create(&message)
	return result.Error
}

// func GetMessageHistory() ([]Message, error) {
// 	var message []Message
// 	result := DB.Order("timestamp desc").Limit(10).Find(&message)
// 	return message, result.Error
// }

func GetRoomMesageHistory(roomID uint) ([]Message, error) {
	var messages []Message
	result := DB.Where("room_id = ?", roomID).Order("timestamp desc").Limit(10).Find(&messages)
	return messages, result.Error
}

// HandleMessages обрабатывает рассылку сообщений
// func HandleMessages() {
// 	for {
// 		msgPayload := <-broadcast

// 		for client := range clients {
// 			err := client.WriteJSON(MessagePayLoad{
// 				Username: msgPayload.Username,
// 				Message:  msgPayload.Message,
// 			})
// 			if err != nil {
// 				fmt.Println("Error writing JSON:", err)
// 				client.Close()
// 				delete(clients, client)
// 			}
// 		}
// 	}
// }

func HandleRoomMessages(roomID uint) {
	for msgPayload := range roomBroadcast[roomID] {
		for client := range clients {
			if client.RoomID == roomID {
				err := client.Conn.WriteJSON(msgPayload)
				if err != nil {
					fmt.Println("Error writing JSON:", err)
					client.Conn.Close()
					delete(clients, client)
				}
			}

		}
	}
}

func getRoomIDFromRequest(c *gin.Context) (uint, error) {
	roomID, err := strconv.ParseUint(c.Query("room_id"), 10, 64)
	return uint(roomID), err
}
