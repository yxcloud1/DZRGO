package filewatch

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/yxcloud1/go-comm/logger"
)

type watchCallback func(data string) error

type FileWatcher struct {
	watchers []*fsnotify.Watcher
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func NewFileWatcher() *FileWatcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &FileWatcher{
		ctx:    ctx,
			cancel: cancel,
	}
}

func (fw *FileWatcher) Stop() {
	fw.cancel()
	for _, w := range fw.watchers {
		_ = w.Close()
	}
	fw.wg.Wait()
}

func (fw *FileWatcher) ReadWholeFileWithRetry(filePath string) ([]byte, error) {
	const maxRetries = 3
	const retryDelay = 200 * time.Millisecond
	var lastErr error
	for i := 1; i <= maxRetries; i++ {
		data, err := os.ReadFile(filePath)
		if err != nil {
			lastErr = err
			time.Sleep(retryDelay)
			continue
		}
		return data, nil
	}
	return nil, lastErr
}

func (fw *FileWatcher) WatchDirectory(dir string, callback watchCallback) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.TxtErr(fmt.Sprintf("监控目录:%s 失败: %+v", dir, err))
		return err
	}
	fw.watchers = append(fw.watchers, watcher)

	err = watcher.Add(dir)
	if err != nil {
		logger.TxtErr(fmt.Sprintf("监控目录:%s 失败: %+v", dir, err))
		return err
	}
	logger.TxtLog(fmt.Sprintf("监控目录:%s", dir))

	fw.wg.Add(1)
	go func() {
		defer fw.wg.Done()
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&(fsnotify.Create|fsnotify.Write) != 0 {
					go func(path string) {
						time.Sleep(200 * time.Millisecond)
						if byts, err := fw.ReadWholeFileWithRetry(path); err == nil {
							if callback != nil {
								_ = callback(string(byts))
							}
						} else {
							logger.TxtErr(fmt.Sprintf("读取文件失败：%s, %+v", path, err))
						}
					}(event.Name)
				}
			case err := <-watcher.Errors:
				logger.TxtErr(fmt.Sprintf("目录监控出错:%+v", err))
			case <-fw.ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (fw *FileWatcher) WatchFile(filePath string, callback watchCallback) error {
	var (
		mu           sync.Mutex
		lastLines    []string
		lastModTime  time.Time
		lastFileSize int64
	)

	getNewLines := func() []string {
		mu.Lock()
		defer mu.Unlock()
		finfo, err := os.Stat(filePath)
		if err != nil {
			return nil
		}
		if finfo.ModTime().Equal(lastModTime) && finfo.Size() == lastFileSize {
			return nil
		}
		lastModTime = finfo.ModTime()
		lastFileSize = finfo.Size()

		file, err := os.Open(filePath)
		if err != nil {
			return nil
		}
		defer file.Close()

		var lines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return nil
		}
		start := len(lines)
		for i := len(lines) - 1; i >= 0 && len(lastLines) > 0; i-- {
			if lines[i] == lastLines[len(lastLines)-1] {
				start = i + 1
				break
			}
		}
		newLines := lines[start:]
		lastLines = lines
		return newLines
	}

	// 初始化记录一次
	getNewLines()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	fw.watchers = append(fw.watchers, watcher)

	err = watcher.Add(filePath)
	if err != nil {
		return err
	}

	fw.wg.Add(1)
	go func() {
		defer fw.wg.Done()
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&(fsnotify.Write) != 0 {
					time.Sleep(200 * time.Millisecond)
					if newLines := getNewLines(); len(newLines) > 0 {
						_ = callback(strings.Join(newLines, "\r\n"))
					}
				}
			case err := <-watcher.Errors:
				logger.TxtErr(fmt.Sprintf("文件监控出错:%+v", err))
			case <-fw.ctx.Done():
				return
			}
		}
	}()

	return nil
}