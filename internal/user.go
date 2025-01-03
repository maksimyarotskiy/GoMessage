package internal

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"unique"`
	Password string
}

func CreateUser(username, password string) error {
	user := User{Username: username, Password: password}
	result := DB.Create(&user)
	return result.Error
}

func GetUserByUsername(username string) (User, error) {
	var user User
	result := DB.Where("username = ?", username).First(&user)
	return user, result.Error
}

func GetUsernameByID(userID uint) (string, error) {
	var user User
	if err := DB.First(&user, userID).Error; err != nil {
		return "", err
	}
	return user.Username, nil
}
