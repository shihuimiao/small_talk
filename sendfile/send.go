package main

import (
	"os"
	"fmt"
	"net"
	"io"
)

func main() {

	var input []string = os.Args

	if len(input) < 2 {
		fmt.Println("必须带上参数")
		return
	}

	path := input[1]

	//检查有没有该文件
	file, e := os.Open(path)

	if e != nil {
		fmt.Println("err=" + e.Error())
		return
	}
	defer file.Close()

	//开始传输文件
	conn, i := net.Dial("tcp", "172.16.0.53:8887")
	if i != nil {
		fmt.Println("conn err =" + i.Error())
		return
	}

	defer conn.Close()

	//先把名字传过去
	conn.Write([]byte("@title=" + file.Name()))

	firstBuf := make([]byte, 100)
	n, err := conn.Read(firstBuf)
	if err != nil {
		fmt.Println("first read err=" + err.Error())
		return
	}

	if string(firstBuf[:n]) != "send" {
		fmt.Println("get err msg = " + string(firstBuf[:n]))
		return
	}

	buf := make([]byte, 1024)

	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			fmt.Println("传输完毕")
			return
		}
		if err != nil {
			fmt.Println("read file err=" + err.Error())
			return
		}
		conn.Write(buf[:n])

	}

}
