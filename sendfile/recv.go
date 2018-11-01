package main

import (
	"net"
	"fmt"
	"regexp"
	"io"
	"os"
)

func main() {

	listener, e := net.Listen("tcp", ":8887")
	if e != nil {
		fmt.Println("err=" + e.Error())
	}

	defer listener.Close()

	for {
		conn, i := listener.Accept()
		if i != nil {
			fmt.Println("err =" + i.Error())
			continue
		}

		go handler(conn)
	}

}

func handler(conn net.Conn) {

	//取addr
	addr := conn.RemoteAddr().String()

	defer conn.Close()

	//读
	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)

		if err == io.EOF {
			fmt.Println("导入完毕")
			return
		}
		if err != nil {
			fmt.Printf("[%s] err =\n", addr, err.Error())
			return
		}

		//取出标题
		compile := regexp.MustCompile(`@title=(.*)`)

		submatch := compile.FindSubmatch(buf[:n])

		if len(submatch) != 2 {

			fmt.Printf("title is wrong len=%d \n", len(submatch))
			fmt.Println(submatch)
			return
		}

		path := submatch[1]

		compile = regexp.MustCompile(`.*/(.*\..*)`)

		submatch = compile.FindSubmatch(path)

		if len(submatch) != 2 {
			fmt.Printf("title is wrong len=%d \n", len(submatch))
			return
		}

		title := string(submatch[1])

		conn.Write([]byte("send"))

		//有了标题   创建文件
		file, e := os.OpenFile(title, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if e != nil {
			fmt.Println("create file err =" + e.Error())
		}
		defer file.Close()

		readBuf := make([]byte, 1024)
		for {
			i, err2 := conn.Read(readBuf)
			if err2 == io.EOF {
				fmt.Println("传输完毕")
				return
			}
			if err2 != nil {
				fmt.Println("read msg err =" + err2.Error())
				return
			}

			_, err3 := file.Write(readBuf[:i])
			if err3 != nil{
				fmt.Println("file write err ="+err3.Error())
				return
			}

		}

	}

}
