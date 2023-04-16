package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

func NewUser(c net.Conn, server *Server) *User {
	userAddr := c.RemoteAddr().String()
	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: c,

		server: server,
	}
	go user.ListenMessage()
	return user
}

// Online 广播用户上线通知
func (u *User) Online() {
	u.server.mapLock.Lock()
	u.server.OnlineUserMap[u.Name] = u
	u.server.mapLock.Unlock()
	u.server.BroadCast(u, "online...")
}

// Offline 广播用户下线通知
func (u *User) Offline() {
	u.server.mapLock.Lock()
	delete(u.server.OnlineUserMap, u.Name)
	u.server.mapLock.Unlock()
	u.server.BroadCast(u, "offline...")
}

func (u *User) DoMessage(message string) {
	if len(message) > 0 {
		msg := strings.Split(message, "|")
		if len(msg) > 1 {
			switch msg[0] {
			case "rename":
				newName := msg[1]
				_, ok := u.server.OnlineUserMap[newName]
				if ok {
					u.SendMsg("this name is already used\n")
					return
				}
				u.server.mapLock.Lock()
				u.server.OnlineUserMap[newName] = u
				delete(u.server.OnlineUserMap, u.Name)
				u.server.mapLock.Unlock()
				u.Name = newName
				u.SendMsg("update newName success\n")
				break
			case "to":
				// to|用户名|消息
				if len(msg) < 3 {
					u.server.BroadCast(u, "Command format error\n")
					return
				}
				remoteUer, ok := u.server.OnlineUserMap[msg[1]]
				if !ok {
					u.server.BroadCast(u, msg[1]+" not found\n")
					return
				}
				remoteUer.SendMsg(msg[2])
			}
		} else {
			switch message {
			case "who":
				u.server.mapLock.Lock()
				for _, user := range u.server.OnlineUserMap {
					onlineMsg := "[" + user.Addr + "]" + user.Name + ":online\n"
					u.SendMsg(onlineMsg)
				}
				u.server.mapLock.Unlock()
				break
			default:
				u.server.BroadCast(u, message)
			}
		}
	}
}

func (u *User) ListenMessage() {
	for {
		msg := <-u.C
		_, err := u.conn.Write([]byte(msg + "\n"))
		if err != nil {
			fmt.Println("write err:", err)
			return
		}
	}
}

// SendMsg 给客户端发送消息
func (u *User) SendMsg(msg string) {
	_, err := u.conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("send Msg err:", err)
	}
}
