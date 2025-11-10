// main.go
package main

import (
	"acetek-mes/conf"
	"acetek-mes/influxdb2"
	"acetek-mes/valconv"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"acetek-mes/redishelper" // 替换为实际模块路径

	"github.com/yxcloud1/go-comm/db"
)

type ArchiveValue struct {
	Quality    string      `json:"quality"`
	Value      interface{} `json:"value"`
	Ts         time.Time   `json:"ts"`
	DataType   string      `json:"dt"`
	updateTime time.Time
}

var (
	archive_values = make(map[string]ArchiveValue)
	mut            sync.Mutex
	mutWrite       sync.Mutex
	mutMap         sync.Mutex
)

func updateTagValue(item_id string, driver_id string, value interface{}, quality string, ts time.Time) error {

	sCommand := `exec sp_dc_update_dc_value @item_id = ? , @driver_id = ? , @value = ? , @quality = ? , @ts = ? `
	return db.DB().ExecuteSQL(sCommand, item_id, driver_id, value, quality, ts)
}

func isChanged(device, point string, data map[string]string) (string, string, *ArchiveValue, error) {
	value := ArchiveValue{}
	byts, err := json.Marshal(data)
	if err != nil {
		return device, point, nil, err
	}
	err = json.Unmarshal(byts, &value)
	if err != nil {
		return device, point, nil, err
	}
	key := fmt.Sprintf("%s.%s", device, point)
	mutMap.Lock()
	defer func() {
		mutMap.Unlock()
	}()
	v, ok := archive_values[key]
	if !ok {
		archive_values[key] = value
		return device, point, &value, nil
	}
	if v.Quality != value.Quality || fmt.Sprintf("%v", v.Value) != fmt.Sprintf("%v", value.Value) || v.updateTime.Add(time.Minute).Before(time.Now()){
		value.updateTime = time.Now()
		archive_values[key] = value
		return device, point, &value, nil
	}
	return device, point, nil, nil
}

// 假设这是数据库写入函数（替换为真实逻辑）
func writeToDatabase(device, point string, data map[string]string) {
	mutWrite.Lock()
	defer mutWrite.Unlock()
	_, _, value, err := isChanged(device, point, data)
	if value != nil && err == nil {
		tags := make(map[string]string)
		tags["quality"] = value.Quality
		fields := make(map[string]interface{})
		fields[point], _ = valconv.StringToTargetType(fmt.Sprintf("%v", value.Value), value.DataType)
		if fields[point] != nil {
			go func() {
				mut.Lock()
				defer mut.Unlock()
				if err = influxdb2.Write(device, tags, fields, value.Ts); err != nil {
					log.Println("arhive error: ", err)
				}
				if err = updateTagValue(point, device, value.Value, value.Quality, value.Ts); err != nil {
					log.Println("update tag error: ", err)
				}
			}()
		}
	} else {
		if err != nil {
			log.Println("archive error:", err)
		}
	}
}

func init() {
	db.SetOption(conf.Conf().DB.Type, conf.Conf().DB.Url)
	influxdb2.SetOption(conf.Conf().InfluxDB.Host, conf.Conf().InfluxDB.Token,
		conf.Conf().InfluxDB.Bucket, conf.Conf().InfluxDB.Origin)
	log.Println("init redis ", redishelper.Instance().InitFromURL(conf.Conf().RedisConfig.Url))

}

func main() {
	if redishelper.Instance().Client() == nil {
		log.Println("redis not connected")
		return
	}
	keys, err := redishelper.Instance().Client().Keys(context.Background(), "point:*:points").Result()

	if err != nil {
		log.Printf("Fetch device keys failed: %v", err)
	}
	log.Println(keys)
	for _, key := range keys {
		deviceID := strings.TrimPrefix(strings.TrimSuffix(key, ":points"), "point:")
		reals, err := redishelper.Instance().Client().SMembers(context.Background(), key).Result()
		if err != nil {
			log.Printf("Read points error for %s: %v", deviceID, err)
			continue
		}
		for _, pt := range reals {
			data, err := redishelper.Instance().GetRealtime(deviceID, pt)
			if err == nil {
				go writeToDatabase(deviceID, pt, data)
			}
		}
		redishelper.Instance().SubscribeDevice(deviceID, func(dev, pt string, data map[string]string) {
			go writeToDatabase(dev, pt, data)
		})
	}
	select {}
}
