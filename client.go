package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

var serverIp string
var serverPort int

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn

	flag int
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", client.ServerIp, client.ServerPort))
	if err != nil {
		fmt.Println("dial err:", err)
		return nil
	}
	client.conn = conn
	return client
}

func (client *Client) Menu() bool {
	var code int
	fmt.Println("1: 公聊模式")
	fmt.Println("2: 私聊模式")
	fmt.Println("3: 更新用户名")
	fmt.Println("0: 退出")
	_, err := fmt.Scanln(&code)
	if err != nil {
		return false
	}
	if code >= 0 && code <= 3 {
		client.flag = code
		return true
	}
	return false
}

func (client *Client) updateUserName() bool {
	fmt.Println(">>> 请输入用户名")
	_, err := fmt.Scanln(&client.Name)
	if err != nil {
		return false
	}
	sendMsg := "rename|" + client.Name + "\n"
	_, err = client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println(">>> conn write err:", err)
		return false
	}
	return true
}

func (client *Client) selectUser() bool {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println(">>> conn write err:", err)
		return false
	}
	return true
}
func (client *Client) PrivateChat() {
	var remote, chatMsg string
	client.selectUser()
	for remote != "exit" {
		fmt.Println(">>> 请输入发送对象，输入exit退出")
		_, err := fmt.Scanln(&remote)
		if err != nil {
			return
		}
		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remote + "|" + chatMsg + "\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println(">>> conn write err:", err)
					break
				}
			}
			chatMsg = ""
			fmt.Println(">>> 请输入发送内容，输入exit退出")
			_, err := fmt.Scanln(&chatMsg)
			if err != nil {
				return
			}
		}
		remote = ""
		fmt.Println(">>> 请输入发送对象，输入exit退出")
		_, err = fmt.Scanln(&remote)
		if err != nil {
			return
		}
	}
}

func (client *Client) PublicChat() {
	var chatMsg string
	fmt.Println(">>> 请输入发送内容，输入exit退出")
	_, err := fmt.Scanln(&chatMsg)
	if err != nil {
		return
	}
	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println(">>> conn write err:", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println(">>> 请输入发送内容，输入exit退出")
		_, err := fmt.Scanln(&chatMsg)
		if err != nil {
			return
		}
	}

}

func (client *Client) Run() {
	for client.flag != 0 {
		for !client.Menu() {
		}
		switch client.flag {
		case 1:
			client.PublicChat()
			break
		case 2:
			client.PrivateChat()
			break
		case 3:
			// 更新用户名
			client.updateUserName()
			break

		}
	}
}

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器的IP地址(默认127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器的端口(默认8888)")
}

func (client *Client) DealResponse() {
	_, err := io.Copy(os.Stdout, client.conn)
	if err != nil {
		return
	}
}

func main() {
	// 命令行解析
	flag.Parse()
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>> 链接服务器失败")
		return
	}
	go client.DealResponse()
	fmt.Println(">>> 链接服务器成功")
	client.Run()

}
