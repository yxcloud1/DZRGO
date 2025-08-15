package conf

import (
	"log"

	"github.com/yxcloud1/go-comm/config"
)

type DB struct {
	Type string `json:"type"`
	Url  string `json:"url"`
}

type Config struct {
	DB             DB             `json:"db"`
	DataCollection DataCollection `json:"dc"`
	RedisConfig    RedisConfig    `json:"redisconfig"`
	InfluxDB       InfluxDB       `json:"influxdb"`
	FileWatch      FileWatch      `json:"filewatch"`
	Api            Api            `json:"api"`
}

type InfluxDB struct {
	Host   string `json:"host"`
	Token  string `json:"token"`
	Bucket string `json:"bucket"`
	Origin string `json:"origin"`
}

type RedisConfig struct {
	Url string `json:"url"`
}

type Api struct {
	ListenAddr string `json:"listenaddr"`
	Path       string `json:"path"`
}

type FileWatch struct {
	ApiURL         string   `json:"apiurl"`
	ApiTimeout     int      `json:"apitimeout"`
	DeviceID       string   `json:"device_id"`
	DeviceType     string   `json:"device_type"`
	WatchDirectory []string `json:"watchdirectory"`
	WatchFile      []string `json:"watchfile"`
}

type DataCollection struct {
	Drivers []string `json:"drivers"`
}

var (
	conf *Config = nil
)

func init() {
	conf = nil
}

func Conf() *Config {
	if conf == nil {
		conf = &Config{}
		log.Println("Workdir:", config.WorkDir())
		log.Println("load config:", config.Load(&conf))
		log.Println("save config:", config.Save(&conf))
	}
	return conf
}
