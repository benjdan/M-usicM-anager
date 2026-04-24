package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"M-usicM-anager/internal/api"
	"M-usicM-anager/internal/db"
	"M-usicM-anager/internal/library"
	"M-usicM-anager/internal/metadata"
	"M-usicM-anager/internal/slskd"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	fmt.Println("M-usicM-anager starting...")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "dbuser-test"),
		getEnv("DB_PASSWORD", "dbuserpass-test"),
		getEnv("DB_NAME", "musicmanager"),
	)

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	slskdClient := slskd.NewClient(
		getEnv("SLSKD_URL", "http://localhost:5030"),
		getEnv("SLSKD_USERNAME", "dbuser-test"),
		getEnv("SLSKD_PASSWORD", "dbuserpass-test"),
	)
	organizer := library.NewOrganizer(getEnv("MUSIC_DIR", "/mnt/music"))
	downloader := library.NewDownloader(database, slskdClient, organizer)

	mb := metadata.NewMusicBrainzClient()
	fanart := metadata.NewFanartClient(getEnv("FANART_API_KEY", ""))

	monitor := library.NewMonitor(database, mb, downloader, 24*time.Hour)
	monitor.Start()

	h := api.NewHandler(database, mb, fanart, downloader)

	r := gin.Default()

	r.SetTrustedProxies([]string{"127.0.0.1"})

	h.RegisterRoutes(r)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"app":    "M-usicM-anager",
		})
	})

	port := getEnv("PORT", "8080")
	log.Printf("M-usicM-anager listening on :%s", port)
	log.Fatal(r.Run(":" + port))
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
