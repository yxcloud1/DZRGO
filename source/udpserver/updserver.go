package udpserver

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/saintfish/chardet"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
)

func HexDump(data []byte) {
	const bytesPerLine = 16
	for i := 0; i < len(data); i += bytesPerLine {
		end := i + bytesPerLine
		if end > len(data) {
			end = len(data)
		}
		line := data[i:end]

		fmt.Printf("%08X  ", i)
		for j := 0; j < bytesPerLine; j++ {
			if j < len(line) {
				fmt.Printf("%02X ", line[j])
			} else {
				fmt.Print("   ")
			}
			if j == 7 {
				fmt.Print(" ")
			}
		}

		fmt.Print(" |")
		for _, b := range line {
			if b >= 32 && b <= 126 {
				fmt.Printf("%c", b)
			} else {
				fmt.Print(".")
			}
		}
		fmt.Println("|")
	}
}

func AutoDecode(data []byte) (string, error) {
	if utf8.Valid(data) {
		return string(data), nil
	}

	candidates := []struct {
		name    string
		decoder *encoding.Decoder
	}{
		{"GB18030", simplifiedchinese.GB18030.NewDecoder()},
		{"GBK", simplifiedchinese.GBK.NewDecoder()},
		{"BIG5", traditionalchinese.Big5.NewDecoder()},
		{"Shift-JIS", japanese.ShiftJIS.NewDecoder()},
		{"EUC-KR", korean.EUCKR.NewDecoder()},
		{"Windows-1252", charmap.Windows1252.NewDecoder()},
	}

	for _, cand := range candidates {
		reader := transform.NewReader(bytes.NewReader(data), cand.decoder)
		decoded, err := ioutil.ReadAll(reader)
		if err == nil && utf8.Valid(decoded) {
			return string(decoded), nil
		}
	}

	return "", errors.New("unable to decode: unsupported or unknown encoding")
}

func decodeToUTF8(data []byte) (string, error) {
	detector := chardet.NewTextDetector()
	result, err := detector.DetectBest(data)
	if err != nil {
		return "", fmt.Errorf("detect failed: %v", err)
	}

	encodingName := normalizeCharset(result.Charset)
	encoding, _ := charset.Lookup(encodingName)
	if encoding == nil {
		return "", fmt.Errorf("unsupported charset: %s", encodingName)
	}

	reader := transform.NewReader(bytes.NewReader(data), encoding.NewDecoder())
	utf8Data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("decode failed: %v", err)
	}

	return string(utf8Data), nil
}

func normalizeCharset(name string) string {
	return "gb18030"
}

const idleTimeout = 30 * time.Second

var (
	listeners []*net.UDPConn
	stopOnce  sync.Once
	stopChan  = make(chan struct{})
	wg        sync.WaitGroup
	mu        sync.Mutex
	stopped   bool
	decoder   func([]byte) (string, error)
)

func Start(handlers map[string]func(clientAddr, message string, raw[] byte)) error {
	mu.Lock()
	stopped = false
	decoder = decodeToUTF8
	mu.Unlock()

	for addr, handler := range handlers {
		udpAddr, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			log.Printf("解析地址失败 %s: %v", addr, err)
			continue
		}

		conn, err := net.ListenUDP("udp", udpAddr)
		if err != nil {
			log.Printf("监听失败 %s: %v", addr, err)
			continue
		}
		fmt.Println("UDP监听启动:", addr)

		mu.Lock()
		listeners = append(listeners, conn)
		mu.Unlock()

		wg.Add(1)
		go readLoop(conn, handler)
	}
	return nil
}

func Stop() {
	stopOnce.Do(func() {
		mu.Lock()
		if stopped {
			mu.Unlock()
			return
		}
		stopped = true
		mu.Unlock()

		close(stopChan)
		for _, conn := range listeners {
			conn.Close()
		}
		listeners = nil
		wg.Wait()
		fmt.Println("所有UDP监听器已关闭")
	})
}

func readLoop(conn *net.UDPConn, callback func(clientAddr, message string, raw []byte)) {
	defer wg.Done()
	buf := make([]byte, 4096)

	for {
		conn.SetReadDeadline(time.Now().Add(idleTimeout))
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			select {
			case <-stopChan:
				return
			default:
				//log.Printf("UDP读取失败: %v", err)
				continue
			}
		}

		data := append([]byte(nil), buf[:n]...) // 拷贝数据防止复用
		wg.Add(1)
		go func(addr *net.UDPAddr, msg []byte) {
			defer wg.Done()
			processPacket(conn, addr, msg, callback)
		}(remoteAddr, data)
	}
}

func processPacket(conn *net.UDPConn, remoteAddr *net.UDPAddr, data []byte, callback func(clientAddr, message string, raw []byte)) {
	addr := remoteAddr.String()
	log.Printf("%s接收原始数据:", addr)
	HexDump(data)

	decoded, err := decoder(data)
	if err != nil {
		log.Printf("解码失败: %v\n", err)
		return
	}
	callback(addr, (decoded), data)
}

func trimNewline(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}