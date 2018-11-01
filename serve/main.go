package main

import (
	"net/http"
	"flag"
	"github.com/gpmgo/gopm/modules/log"
	"text/template"
)

var addr = flag.String("addr", "8777", "the port of service")

func main() {

	flag.Parse()

	go w.run()
	http.HandleFunc("/ws", connecttows)
	http.HandleFunc("/", homepage)

	if err := http.ListenAndServe(":" + *addr, nil); err != nil {
		log.Fatal("serve err:" + err.Error())
	}

}

func homepage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Url path not found", http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	template.Must(template.ParseFiles("room.html")).Execute(w, r.Host)

}
