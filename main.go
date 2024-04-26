package main

import (
	"log"
	"net/http"
	"os"
	roomhandler "sprint-planner/room-handler"
	wshandler "sprint-planner/ws-handler"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	wshandler.Init()
	roomhandler.Init()

	log.Println("http server started on: " + os.Getenv("PORT"))
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
