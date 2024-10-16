package internal

import "gorm.io/gorm"

type PrivateMessage struct {
	gorm.Model
	SenderId   uint
	ReceiverID uint
	Message    string
}

func CreatePrivateMessage(senderID uint, receiverID uint, message string) error {
	privateMsg := PrivateMessage{SenderId: senderID, ReceiverID: receiverID, Message: message}
	result := DB.Create(&privateMsg)
	return result.Error
}

func GetPrivateMessages(senderID uint, receiverID uint) ([]PrivateMessage, error) {
	var messages []PrivateMessage
	result := DB.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)", senderID, receiverID, receiverID, senderID).Find(&messages)
	return messages, result.Error
}
