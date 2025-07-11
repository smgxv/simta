package main

import (
	"fmt"
	"net/http"
	"notification_service/handlers"
)

func main() {
	http.HandleFunc("/broadcast", handlers.BroadcastNotification)

	fmt.Println("✅ Notification Service running on port 8083")
	http.ListenAndServe(":8083", nil)
}
