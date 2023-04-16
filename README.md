> 编译服务端

```bash
# 根据运行平台 修改-o 参数的后缀名
go build -o server.exe main.go Server.go user.go
```



> 编译客户端

```bash
go build -o client.exe client.go
```



> 运行

```bash
# 运行服务端
server.exe

# 运行客户端
client.exe -i 127.0.0.1 -p 8888
```

