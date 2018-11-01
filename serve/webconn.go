package main

import (
	"net/http"
	"github.com/gorilla/websocket"
	"log"
	"regexp"
)

const (
	PASSWORD = "123456"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type warehouse struct {
	conns      map[string]*connection //用id来区分   这个map不加锁   是因为用了chan不用考虑并发去操作这个map了
	register   chan *connection       //用来注册conn
	unregister chan string            // 用来注销
	broadcast  chan msg               //用来广播
}

type connection struct {
	conn *websocket.Conn
	send chan []byte
	name string
}

type msg struct {
	fromuser string
	touser   string
	msg      []byte
	mtype    int
}

var w = warehouse{
	conns:      make(map[string]*connection),
	register:   make(chan *connection),
	unregister: make(chan string),
	broadcast:  make(chan msg),
}

func (w *warehouse) run() {
	for {
		select {
		case c := <-w.register:
			//判断有没有这个连接
			if _, ok := w.conns[c.name]; ok {
				//如果存在就不让注册
				c.send <- []byte("serve : username is already exist")
				close(c.send)
				break
			}
			//注册这个连接
			w.conns[c.name] = c
		case name := <-w.unregister:
			//注销这个用户
			close(w.conns[name].send)
			delete(w.conns, name)
		case b := <-w.broadcast:
			log.Println(b)
			var message []byte
			if b.mtype == 1 {
				//系统消息
				message = []byte("serve :" + string(b.msg))
			} else if b.mtype == 2 {
				//私聊
				message = []byte(b.fromuser + " says " + string(b.msg))
			} else {
				message = b.msg
			}

			if b.touser == "all" {
				for _, v := range w.conns {
					v.send <- message
				}
			} else {
				if _, ok := w.conns[b.touser]; ok {
					w.conns[b.touser].send <- message
				} else {
					w.conns[b.fromuser].send <- []byte("the user is not online")
				}
			}
		}
	}
}

func connecttows(w http.ResponseWriter, r *http.Request) {
	conn, e := upgrader.Upgrade(w, r, nil)
	if e != nil {
		log.Println("upgrade err:" + e.Error())
		return
	}

	c := &connection{conn: conn, send: make(chan []byte)}

	//一个写携程
	go writeProcess(c)

	readProcess(c)
}

func writeProcess(c *connection) {
	defer func() {
		c.conn.Close()
	}()
	for msg := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}

func readProcess(c *connection) {
	defer func() {
		close(c.send)
	}()
	for {
		_, p, err := c.conn.ReadMessage()
		log.Println("readmessage:" + string(p))
		if err != nil {
			close(c.send)
		}

		m := msg{mtype: 1, touser: "all"} //mtype  1 系统消息  2 私聊   or  群发

		//分析数据
		compile := regexp.MustCompile(`name=(.*)&pwd=(.*)`)
		submatch := compile.FindSubmatch(p)

		if len(submatch) == 3 {
			if string(submatch[2]) != PASSWORD {
				c.send <- []byte("serve :password error")
				continue
			}
			//说明是登陆的
			c.name = string(submatch[1])
			//注册
			w.register <- c
			log.Println(c.name + " join this connect")
			p = []byte(c.name + " join")
		}

		compile = regexp.MustCompile(`@(.*)\s(.*)`)
		submatch = compile.FindSubmatch(p)

		if len(submatch) == 3 {
			//表明私聊
			m.touser = string(submatch[1])
			m.mtype = 2
			p = submatch[2]
		}

		m.fromuser = c.name
		m.msg = p

		w.broadcast <- m

	}
}
