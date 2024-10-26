package internal

import "gorm.io/gorm"

type Room struct {
	gorm.Model
	Name        string `gorm:"unique"`
	Description *string
	Massages    []Message
}

func CreateRoom(name string) (*Room, error) {
	room := Room{Name: name}
	result := DB.Create(&room)
	return &room, result.Error
}

func GetRoomByID(id uint) (Room, error) {
	var room Room
	result := DB.First(&room, id)
	return room, result.Error
}

func GetRoomByName(name string) (*Room, error) {
	var room Room
	err := DB.Where("name = ?", name).First(&room).Error
	return &room, err
}

func GetAllRooms() ([]Room, error) {
	var rooms []Room
	result := DB.Find(&rooms)
	return rooms, result.Error
}
