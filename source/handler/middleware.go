package handler

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)                  // 写入内存
	return w.ResponseWriter.Write(b) // 继续写入响应
}

func GlobalMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		method := c.Request.Method
		path := c.Request.URL.Path

		// 读取请求体
		var reqBody []byte
		if c.Request.Body != nil {
			reqBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		// 替换 response writer
		bw := &bodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = bw

		// 执行请求
		c.Next()

		statusCode := c.Writer.Status()

		duration := time.Since(start)
		var sb strings.Builder
		sb.WriteString("========== [HTTP REQUEST] ==========\r\n")
		for k, v := range c.Request.Header{
			sb.WriteString(fmt.Sprintf("%s: %s\r\n", k, strings.Join(v, ", ")))
		}
		sb.WriteString(fmt.Sprintf("[Time]     %v\r\n", start.Format("2006-01-02 15:04:05")))
		sb.WriteString(fmt.Sprintf("[Method]   %s\r\n", method))
		sb.WriteString(fmt.Sprintf("[Path]     %s\r\n", path))
		if len(reqBody) > 0 {
			sb.WriteString(fmt.Sprintf("[Request]  %s\r\n", string(reqBody)))
			sb.WriteString(fmt.Sprintf("[RequRaw]  %s\r\n", hex.EncodeToString(reqBody)))
		} else {
			sb.WriteString(fmt.Sprintf("[Request]  %v\r\n", c.Request.URL.Query()))
		}
		sb.WriteString(fmt.Sprintf("[Response] %s\r\n", bw.body.String()))
		sb.WriteString(fmt.Sprintf("[Http Code] %v\r\n", statusCode))
		sb.WriteString(fmt.Sprintf("[Duration] %v\r\n", duration))
		sb.WriteString("====================================")
		//logger.TxtLog(sb.String())
		log.Println(sb.String())
	}
}
