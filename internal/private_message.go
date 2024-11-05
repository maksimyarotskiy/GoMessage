package internal

import (
	"time"

	"gorm.io/gorm"
)

type PrivateMessage struct {
	gorm.Model
	SenderId   uint      `json:"sender_id"`
	ReceiverID uint      `json:"receiver_id"`
	Message    string    `json:"message"`
	Timestamp  time.Time `json:"timestamp"`
}

func CreatePrivateMessage(message *PrivateMessage) error {
	result := DB.Create(message)
	return result.Error
}

func GetPrivateMessages(senderID uint, receiverID uint) ([]PrivateMessage, error) {
	var messages []PrivateMessage
	result := DB.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)", senderID, receiverID, receiverID, senderID).Find(&messages)
	return messages, result.Error
}
