package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/HonLaderDev/HonLaderAuth-client/client"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	authClient := client.HonLaderAuthClientConfig{
		BaseURL: envOr("HONLADER_AUTH_URL", "http://127.0.0.1:8080"),
		APIKey:  envOr("HONLADER_AUTH_APIKEY", "test"),
	}.New()

	nemcClient, err := authClient.NEMCClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	detail, err := nemcClient.UserDetail().GetPEUserDetail()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Name:", detail.Name)
	fmt.Println("UserID:", detail.UserID)
}

func envOr(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
