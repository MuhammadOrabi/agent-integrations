package main

import (
	"agent-integrations/redis"
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
)

var ctx = context.Background()

const charset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" + "0123456789"

func initSession(c *gin.Context) {

	body, err := c.GetRawData()
	if err != nil {
		panic(err)
	}

	rdb := redis.NewRedisClient()

	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 40)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	err = rdb.Set(ctx, string(b), body, 0).Err()
	if err != nil {
		panic(err)
	}

	c.JSON(200, gin.H{
		"sessionID": string(b),
	})
}

func getSession(c *gin.Context) {
	rdb := redis.NewRedisClient()

	val, err := rdb.Get(ctx, c.Param("id")).Result()
	if err != nil {
		panic(err)
	}

	data := json.RawMessage{}
	json.Unmarshal([]byte(val), &data)

	c.JSON(200, gin.H{
		"data": data,
	})
}
