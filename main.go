package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-yaml/yaml"
	// "gopkg.in/yaml.v2"

	"github.com/gin-gonic/gin"
)

type FloodControlConfig struct {
	Limit  int           `yaml:"limit"`
	Period time.Duration `yaml:"period"`
}

func main() {

	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Не удалось прочитать файл YAML: %v", err)
	}

	router := gin.Default()

	var config FloodControlConfig
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Не удалось декодировать YAML: %v", err)
	}

	fmt.Println(config)

	fc := NewFloodControl(config)

	userID := int64(123)

	router.GET("/api/flood", func(c *gin.Context) {

		result, err := fc.Check(context.Background(), userID)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"status": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"status": result})
		}
	})

	router.Run(":8080")
}

type FloodControl interface {
	// Check возвращает false если достигнут лимит максимально разрешенного
	// кол-ва запросов согласно заданным правилам флуд контроля.
	Check(ctx context.Context, userID int64) (bool, error)
}

type FloodControlImpl struct {
	mutex       sync.Mutex
	callHistory map[int64][]time.Time
	config      FloodControlConfig
}

func NewFloodControl(config FloodControlConfig) *FloodControlImpl {
	return &FloodControlImpl{
		callHistory: make(map[int64][]time.Time),
		config:      config,
	}
}

func (fc *FloodControlImpl) Check(ctx context.Context, userID int64) (bool, error) {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()

	now := time.Now()

	if calls, ok := fc.callHistory[userID]; ok {
		for len(calls) > 0 && now.Sub(calls[0]) > fc.config.Period {
			calls = calls[1:]
		}

		if len(calls) >= fc.config.Limit {
			return false, errors.New("limit exceeded")
		}
	}

	fc.callHistory[userID] = append(fc.callHistory[userID], now)

	return true, nil
}
