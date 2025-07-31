package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
    ln, err := net.Listen("tcp", ":9100") // 常用于网络打印端口
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("打印服务已启动，监听端口 9100")

    for {
        conn, err := ln.Accept()
        if err != nil {
            log.Println("连接失败:", err)
            continue
        }
        go handleConnection(conn)
    }
}

func handleConnection(conn net.Conn) {
    defer conn.Close()

    clientAddr := conn.RemoteAddr().String()
    fmt.Println("接收到打印连接来自:", clientAddr)

    scanner := bufio.NewScanner(conn)
    for scanner.Scan() {
        line := scanner.Text()
        fmt.Printf("[来自 %s] %s\n", clientAddr, line)

        // 可选：写入日志文件
        f, _ := os.OpenFile("print_log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        f.WriteString(line + "\n")
        f.Close()
    }

    if err := scanner.Err(); err != nil {
        log.Println("读取失败:", err)
    }

    fmt.Println("连接断开:", clientAddr)
}