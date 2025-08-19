package handler

import (
	"strings"
	"sync"
	"time"
	"unicode"
)

type DelayedMessage struct {
	delay       time.Duration
	timer       *time.Timer
	mu          sync.Mutex
	deviceType  string
	deviceId    string
	message     string
	endFlag     string
	raw_message []byte
	action      func(dt *DelayedMessage)
}

func NewDelayedTask(deviceType string, deviceId string, endFlag string,delay time.Duration, action func(dt *DelayedMessage)) *DelayedMessage {
	return &DelayedMessage{
		delay:      delay,
		action:     action,
		deviceType: deviceType,
		deviceId:   deviceId,
		endFlag: endFlag,
	}
}

func (d *DelayedMessage) Receive(message string, raw_message []byte) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.timer != nil && d.timer.Stop() {
		d.message += message
		d.raw_message = append(d.raw_message, raw_message...)
	} else {
		d.message = message
		d.raw_message = raw_message
	}
	if d.endFlag != "" && strings.HasSuffix(strings.TrimRightFunc(d.message, unicode.IsSpace),
											strings.TrimRightFunc(d.endFlag, unicode.IsSpace)) {
			if(d.timer != nil){
			d.timer.Stop()
			}
			d.action(d)
			return
	}
	d.timer = time.AfterFunc(d.delay, func() {
		d.action(d)
	})
}
