package internal

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	RoomID    uint
	UserID    uint
	Message   string
	Timestamp time.Time
}

func CreateMessage(roomID, userID uint, message string) error {
	msg := Message{RoomID: roomID, UserID: userID, Message: message}
	result := DB.Create(&msg)
	return result.Error
}

func GetMessagesByRoomID(roomID uint) ([]Message, error) {
	var messages []Message
	result := DB.Where("room_id = ?", roomID).Find(&messages)
	return messages, result.Error
}
