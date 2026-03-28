package main

import (
	"fmt"
	"log"

	"mychat_server/internal/config"
	"mychat_server/internal/https_server"
	"mychat_server/internal/service/chat"
	gormservice "mychat_server/internal/service/gorm"
)

func main() {
	if err := gormservice.InitDB(); err != nil {
		log.Fatal(err)
	}
	chat.InitRuntime()
	defer chat.StopRuntime()

	cfg := config.GetConfig()
	addr := fmt.Sprintf("%s:%d", cfg.MainConfig.Host, cfg.MainConfig.Port)
	log.Printf("server start at http://%s", addr)

	if err := https_server.GE.Run(addr); err != nil {
		log.Fatal(err)
	}
}
