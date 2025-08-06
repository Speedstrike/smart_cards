package main

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func connect() *redis.Client {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("no environment file found, or failed to load")
	}

	dbNo, err := strconv.Atoi(os.Getenv("REDIS_DB_NO"))
	if err != nil {
		log.Fatal("cannot convert the string to int!")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Username: os.Getenv("REDIS_USER"),
		Password: os.Getenv("REDIS_PASS"),
		DB:       dbNo,
	})

	return rdb
}
