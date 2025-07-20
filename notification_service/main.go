package main

import (
	"log"
	"net/http"
	"os"

	"notification_service/handlers"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	r := mux.NewRouter()

	// Buat folder uploads kalau belum ada
	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		log.Fatal("Gagal membuat folder uploads:", err)
	}

	// Register endpoint
	r.HandleFunc("/broadcast", handlers.BroadcastNotification).Methods("POST", "OPTIONS")
	r.HandleFunc("/notifications", handlers.GetNotifications).Methods("GET", "OPTIONS")
	r.HandleFunc("/notification/{id}", handlers.GetNotificationByID).Methods("GET", "OPTIONS")
	r.HandleFunc("/download/{filename}", handlers.DownloadFile).Methods("GET", "OPTIONS")

	// Setup CORS agar frontend (port 8443) bisa akses
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://104.43.89.154:8443"}, // Updated to HTTPS and new port
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Bungkus router dengan CORS handler
	handler := c.Handler(r)

	// Create cert directory if it doesn't exist
	if err := os.MkdirAll("cert", os.ModePerm); err != nil {
		log.Fatal(err)
	}

	// Copy certificates from ta_service if they don't exist
	if _, err := os.Stat("cert/server.crt"); os.IsNotExist(err) {
		if err := copyFile("../ta_service/cert/server.crt", "cert/server.crt"); err != nil {
			log.Fatal(err)
		}
	}
	if _, err := os.Stat("cert/server.key"); os.IsNotExist(err) {
		if err := copyFile("../ta_service/cert/server.key", "cert/server.key"); err != nil {
			log.Fatal(err)
		}
	}

	// Redirect HTTP to HTTPS
	go func() {
		log.Println("HTTP Service running on port 8083 (redirecting to HTTPS)")
		err := http.ListenAndServe(":8083", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
		}))
		if err != nil {
			log.Fatal("HTTP Server Error: ", err)
		}
	}()

	// Start HTTPS server
	log.Println("✅ HTTPS Notification Service running on :8446")
	log.Fatal(http.ListenAndServeTLS(":8446", "cert/server.crt", "cert/server.key", handler))
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0600)
}
