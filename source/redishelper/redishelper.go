// redishelper/redis_helper.go

package redishelper

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type RedisHelper struct {
	client  *redis.Client
	config  *RedisConfig
	once    sync.Once
	mu      sync.RWMutex
}

type RedisConfig struct {
	Name         string
	Host         string
	Port         int
	Password     string
	DB           int
	StreamMaxLen int64
}

var (
	instance *RedisHelper
	once     sync.Once
)

func Instance() *RedisHelper {
	once.Do(func() {
		instance = &RedisHelper{

		}
	})
	return instance
}

// InitFromURL("redis://:pwd@localhost:6379?db=0&streamMaxLen=1000")
func (h *RedisHelper) InitFromURL(rawURL string) error {
	var initErr error
	h.once.Do(func() {
		u, err := url.Parse(rawURL)
		if err != nil {
			initErr = err
			return
		}
		port := 6379
		host := u.Hostname()
		if u.Port() != "" {
			fmt.Sscanf(u.Port(), "%d", &port)
		}
		password := ""
		if u.User != nil {
			pass, _ := u.User.Password()
			password = pass
		}
		db := 0
		if q := u.Query().Get("db"); q != "" {
			fmt.Sscanf(q, "%d", &db)
		}
		streamLen := int64(1000)
		if q := u.Query().Get("streamMaxLen"); q != "" {
			fmt.Sscanf(q, "%d", &streamLen)
		}
		opt := &redis.Options{
			Addr:     fmt.Sprintf("%s:%d", host, port),
			Password: password,
			DB:       db,
			MaxActiveConns: 200,
			PoolSize: 200,
		}
		client := redis.NewClient(opt)
		if err := client.Ping(ctx).Err(); err != nil {
			initErr = fmt.Errorf("redis ping error: %w", err)
			return
		}
		h.client = client
		h.config = &RedisConfig{
			Name:         host,
			Host:         host,
			Port:         port,
			Password:     password,
			DB:           db,
			StreamMaxLen: streamLen,
		}
	})
	return initErr
}

// 写入实时值和历史流
func (h *RedisHelper) SetRealtime(deviceID, point string, value any, quality string, timestamp time.Time) error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.client == nil {
		return errors.New("Redis not initialized")
	}
	ts := timestamp.UTC().Format(time.RFC3339)
	key := fmt.Sprintf("real:%s:%s", deviceID, point)
	streamKey := fmt.Sprintf("stream:%s:%s", deviceID, point)
	fields := map[string]interface{}{
		"value":   fmt.Sprintf("%v", value),
		"ts":      ts,
		"quality": quality,
		"dt": fmt.Sprintf("%T", value),
	}
	if err := h.client.HSet(ctx, key, fields).Err(); err != nil {
		return err
	}
	_ = h.client.SAdd(ctx, fmt.Sprintf("point:%s:points", deviceID), point).Err()
	return h.client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		MaxLen: h.config.StreamMaxLen,
		Values: fields,
	}).Err()
}

// 读取实时值
func (h *RedisHelper) GetRealtime(deviceID, point string) (map[string]string, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.client == nil {
		return nil, errors.New("Redis not initialized")
	}
	key := fmt.Sprintf("real:%s:%s", deviceID, point)
	return h.client.HGetAll(ctx, key).Result()
}

// 订阅某个设备的所有点
func (h *RedisHelper) SubscribeDevice(deviceID string, callback func(dev, pt string, data map[string]string)) error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.client == nil {
		return errors.New("Redis not initialized")
	}
	points, err := h.client.SMembers(ctx, fmt.Sprintf("point:%s:points", deviceID)).Result()
	if err != nil {
		return err
	}
	for _, point := range points {
		go h.subscribeStream(deviceID, point, callback)
	}
	return nil
}

func (h *RedisHelper) subscribeStream(deviceID, point string, callback func(string, string, map[string]string)) {
	streamKey := fmt.Sprintf("stream:%s:%s", deviceID, point)
	log.Println("subscribe ", streamKey)
	lastID := "$"
	for {
		streams, err := h.client.XRead(ctx, &redis.XReadArgs{
			Streams: []string{streamKey, lastID},
			Block:   0,
			Count:   1,
		}).Result()
		if err != nil {
			fmt.Println("XRead error:", err)
			time.Sleep(time.Second)
			continue
		}
		for _, s := range streams {
			for _, msg := range s.Messages {
				data := make(map[string]string)
				for k, v := range msg.Values {
					data[k] = fmt.Sprintf("%v", v)
				}
				callback(deviceID, point, data)
				lastID = msg.ID
			}
		}
	}
}

func (h *RedisHelper) Client() *redis.Client{
	return h.client
}