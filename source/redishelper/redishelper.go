package redishelper

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

var loggerFunc = func(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func SetLogger(f func(format string, v ...interface{})) {
	loggerFunc = f
}

func DisableLog() {
	loggerFunc = func(format string, v ...interface{}) {}
}

type RedisHelper struct {
	client     *redis.Client
	config     *RedisConfig
	mu         sync.RWMutex
	stopCh     chan struct{}
	ready      bool
	subscribed map[string]func(dev, pt string, data map[string]string) // deviceID -> callback
	subMu      sync.Mutex
}

type RedisConfig struct {
	Name         string
	Host         string
	Port         int
	Password     string
	DB           int
	StreamMaxLen int64
	URL          string
}

var (
	instance *RedisHelper
	once     sync.Once
)

func Instance() *RedisHelper {
	once.Do(func() {
		instance = &RedisHelper{
			stopCh:     make(chan struct{}),
			subscribed: make(map[string]func(string, string, map[string]string)),
		}
	})
	return instance
}

func (h *RedisHelper) InitFromURL(rawURL string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	u, err := url.Parse(rawURL)
	if err != nil {
		return err
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
		PoolSize: 1200,
	}

	client := redis.NewClient(opt)
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping error: %w", err)
	}

	h.client = client
	h.config = &RedisConfig{
		Name:         host,
		Host:         host,
		Port:         port,
		Password:     password,
		DB:           db,
		StreamMaxLen: streamLen,
		URL:          rawURL,
	}
	h.ready = true

	go h.keepAlive()

	loggerFunc("[RedisHelper] Connected to %s:%d (db=%d)", host, port, db)
	return nil
}

func (h *RedisHelper) keepAlive() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	failCount := 0
	for {
		select {
		case <-ticker.C:
			h.mu.RLock()
			client := h.client
			cfg := h.config
			h.mu.RUnlock()

			if client == nil || cfg == nil {
				continue
			}

			if err := client.Ping(ctx).Err(); err != nil {
				failCount++
				loggerFunc("[RedisHelper] Ping failed (%d): %v", failCount, err)
				time.Sleep(time.Second * time.Duration(failCount))
				h.reconnect()
			} else {
				failCount = 0
			}
		case <-h.stopCh:
			return
		}
	}
}

func (h *RedisHelper) reconnect() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.config == nil {
		return
	}

	u, _ := url.Parse(h.config.URL)
	port := 6379
	if u.Port() != "" {
		fmt.Sscanf(u.Port(), "%d", &port)
	}
	opt := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", u.Hostname(), port),
		Password: h.config.Password,
		DB:       h.config.DB,
		PoolSize: 1200,
	}

	client := redis.NewClient(opt)
	if err := client.Ping(ctx).Err(); err != nil {
		loggerFunc("[RedisHelper] Reconnect failed: %v", err)
		return
	}

	h.client = client
	h.ready = true
	loggerFunc("[RedisHelper] Reconnected successfully.")

	h.subMu.Lock()
	for dev, cb := range h.subscribed {
		go func(deviceID string, callback func(string, string, map[string]string)) {
			loggerFunc("[RedisHelper] Restoring subscription for %s", deviceID)
			err := h.SubscribeDevice(deviceID, callback)
			if err != nil {
				loggerFunc("[RedisHelper] Failed to resubscribe %s: %v", deviceID, err)
			}
		}(dev, cb)
	}
	h.subMu.Unlock()
}

func (h *RedisHelper) SetRealtime(deviceID, point string, value any, quality string, timestamp time.Time) error {
	h.mu.RLock()
	client := h.client
	cfg := h.config
	h.mu.RUnlock()

	if client == nil {
		return errors.New("Redis not initialized")
	}
	ts := timestamp.UTC().Format(time.RFC3339)
	key := fmt.Sprintf("real:%s:%s", deviceID, point)
	streamKey := fmt.Sprintf("stream:%s:%s", deviceID, point)

	fields := map[string]interface{}{
		"value":   fmt.Sprintf("%v", value),
		"ts":      ts,
		"quality": quality,
		"dt":      fmt.Sprintf("%T", value),
	}

	if err := client.HSet(ctx, key, fields).Err(); err != nil {
		return err
	}
	_ = client.SAdd(ctx, fmt.Sprintf("point:%s:points", deviceID), point).Err()

	return client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		MaxLen: cfg.StreamMaxLen,
		Values: fields,
	}).Err()
}

func (h *RedisHelper) GetRealtime(deviceID, point string) (map[string]string, error) {
	h.mu.RLock()
	client := h.client
	h.mu.RUnlock()

	if client == nil {
		return nil, errors.New("Redis not initialized")
	}
	key := fmt.Sprintf("real:%s:%s", deviceID, point)
	return client.HGetAll(ctx, key).Result()
}

func (h *RedisHelper) SubscribeDevice(deviceID string, callback func(dev, pt string, data map[string]string)) error {
	h.mu.RLock()
	client := h.client
	h.mu.RUnlock()

	if client == nil {
		return errors.New("Redis not initialized")
	}

	points, err := client.SMembers(ctx, fmt.Sprintf("point:%s:points", deviceID)).Result()
	if err != nil {
		return err
	}
	if len(points) == 0 {
		return errors.New("no points found for device")
	}

	h.subMu.Lock()
	h.subscribed[deviceID] = callback
	h.subMu.Unlock()

	go h.subscribeAllStreams(deviceID, points, callback)
	return nil
}

func (h *RedisHelper) subscribeAllStreams(deviceID string, points []string, callback func(string, string, map[string]string)) {
	streams := make([]string, 0, len(points))
	streamIds := make([]string, 0, len(points))
	for _, pt := range points {
		streamKey := fmt.Sprintf("stream:%s:%s", deviceID, pt)
		streams = append(streams, streamKey) // 每个流后跟一个 ID
		streamIds = append(streamIds, "$")
	}
	streams = append(streams, streamIds...)
	log.Println("subscribe ", deviceID)
	//loggerFunc("[Subscriber] Subscribing %d points for %s", len(points), deviceID)

	for {
		h.mu.RLock()
		client := h.client
		h.mu.RUnlock()

		if client == nil {
			time.Sleep(time.Second)
			continue
		}
		//log.Println(streams)
		msgs, err := client.XRead(ctx, &redis.XReadArgs{
			Streams: streams,
			Block:   0,
		}).Result()

		if err != nil {
			loggerFunc("[Subscriber] XRead error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		for _, s := range msgs {
			for _, msg := range s.Messages {
				data := make(map[string]string)
				for k, v := range msg.Values {
					data[k] = fmt.Sprintf("%v", v)
				}
				parts := strings.Split(s.Stream, ":")
				point := parts[len(parts)-1]
				callback(deviceID, point, data)
				callback(deviceID, s.Stream, data)
			}
		}
	}
}

func (h *RedisHelper) Client() *redis.Client {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.client
}
