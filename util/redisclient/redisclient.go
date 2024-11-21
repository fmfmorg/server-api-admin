package redisclient

import (
	"context"
	"fmt"
	"log"
	"server-api-admin/config"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
	Ctx    = context.Background()
	// serverID string
)

func init() {
	// serverID = getServerID()

	Client = redis.NewClient(&redis.Options{
		Addr:         config.RedisAddress,  // Update this with your Redis server address
		Password:     config.RedisPassword, // Set this if your Redis server requires authentication
		DB:           0,                    // Use default DB
		PoolSize:     100,                  // Increased pool size to handle higher concurrency
		MinIdleConns: 10,                   // Increased minimum idle connections
		ReadTimeout:  30 * time.Second,     // Read timeout
		WriteTimeout: 30 * time.Second,     // Write timeout
		MaxRetries:   3,
	})

	// Ping Redis to check the connection
	_, err := Client.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	fmt.Println("Connected to Redis")
}

/*
func getServerID() string {
	// Implement logic to get the server's IP address
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("Error getting hostname: %v", err)
		return randomUUID4()
	}

	ips, err := net.LookupIP(hostname)
	if err != nil {
		log.Printf("Error getting IP address: %v", err)
		return randomUUID4()
	}

	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			return ipv4.String()
		}
	}

	return randomUUID4()
}

func randomUUID4() string {
	// Generate a UUIDv4
	uuid := uuid.New()
	return uuid.String()
}
*/
