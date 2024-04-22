package main

import (
	"log"
	"net/http"
	"os"
	"sprint-planner/api"
	wshandler "sprint-planner/ws-handler"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	wshandler.Init()
	api.InitRoutes()

	log.Println("http server started on: " + os.Getenv("PORT"))
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
