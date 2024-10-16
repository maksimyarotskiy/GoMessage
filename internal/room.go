package internal

import "gorm.io/gorm"

type Room struct {
	gorm.Model
	Name     string `gorm:"unique"`
	Massages []Message
}

func CreateRoom(name string) error {
	room := Room{Name: name}
	result := DB.Create(&room)
	return result.Error
}

func GetRoomByID(id uint) (Room, error) {
	var room Room
	result := DB.First(&room, id)
	return room, result.Error
}
