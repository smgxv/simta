package main

import (
	"fmt"
	"log"
	"net/http"
	"user_service/handlers"
	"user_service/middleware"
)

// ... existing code ...
func main() {
	http.HandleFunc("/users", middleware.AuthMiddleware(handlers.UserHandler))
	http.HandleFunc("/users/add", middleware.AuthMiddleware(handlers.AddUser))
	http.HandleFunc("/users/edit", middleware.AuthMiddleware(handlers.EditUser))
	http.HandleFunc("/users/detail", middleware.AuthMiddleware(handlers.GetUserDetail))
	http.HandleFunc("/users/delete", middleware.AuthMiddleware(handlers.DeleteUser))

	http.HandleFunc("/dosen", middleware.AuthMiddleware(handlers.GetAllDosen))
	http.HandleFunc("/taruna", middleware.AuthMiddleware(handlers.GetAllTaruna))
	http.HandleFunc("/taruna/edituser", middleware.AuthMiddleware(handlers.EditUserTaruna))
	http.HandleFunc("/dosen/edituser", middleware.AuthMiddleware(handlers.EditUserDosen))
	http.HandleFunc("/taruna/topik", middleware.AuthMiddleware(handlers.GetTarunaWithTopik))

	http.HandleFunc("/dosbing_proposal", middleware.AuthMiddleware(handlers.AssignDosbingProposal))

	http.HandleFunc("/taruna/dosbing", middleware.AuthMiddleware(handlers.GetTarunaWithDosbing))

	fmt.Println("API Server running on port 8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
