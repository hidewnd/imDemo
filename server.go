package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int
	// 在线用户
	OnlineUserMap map[string]*User
	mapLock       sync.RWMutex
	// 消息广播channel
	Message chan string
}

// NewServer 创建Server接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:            ip,
		Port:          port,
		OnlineUserMap: make(map[string]*User),
		Message:       make(chan string),
	}
	return server
}

func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	s.Message <- sendMsg
}

func (s *Server) ListenMessage() {
	for {
		msg := <-s.Message
		s.mapLock.Lock()
		for _, cli := range s.OnlineUserMap {
			cli.C <- msg
		}
		s.mapLock.Unlock()
	}

}

// Handler handler处理
func (s *Server) Handler(conn net.Conn) {
	fmt.Print("链接建立成功")
	// 加入在线用户集
	user := NewUser(conn, s)
	user.Online()
	fmt.Println("[" + user.Addr + "]" + user.Name + ": online...")
	isLive := make(chan bool)
	// 广播用户发送信息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				fmt.Println("[" + user.Addr + "]" + user.Name + ": offline...")
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("conn Read err:", err)
				return
			}
			fmt.Println("[" + user.Addr + "]" + user.Name + ":" + string(buf))
			// 提取用户信息 去除 '\n'
			msg := string(buf[:n-1])
			user.DoMessage(msg)
		}
	}()
	// 超时强制监听检测处理
	for {
		select {
		case <-isLive:
		case <-time.After(time.Second * 60):
			user.DoMessage("You have been forced offline")
			fmt.Println("[" + user.Addr + "]" + user.Name + ": have been forced offline")
			close(user.C)
			err := conn.Close()
			if err != nil {
				return
			}
		}
	}
}

// Start 启动服务器接口
func (s *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net.listener err:", err)
		return
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println("net.listener err:", err)
		}
	}(listener)
	// 启动在线用户监听
	go s.ListenMessage()
	fmt.Println("start success....")
	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener Accept err:", err)
			continue
		}
		// handler
		go s.Handler(conn)
	}
}
