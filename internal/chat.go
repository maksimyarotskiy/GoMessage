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
var privateClients = make(map[uint]*websocket.Conn)

var roomBroadcast = make(map[uint]chan RoomMessagePayload)

type RoomMessagePayload struct {
	Username string `json:"username"`
	Message  string `json:"message"`
	RoomID   uint   `json:"room_id"`
}

func HandlePrivateConnections(c *gin.Context) {
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

	receiverID, err := getOtherUserIDFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not upgrade connection"})
		return
	}
	defer conn.Close()

	privateClients[user.ID] = conn

	// Получение и отправка истории сообщений
	history, err := GetPrivateMessages(user.ID, receiverID)
	if err == nil {
		for _, msg := range history {
			err = conn.WriteJSON(msg)
			if err != nil {
				fmt.Println("Error sending message history:", err)
				conn.Close()
				delete(privateClients, user.ID)
				return
			}
		}
	}

	for {
		var msgPayload PrivateMessage
		err := conn.ReadJSON(&msgPayload)
		if err != nil {
			fmt.Println("Error reading JSON:", err)
			delete(privateClients, user.ID)
			return
		}

		//Сохранение в БД
		msgPayload.SenderId = user.ID
		msgPayload.ReceiverID = receiverID
		msgPayload.Timestamp = time.Now()

		err = CreatePrivateMessage(&msgPayload)
		if err != nil {
			fmt.Println("Error saving private message:", err)
			continue
		}

		// Отправка сообщения получателю, если он подключен
		if receiverConn, ok := privateClients[receiverID]; ok {
			err = receiverConn.WriteJSON(msgPayload)
			if err != nil {
				fmt.Println("Error sending private message:", err)
				receiverConn.Close()
				delete(privateClients, receiverID)
			}
		}
	}

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
		return
	}

	room, err := GetRoomByID(roomID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not upgrade connection"})
		return
	}
	defer conn.Close()

	client := &Client{Conn: conn, RoomID: room.ID}
	clients[client] = true

	history, err := GetRoomMesageHistory(room.ID)
	if err == nil {
		for _, msg := range history {
			err := conn.WriteJSON(RoomMessagePayload{
				Username: msg.Username,
				Message:  msg.Message,
				RoomID:   room.ID,
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
		var msgPayload RoomMessagePayload
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
			RoomID:    room.ID,
			Message:   msgPayload.Message,
			Timestamp: time.Now(),
		}
		if err := SaveMessage(message); err != nil {
			fmt.Println("Error saving message:", err)
			continue
		}

		if roomBroadcast[room.ID] == nil {
			roomBroadcast[room.ID] = make(chan RoomMessagePayload)
			go HandleRoomMessages(room.ID)
		}
		roomBroadcast[room.ID] <- msgPayload

		//msgPayload.Username = username.(string)
		//broadcast <- msgPayload
	}
}

func SaveMessage(message Message) error {
	result := DB.Create(&message)
	return result.Error
}

func GetRoomMesageHistory(roomID uint) ([]Message, error) {
	var messages []Message
	result := DB.Where("room_id = ?", roomID).Order("timestamp desc").Limit(10).Find(&messages)
	return messages, result.Error
}

func GetPrivateMessageHistory(userID1, userID2 uint) ([]PrivateMessage, error) {
	var messages []PrivateMessage
	result := DB.Where(
		"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		userID1, userID2, userID2, userID1,
	).Order("timestamp desc").Limit(50).Find(&messages)
	return messages, result.Error
}

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

func getOtherUserIDFromRequest(c *gin.Context) (uint, error) {
	otherUserID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	return uint(otherUserID), err
}
