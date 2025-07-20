package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"strings"
	"ta_service/controllers"
	"ta_service/handlers"
	"ta_service/middleware"
	"time"

	"github.com/gorilla/mux"
)

// Tambahkan middleware CORS
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("CORS: Processing request from %s", r.RemoteAddr)

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			log.Printf("CORS: Handling preflight request from %s", r.RemoteAddr)
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Request started: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		next.ServeHTTP(w, r)

		log.Printf("Request completed: %s %s - took %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func setupRoutes() *mux.Router {
	r := mux.NewRouter()

	// Add logging middleware
	r.Use(loggingMiddleware)
	r.Use(corsMiddleware)

	// Menyajikan file statis dari direktori style
	r.PathPrefix("/style/").Handler(http.StripPrefix("/style/", http.FileServer(http.Dir("static/style"))))
	r.PathPrefix("/admin/src/").Handler(http.StripPrefix("/admin/src/", http.FileServer(http.Dir("static/admin/src"))))
	r.PathPrefix("/admin/vendors/").Handler(http.StripPrefix("/admin/vendors/", http.FileServer(http.Dir("static/admin/vendors"))))
	r.PathPrefix("/taruna/src/").Handler(http.StripPrefix("/taruna/src/", http.FileServer(http.Dir("static/taruna/src"))))
	r.PathPrefix("/taruna/vendors/").Handler(http.StripPrefix("/taruna/vendors/", http.FileServer(http.Dir("static/taruna/vendors"))))
	r.PathPrefix("/dosen/src/").Handler(http.StripPrefix("/dosen/src/", http.FileServer(http.Dir("static/dosen/src"))))
	r.PathPrefix("/dosen/vendors/").Handler(http.StripPrefix("/dosen/vendors/", http.FileServer(http.Dir("static/dosen/vendors"))))

	// Public routes (tanpa middleware)
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/login.html")
	}).Methods("GET")

	r.HandleFunc("/login", handlers.LoginHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/loginusers", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/login.html")
	}).Methods("GET")
	r.HandleFunc("/logout", handlers.LogoutHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/dashboard", controllers.Index).Methods("GET")

	// Routes dengan middleware
	adminRouter := r.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middleware.RoleRedirectMiddleware)

	// Perbaiki routing untuk admin dashboard
	adminRouter.HandleFunc("/dashboard", controllers.AdminDashboard).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/calendar", controllers.Calendar).Methods("GET")
	adminRouter.HandleFunc("/listuser", controllers.ListUser).Methods("GET")
	adminRouter.HandleFunc("/adduser", controllers.AddUser).Methods("GET", "POST")
	adminRouter.HandleFunc("/profile", controllers.Profile).Methods("GET")
	adminRouter.HandleFunc("/edituser", controllers.EditUser).Methods("GET")
	adminRouter.HandleFunc("/deleteuser", controllers.DeleteUser).Methods("GET", "POST")
	adminRouter.HandleFunc("/listdosen", controllers.ListDosen).Methods("GET")
	adminRouter.HandleFunc("/listicp", controllers.ListICP).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/penelaah_icp", controllers.ListPenelaahICP).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/list_icp", controllers.ListICP).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/listproposal", controllers.ListProposal).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/detail_berkas_seminar_proposal", controllers.DetailBerkasProposal).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/detail_telaah_icp", controllers.DetailTelaahICP).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/dosbing_proposal", controllers.ListPembimbingProposal).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/penguji_proposal", controllers.ListPengujiProposal).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/penguji_laporan70", controllers.ListPengujiLaporan70).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/listlaporan70", controllers.ListLaporan70).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/detail_berkas_seminar_laporan70", controllers.DetailBerkasLaporan70).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/penguji_laporan100", controllers.ListPengujiLaporan100).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/listlaporan100", controllers.ListLaporan100).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/detail_berkas_seminar_laporan100", controllers.DetailBerkasLaporan100).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/repositori", controllers.Repositori).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/detail_berkas_tugas_akhir", controllers.DetailTugasAkhir).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/notification", controllers.Notification).Methods("GET", "POST")

	// Tambahkan routes untuk taruna
	tarunaRoutes := r.PathPrefix("/taruna").Subrouter()
	tarunaRoutes.Use(middleware.RoleRedirectMiddleware)

	// Route dashboard taruna
	tarunaRoutes.HandleFunc("/dashboard", controllers.TarunaDashboard).Methods("GET", "OPTIONS")
	tarunaRoutes.HandleFunc("/icp", controllers.ICP).Methods("GET", "OPTIONS")
	tarunaRoutes.HandleFunc("/editicp", controllers.EditICP).Methods("GET", "OPTIONS")
	tarunaRoutes.HandleFunc("/viewicp", controllers.ViewICPTaruna).Methods("GET", "OPTIONS")
	tarunaRoutes.HandleFunc("/profile", controllers.ProfileTaruna).Methods("GET")
	tarunaRoutes.HandleFunc("/editprofile", controllers.EditProfileTaruna).Methods("GET")
	tarunaRoutes.HandleFunc("/proposal", controllers.Proposal).Methods("GET", "OPTIONS")
	tarunaRoutes.HandleFunc("/ta70", controllers.Laporan70).Methods("GET", "OPTIONS")
	tarunaRoutes.HandleFunc("/ta100", controllers.Laporan100).Methods("GET", "OPTIONS")
	tarunaRoutes.HandleFunc("/detailinformasitaruna", controllers.DetailInformasiTaruna).Methods("GET", "OPTIONS")

	// Tambahkan routes untuk dosen
	dosenRoutes := r.PathPrefix("/dosen").Subrouter()
	dosenRoutes.Use(middleware.RoleRedirectMiddleware)

	// Route dashboard dosen
	dosenRoutes.HandleFunc("/dashboard", controllers.DosenDashboard).Methods("GET", "OPTIONS")
	dosenRoutes.HandleFunc("/bimbingan_icp", controllers.ReviewICP).Methods("GET", "OPTIONS")
	dosenRoutes.HandleFunc("/pengujian_icp", controllers.PengujiICP).Methods("GET", "OPTIONS")
	dosenRoutes.HandleFunc("/profile", controllers.ProfileDosen).Methods("GET", "OPTIONS")
	dosenRoutes.HandleFunc("/editprofile", controllers.EditProfileDosen).Methods("GET", "OPTIONS")
	dosenRoutes.HandleFunc("/viewicp", controllers.ViewICPDosen).Methods("GET", "OPTIONS")
	dosenRoutes.HandleFunc("/viewicp_review", controllers.ViewICPReviewDosen).Methods("GET", "OPTIONS")
	dosenRoutes.HandleFunc("/viewicp_revisi", controllers.ViewICPRevisiDosen).Methods("GET", "OPTIONS")
	dosenRoutes.HandleFunc("/bimbingan_proposal", controllers.BimbinganProposal).Methods("GET", "OPTIONS")
	dosenRoutes.HandleFunc("/pengujian_proposal", controllers.PengujiProposal).Methods("GET", "OPTIONS")
	dosenRoutes.HandleFunc("/bimbingan_laporan70", controllers.BimbinganLaporan70).Methods("GET", "OPTIONS")
	dosenRoutes.HandleFunc("/pengujian_laporan70", controllers.PengujiLaporan70).Methods("GET", "OPTIONS")
	dosenRoutes.HandleFunc("/bimbingan_laporan100", controllers.BimbinganLaporan100).Methods("GET", "OPTIONS")
	dosenRoutes.HandleFunc("/pengujian_laporan100", controllers.PengujiLaporan100).Methods("GET", "OPTIONS")
	dosenRoutes.HandleFunc("/detailinformasidosen", controllers.DetailInformasiDosen).Methods("GET", "OPTIONS")

	// Default redirect ke login
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/loginusers", http.StatusSeeOther)
	}).Methods("GET")

	return r
}

