package tcpserver

import (
	"bufio"
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

		// 偏移地址
		fmt.Printf("%08X  ", i)

		// 十六进制部分
		for j := 0; j < bytesPerLine; j++ {
			if j < len(line) {
				fmt.Printf("%02X ", line[j])
			} else {
				fmt.Print("   ") // 占位对齐
			}
			if j == 7 {
				fmt.Print(" ") // 中间空格
			}
		}

		// ASCII 部分
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
	/*switch name {
	case "GB-18030", "gb18030":
		return "gb18030"
	case "GBK", "gbk":
		return "gbk"
	case "ISO-8859-1":
		return "iso-8859-1"
	case "WINDOWS-1252", "windows-1252":
		return "windows-1252"
	default:
		return strings.ToLower(name)
	}*/
}

const idleTimeout = 30 * time.Second

var (
	listeners []net.Listener
	stopOnce  sync.Once
	stopChan  = make(chan struct{})
	wg        sync.WaitGroup
	mu        sync.Mutex
	stopped   bool
	decoder   func([]byte) (string, error)
)

// Start 启动多个端口，每个端口有独立的 callback
func Start(handlers map[string]func(clientAddr, message string)) error {
	mu.Lock()
	stopped = false
	decoder = decodeToUTF8
	mu.Unlock()

	for addr, handler := range handlers {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("监听失败 %s: %v", addr, err)
		}
		fmt.Println("监听启动:", addr)

		mu.Lock()
		listeners = append(listeners, ln)
		mu.Unlock()

		wg.Add(1)
		go acceptLoop(ln, handler)
	}
	return nil
}

// Stop 关闭所有监听
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
		for _, ln := range listeners {
			ln.Close()
		}
		listeners = nil
		wg.Wait()
		fmt.Println("所有监听器已关闭")
	})
}

func acceptLoop(listener net.Listener, callback func(clientAddr, message string)) {
	defer wg.Done()
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-stopChan:
				return
			default:
				fmt.Println("接收连接失败:", err)
				continue
			}
		}
		wg.Add(1)
		go handleConnection(conn, callback)
	}
}

func handleConnection(conn net.Conn, callback func(clientAddr, message string)) {
	defer wg.Done()
	defer conn.Close()

	addr := conn.RemoteAddr().String()
	log.Printf("客户端已连接: %s\n", addr)

	reader := bufio.NewReader(conn)

	for {
		conn.SetReadDeadline(time.Now().Add(idleTimeout))

		data, err := reader.ReadBytes('\n')
		if err != nil {
			log.Printf("连接关闭: %s，原因: %v\n", addr, err)
			if len(data) > 0 {
				log.Printf("%s接收原始数据:", addr)
				HexDump(data)
				if decoder == nil {
					log.Printf("未设置解码器，原始数据: %v\n", data)
				}
				decoded, err := decoder(data)
				if err != nil {
					log.Printf("解码失败: %v\n", err)
				}

				callback(addr, trimNewline(decoded))
			}
			return
		}

		if decoder == nil {
			log.Printf("未设置解码器，原始数据: %v\n", data)
			continue
		}
		log.Printf("%s接收原始数据:", addr)
		HexDump(data)
		decoded, err := decoder(data)
		if err != nil {
			log.Printf("解码失败: %v\n", err)
			continue
		}

		callback(addr, trimNewline(decoded))

		//conn.Write([]byte("收到：" + decoded + "\n"))
	}
}

func trimNewline(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}
