package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	goredis "github.com/go-redis/redis"
)

func main() {
	// Redis connection
	client := goredis.NewClient(&goredis.Options{
		// Config from server cmd flags
		Addr:     "localhost:6379", // DB host:port
		Password: "",               // Password
		DB:       0,                // Redis DB number
	})
	defer client.Close()

	// Check the connection is OK
	_, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}

	r := NewResolver(client)
	r.engine.Run()
}

type requestHelper struct {
	URL string `binding:"required"`
	ID  string
}

type Resolver struct {
	client *goredis.Client
	engine *gin.Engine
}

// gin router
func NewResolver(client *goredis.Client) *Resolver {
	r := &Resolver{
		client: client,
		engine: gin.Default(),
	}

	r.engine.GET("/:shorthex", func(c *gin.Context) {
		hex := c.Param("shorthex")
		origin, err := r.GetOriginURL(hex)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.Redirect(http.StatusMovedPermanently, origin)
	})

	r.engine.POST("/shorten", func(c *gin.Context) {
		var data requestHelper
		if err := c.ShouldBind(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		hex, err := r.SetShortenURL(data.URL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusOK, requestHelper{
			URL: data.URL,
			ID:  hex,
		})
	})

	return r
}

func (r *Resolver) SetShortenURL(url string) (string, error) {
	if !govalidator.IsURL(url) {
		return "", errors.New("invalid url")
	}
	hex, err := randomHex(3)
	if err != nil {
		return "", errors.New("gen hex failed")
	}
	err = r.client.Set("short:"+hex, url, 0).Err()
	if err != nil {
		return "", errors.New("set shorten failed")
	}

	return hex, nil
}

func (r *Resolver) GetOriginURL(hex string) (string, error) {
	val, err := r.client.Get("short:" + hex).Result()
	if err == goredis.Nil {
		return "", fmt.Errorf("short url not found: %v", err)
	}
	if err != nil {
		return "", errors.New("set shorten failed")
	}
	return val, nil
}

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
