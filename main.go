package main

import (
	"net/http"
	"flag"
	"github.com/gpmgo/gopm/modules/log"
)

var addr = flag.String("addr", "8777", "the port of service")

func main() {

	flag.Parse()

	go w.run()
	http.HandleFunc("/ws", connecttows)

	if err := http.ListenAndServe(":" + *addr, nil); err != nil {
		log.Fatal("serve err:" + err.Error())
	}

}
