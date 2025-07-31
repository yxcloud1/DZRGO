package main

import (
	"acetek-mes/conf"
	"acetek-mes/filewatch"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/yxcloud1/go-comm/logger"
	"github.com/yxcloud1/go-comm/winservice"
)

var (
	fw *filewatch.FileWatcher
)

func post(data string) ([]byte, error){
	jsonData, _ := json.Marshal(map[string]string{
		"data": data,
	})
	timeout := conf.Conf().FileWatch.ApiTimeout
	if timeout < 3{
		timeout = 3
	}
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	url := conf.Conf().FileWatch.ApiURL + "/" + conf.Conf().FileWatch.DeviceType + "/" + conf.Conf().FileWatch.DeviceID

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}


func watchCallback(data string) error {
	resp, err := post(data)
	logger.TxtLog(fmt.Sprintf("post data: %s, resp: %s, err: %+v", data, string(resp), err))
	return err
}

func run() error {
	fw = filewatch.NewFileWatcher()
	if conf.Conf().FileWatch.WatchDirectory != nil {
		for _, v := range conf.Conf().FileWatch.WatchDirectory {
			fw.WatchDirectory(v, watchCallback)
		}
	}
	if conf.Conf().FileWatch.WatchFile != nil {
		for _, v := range conf.Conf().FileWatch.WatchFile {
			fw.WatchFile(v, watchCallback)
		}
	}
	return nil
}
func stop() error {
	fw.Stop()
	return nil
}

func main() {
	winservice.RunAsService("LimsFileWatch", "LimsFileWatch", "LimsFileWatch", run, stop);
}
