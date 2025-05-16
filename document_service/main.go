package main

import (
	"document_service/handlers"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	r := mux.NewRouter()

	// Buat direktori uploads jika belum ada
	if err := os.MkdirAll("uploads/icp", os.ModePerm); err != nil {
		log.Fatal(err)
	}

	// Route untuk upload ICP
	r.HandleFunc("/upload/icp", handlers.UploadICPHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/icp", handlers.GetICPHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/download", handlers.DownloadFileHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/icp/{id}", handlers.GetICPByIDHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/icp/edit", handlers.EditICPHandler).Methods("PUT", "OPTIONS")

	// Route untuk review ICP
	r.HandleFunc("/reviewicp", handlers.GetICPByDosenIDHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/reviewicp/list", handlers.GetReviewICPHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/reviewicp/dosen/list", handlers.GetReviewICPDosenHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/reviewicp/taruna/list", handlers.GetRevisiICPTarunaHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/updateicpstatus", handlers.UpdateICPStatusHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/upload/reviewicp/dosen", handlers.UploadDosenReviewICPHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/upload/revisiicp/taruna", handlers.UploadTarunaRevisiICPHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/upload/reviewicp", handlers.UploadReviewICPHandler).Methods("POST", "OPTIONS")

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Wrap router dengan CORS middleware
	handler := c.Handler(r)

	// Start server
	log.Println("Server started on :8082")
	log.Fatal(http.ListenAndServe(":8082", handler))
}
