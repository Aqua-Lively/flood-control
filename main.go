package main

// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"sync"
// 	"time"
// )

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// func main() {
// 	// Создаем экземпляр FloodControl
// 	fc := NewFloodControl()

// 	// Параметры для проверки флуд-контроля
// 	userID := int64(123)
// 	N := 60 // Период времени в секундах
// 	K := 10 // Количество разрешенных вызовов за период времени N

// 	// Проверяем флуд-контроль
// 	for i := 0; i < 100; i++ {
// 		result, err := fc.Check(context.Background(), userID)
// 		if err != nil {
// 			fmt.Printf("Error checking flood control: %v\n", err)
// 		} else {
// 			fmt.Printf("Check result: %v\n", result)
// 		}
// 		time.Sleep(time.Duration(N/10) * time.Second)
// 	}
// }

func main() {
	// Создаем новый экземпляр Gin
	router := gin.Default()

	// Определяем маршрут для обработки GET-запросов на "/api/hello"
	router.GET("/api/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello, World!"})
	})

	// Запускаем сервер на порту 8080
	router.Run(":8080")
}

type FloodControlConfig struct {
	Limit  int
	Period time.Duration
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
