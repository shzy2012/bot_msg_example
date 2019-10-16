package main

import (
	"bot_msg_example/service"
	"flag"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":9010", "http service address")

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("content-type", "application/json")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func main() {

	flag.Parse()

	//吾来平台 pubkey, secret
	pubkey, secret := "84Mb38jJR6Bg7qB3MT09cSHPdrEImIgv00362b370af3cc02eb", "x9xBAnz8JHrI8GdsdVoJ" //os.Getenv("pubkey"), os.Getenv("secret")
	hub := service.NewHub(pubkey, secret)
	go hub.Run()

	//设置http路由
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		service.ServeWs(hub, w, r)
	})

	//设置bot消息路由/消息投递
	http.HandleFunc("/bot/", func(w http.ResponseWriter, r *http.Request) {
		service.ServeBotMsg(hub, w, r)
	})

	//开启http服务
	log.Printf("listen on: %s\n", *addr)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
