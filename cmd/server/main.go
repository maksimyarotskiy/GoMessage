package main

import (
	"fmt"
	"os"

	"GoMessage/internal"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Загрузка переменных окружения из .env
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	internal.InitDB()

	internal.DB.AutoMigrate(&internal.User{}, &internal.Room{}, &internal.Message{}, &internal.PrivateMessage{})

	router := gin.Default()

	router.GET("/", homePage)
	router.POST("/register", internal.Register) // Регистрируем маршрут для регистрации
	router.POST("/login", internal.Login)
	router.GET("/ws", internal.AuthMiddleware(), internal.HandleConnections) // WebSocket
	router.GET("/ws/private", internal.AuthMiddleware(), internal.HandlePrivateConnections)
	router.POST("/rooms", internal.AuthMiddleware(), internal.CreateRoomHandler)            // Создать комнату
	router.GET("/rooms", internal.AuthMiddleware(), internal.GetRoomsHandler)               // Все комнаты
	router.DELETE("/rooms/:room_id", internal.AuthMiddleware(), internal.DeleteRoomHandler) // Удалить комнату

	// Запускаем обработку сообщений в горутине
	// go internal.HandleMessages()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Server started on :" + port)

	// Запускаем сервер
	if err := router.Run(":" + port); err != nil {
		panic("Error starting server: " + err.Error())
	}
}

// homePage обрабатывает GET-запросы к главной странице
func homePage(c *gin.Context) {
	c.String(200, "Welcome to the Chat Room!")
}
