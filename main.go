package main

import (
	"agent-integrations/esclation"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err.Error())
	}

	gin.SetMode(os.Getenv("GIN_MODE"))
	r := gin.Default()

	r.GET("/ping", pingAPI)
	r.POST("/api/init-session", initSession)
	r.GET("/api/session/:id", getSession)

	r.GET("/websocket", esclation.SocketServer)

	r.Run(":" + os.Getenv("GIN_PORT"))
}

func pingAPI(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
