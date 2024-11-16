package internal

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Room struct {
	gorm.Model
	Name        string `gorm:"unique"`
	Description string
	Massages    []Message
	OwnerID     uint `json:"owner_id"`
}

func CreateRoom(name string, description string, ownerID uint) (Room, error) {
	room := Room{Name: name, Description: description, OwnerID: ownerID}
	result := DB.Create(&room)
	return room, result.Error
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

func GetRoomsHandler(c *gin.Context) {
	var rooms []Room
	result := DB.Find(&rooms)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		return
	}
	c.JSON(http.StatusOK, rooms)
}

func IsRoomOwner(userID uint, roomID uint) bool {
	var room Room
	if err := DB.First(&room, roomID).Error; err != nil {
		return false
	}
	return room.OwnerID == userID
}

func CreateRoomHandler(c *gin.Context) {
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

	var roomData struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&roomData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	room, err := CreateRoom(roomData.Name, roomData.Description, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create room"})
		return
	}

	c.JSON(http.StatusOK, room)
}

func DeleteRoomHandler(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
	}

	user, err := GetUserByUsername(username.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	roomID, err := strconv.ParseUint(c.Param("room_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	if !IsRoomOwner(user.ID, uint(roomID)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not the owner of this room"})
		return
	}

	if err := DB.Delete(&Room{}, roomID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete room"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Room deleted"})
}
