package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"math/rand"
	"net/http"
	"redislog"
	"time"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// dataFromDB mock retrieving data from database
func dataFromDB() []*User {
	return []*User{
		{
			"xiaoming",
			12,
		},
		{
			"xiaohong",
			13,
		},
		{
			"xiaobei",
			14,
		},
	}
}

// RedisLogger a gin.HandlerFunc wrapper
// extract request information, assemble to a record, and send to Redis server via goroutine
func RedisLogger(f gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		record := redislog.Record{
			RemoteAddr: c.Request.RemoteAddr,
			URL:        c.Request.URL.RequestURI(),
			AccessTime: time.Now().Unix(),
		}
		f(c)
		record.TimeExecuted = time.Now().Unix() - record.AccessTime
		record.BodyBytesSent = int64(c.Writer.Size())
		go redislog.SendRecord(record)
	}
}

func main() {
	engine := gin.Default()
	engine.GET("api/users", RedisLogger(func(c *gin.Context) {
		time.Sleep(time.Second * time.Duration(rand.Int63n(5)))
		c.JSON(http.StatusOK, gin.H{
			"data": dataFromDB(),
		})
	}))
	engine.GET("api/user/:name/age", RedisLogger(func(c *gin.Context) {
		users := dataFromDB()
		name := c.Param("name")
		var user *User
		for _, u := range users {
			if u.Name == name {
				user = u
			}
		}
		var age interface{}
		if user == nil {
			age = "unknown"
		} else {
			age = user.Age
		}
		c.JSON(http.StatusOK, gin.H{
			"age": age,
		})
	}))

	log.Fatalln(engine.Run(":9090"))
}
