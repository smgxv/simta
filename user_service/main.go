package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"user_service/handlers"
	"user_service/middleware"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func setupRoutes() *mux.Router {
	r := mux.NewRouter()

	// Users routes
	r.HandleFunc("/users", middleware.AuthMiddleware(handlers.UserHandler)).Methods("GET", "POST", "OPTIONS")
	r.HandleFunc("/users/add", middleware.AuthMiddleware(handlers.AddUser)).Methods("POST", "OPTIONS")
	r.HandleFunc("/users/edit", middleware.AuthMiddleware(handlers.EditUser)).Methods("PUT", "OPTIONS")
	r.HandleFunc("/users/detail", middleware.AuthMiddleware(handlers.GetUserDetail)).Methods("GET", "OPTIONS")
	r.HandleFunc("/users/delete", middleware.AuthMiddleware(handlers.DeleteUser)).Methods("DELETE", "OPTIONS")

	// Dosen and Taruna routes
	r.HandleFunc("/dosen", middleware.AuthMiddleware(handlers.GetAllDosen)).Methods("GET", "OPTIONS")
	r.HandleFunc("/taruna", middleware.AuthMiddleware(handlers.GetAllTaruna)).Methods("GET", "OPTIONS")
	r.HandleFunc("/taruna/edituser", middleware.AuthMiddleware(handlers.EditUserTaruna)).Methods("PUT", "OPTIONS")
	r.HandleFunc("/dosen/edituser", middleware.AuthMiddleware(handlers.EditUserDosen)).Methods("PUT", "OPTIONS")
	r.HandleFunc("/taruna/topik", middleware.AuthMiddleware(handlers.GetTarunaWithTopik)).Methods("GET", "OPTIONS")

	// Proposal routes
	r.HandleFunc("/dosbing_proposal", middleware.AuthMiddleware(handlers.AssignDosbingProposal)).Methods("POST", "OPTIONS")
	r.HandleFunc("/penguji_proposal", middleware.AuthMiddleware(handlers.AssignPengujiProposal)).Methods("POST", "OPTIONS")
	r.HandleFunc("/final_proposal", middleware.AuthMiddleware(handlers.GetFinalProposalByTarunaIDHandler)).Methods("GET", "OPTIONS")

	// Dosen Dashboard routes
	r.HandleFunc("/dosen/dashboard", middleware.AuthMiddleware(handlers.DosenDashboardHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/dosen/dashboard/icp", middleware.AuthMiddleware(handlers.ICPDitelaahHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/dosen/dashboard/bimbingan", middleware.AuthMiddleware(handlers.GetBimbinganByDosenHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/dosen/dashboard/pengujianproposal", middleware.AuthMiddleware(handlers.GetPengujianProposalHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/dosen/dashboard/pengujianlaporan70", middleware.AuthMiddleware(handlers.GetPengujianLaporan70Handler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/dosen/dashboard/pengujianlaporan100", middleware.AuthMiddleware(handlers.GetPengujianLaporan100Handler)).Methods("GET", "OPTIONS")

	// TARUNA ROUTE
	r.HandleFunc("/taruna/dosbing", middleware.AuthMiddleware(handlers.GetTarunaWithDosbing)).Methods("GET", "OPTIONS")
	r.HandleFunc("/taruna/pengujiproposal", middleware.AuthMiddleware(handlers.GetTarunaWithPengujiProposal)).Methods("GET", "OPTIONS")
	r.HandleFunc("/taruna/dashboard", middleware.AuthMiddleware(handlers.TarunaDashboardHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/taruna/dashboard/icp", middleware.AuthMiddleware(handlers.TarunaDashboardHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/taruna/dashboard/dosen", middleware.AuthMiddleware(handlers.TarunaDashboardHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/taruna/penelaahicp", middleware.AuthMiddleware(handlers.GetTarunaWithPenelaahICP)).Methods("GET", "OPTIONS")

	// ICP routes
	r.HandleFunc("/final_icp", middleware.AuthMiddleware(handlers.GetFinalICPByTarunaIDHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/penelaah_icp", middleware.AuthMiddleware(handlers.AssignPenelaahICP)).Methods("POST", "OPTIONS")

	// Laporan 70 routes
	r.HandleFunc("/taruna/pengujilaporan70", middleware.AuthMiddleware(handlers.GetTarunaWithPengujiLaporan70)).Methods("GET", "OPTIONS")
	r.HandleFunc("/final_laporan70", middleware.AuthMiddleware(handlers.GetFinalLaporan70ByTarunaIDHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/penguji_laporan70", middleware.AuthMiddleware(handlers.AssignPengujiLaporan70)).Methods("POST", "OPTIONS")

	// Laporan 100 routes
	r.HandleFunc("/taruna/pengujilaporan100", middleware.AuthMiddleware(handlers.GetTarunaWithPengujiLaporan100)).Methods("GET", "OPTIONS")
	r.HandleFunc("/final_laporan100", middleware.AuthMiddleware(handlers.GetFinalLaporan100ByTarunaIDHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/penguji_laporan100", middleware.AuthMiddleware(handlers.AssignPengujiLaporan100)).Methods("POST", "OPTIONS")

	return r
}

func main() {
	router := setupRoutes()

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://104.43.89.154:8443"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

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

	// Wrap router with CORS middleware
	handler := c.Handler(router)

	// Redirect HTTP to HTTPS
	go func() {
		fmt.Println("HTTP API Server running on port 8081 (redirecting to HTTPS)...")
		err := http.ListenAndServe(":8081", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
		}))
		if err != nil {
			log.Fatal("HTTP Server Error: ", err)
		}
	}()

	// Start HTTPS server
	fmt.Println("HTTPS API Server running on port 8444...")
	log.Fatal(http.ListenAndServeTLS(":8444", "cert/server.crt", "cert/server.key", handler))
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0600)
}
