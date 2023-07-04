package main

import (
	"log"
	"os"
	"time"

	"github.com/dro14/profi-bot/processor/telegram"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

func main() {

	time.Local, _ = time.LoadLocation("Asia/Tashkent")
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	gin.SetMode(gin.ReleaseMode)

	file, err := os.Create("gin.log")
	if err != nil {
		log.Fatalf("can't create gin.log: %v", err)
	}
	gin.DefaultWriter = file

	r := gin.Default()
	r.POST("/main", telegram.ProcessUpdate)

	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}

	telegram.Init()
	log.Printf("listening on port %s", port)
	err = r.Run(":" + port)
	if err != nil {
		log.Fatalf("can't run server: %v", err)
	}
}
