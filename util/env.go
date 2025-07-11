package util

import (
	"github.com/joho/godotenv"
	"log"
)

func InitEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
}