func main() {
	// Initialize router
	r := setupRoutes()

	// Configure TLS
	tlsConfig := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}

	// Create HTTPS server
	httpsServer := &http.Server{
		Addr:         ":8443",
		Handler:      r,
		TLSConfig:    tlsConfig,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Create HTTP server (for redirect)
	httpServer := &http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host := strings.Split(r.Host, ":")[0]
			target := "https://" + host + ":8443" + r.URL.Path
			if len(r.URL.RawQuery) > 0 {
				target += "?" + r.URL.RawQuery
			}
			log.Printf("HTTP request received: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
			log.Printf("Redirecting to: %s", target)
			http.Redirect(w, r, target, http.StatusMovedPermanently)
		}),
	}

	// Check SSL certificates
	log.Println("Checking SSL certificates...")
	certFile := "cert/server.crt"
	keyFile := "cert/server.key"

	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		log.Fatalf("Certificate file not found: %s", certFile)
	}
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		log.Fatalf("Key file not found: %s", keyFile)
	}

	// Start HTTP server (for redirect)
	go func() {
		log.Printf("HTTP Service running on port 8080 (redirecting to HTTPS)")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Start HTTPS server
	log.Printf("Starting server with SSL certificates in: cert/")
	log.Printf("HTTPS Service running on port 8443")
	if err := httpsServer.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTPS server error: %v", err)
	}
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0600)
}
