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

	// Register all routes
	r.HandleFunc("/upload/icp", handlers.UploadICPHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/icp", handlers.GetICPHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/download", handlers.DownloadFileHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/icp/{id}", handlers.GetICPByIDHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/icp/edit", handlers.EditICPHandler).Methods("PUT", "OPTIONS")

	// Route untuk review ICP
	r.HandleFunc("/reviewicp", handlers.GetICPByDosenIDHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/reviewicp/list", handlers.GetReviewICPHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/reviewicp/dosen/list", handlers.GetReviewICPDosenHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/reviewicp/dosen/detail", handlers.GetReviewICPDosenDetailHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/reviewicp/taruna/list", handlers.GetRevisiICPTarunaHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/updateicpstatus", handlers.UpdateICPStatusHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/upload/reviewicp/dosen", handlers.UploadDosenReviewICPHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/upload/revisiicp/taruna", handlers.UploadTarunaRevisiICPHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/upload/reviewicp", handlers.UploadReviewICPHandler).Methods("POST", "OPTIONS")

	// Final ICP routes
	r.HandleFunc("/finalicp/upload", handlers.UploadFinalICPHandler)
	r.HandleFunc("/finalicp/list", handlers.GetFinalICPHandler)
	r.HandleFunc("/finalicp/all", handlers.GetAllFinalICPWithTarunaHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/finalicp/status", handlers.UpdateFinalICPStatusHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/finalicp/download/{id}", handlers.DownloadFinalICPHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/finalicp/taruna-topics", handlers.GetTarunaTopicsHandler).Methods("GET", "OPTIONS")

	// Hasil Telaah ICP routes
	r.HandleFunc("/hasiltelaah/upload", handlers.UploadHasilTelaahHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/hasiltelaah/taruna", handlers.GetHasilTelaahTarunaHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/hasiltelaah/monitoring", handlers.GetMonitoringTelaahHandler).Methods("GET", "OPTIONS")

	// Proposal routes
	r.HandleFunc("/upload/proposal", handlers.UploadProposalHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/proposal", handlers.GetProposalHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/proposal/{id}", handlers.GetProposalByIDHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/proposal/edit", handlers.EditProposalHandler).Methods("PUT", "OPTIONS")

	// Review Proposal routes
	r.HandleFunc("/reviewproposal", handlers.GetProposalByDosenIDHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/reviewproposal/dosen/list", handlers.GetReviewProposalDosenHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/reviewproposal/dosen/detail", handlers.GetReviewProposalDosenDetailHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/upload/reviewproposal/dosen", handlers.UploadDosenReviewProposalHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/updateproposalstatus", handlers.UpdateProposalStatusHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/upload/revisiproposal/taruna", handlers.UploadTarunaRevisiProposalHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/upload/reviewproposal", handlers.UploadReviewProposalHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/reviewproposal/taruna/list", handlers.GetRevisiProposalTarunaHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/reviewproposal/dosen/list", handlers.GetReviewProposalDosenHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/reviewproposal/dosen/detail", handlers.GetReviewProposalDosenDetailHandler).Methods("GET", "OPTIONS")

	// Final Proposal routes
	r.HandleFunc("/finalproposal/upload", handlers.UploadFinalProposalHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/finalproposal/list", handlers.GetFinalProposalHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/finalproposal/all", handlers.GetAllFinalProposalWithTarunaHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/finalproposal/status", handlers.UpdateFinalProposalStatusHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/finalproposal/download/{id}", handlers.DownloadFinalProposalHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/finalproposal/taruna-topics", handlers.GetTarunaTopicsHandler).Methods("GET", "OPTIONS")

	// Register seminar proposal routes
	r.HandleFunc("/upload/seminarproposal", handlers.UploadSeminarProposalHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/seminarproposal/list", handlers.GetSeminarProposalHandler).Methods("GET", "OPTIONS")

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
