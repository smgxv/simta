package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"ta_service/controllers"
	"ta_service/handlers"
	"ta_service/middleware"

	"github.com/gorilla/mux"
)

// Tambahkan middleware CORS
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func setupRoutes() *mux.Router {
	r := mux.NewRouter()

	// Tambahkan middleware CORS ke router
	r.Use(corsMiddleware)

	// Menyajikan file statis dari direktori style
	r.PathPrefix("/style/images/").Handler(http.StripPrefix("/style/images/", http.FileServer(http.Dir("static/style/images"))))
	r.PathPrefix("/style/css/").Handler(http.StripPrefix("/style/css/", http.FileServer(http.Dir("static/style/css"))))
	r.PathPrefix("/style/fonts/").Handler(http.StripPrefix("/style/fonts/", http.FileServer(http.Dir("static/style/fonts"))))
	r.PathPrefix("/style/js/").Handler(http.StripPrefix("/style/js/", http.FileServer(http.Dir("static/style/js"))))
	r.PathPrefix("/style/includes/").Handler(http.StripPrefix("/style/includes/", http.FileServer(http.Dir("static/style/includes"))))
	r.PathPrefix("/style/vendor/").Handler(http.StripPrefix("/style/vendor/", http.FileServer(http.Dir("static/style/vendor"))))

	// Menyajikan file statis dari direktori admin/src
	r.PathPrefix("/admin/src/fonts/").Handler(http.StripPrefix("/admin/src/fonts/", http.FileServer(http.Dir("static/admin/src/fonts"))))
	r.PathPrefix("/admin/src/images/").Handler(http.StripPrefix("/admin/src/images/", http.FileServer(http.Dir("static/admin/src/images"))))
	r.PathPrefix("/admin/src/plugins/").Handler(http.StripPrefix("/admin/src/plugins/", http.FileServer(http.Dir("static/admin/src/plugins"))))
	r.PathPrefix("/admin/src/scripts/").Handler(http.StripPrefix("/admin/src/scripts/", http.FileServer(http.Dir("static/admin/src/scripts"))))
	r.PathPrefix("/admin/src/styles/").Handler(http.StripPrefix("/admin/src/styles/", http.FileServer(http.Dir("static/admin/src/styles"))))

	// Menyajikan file statis dari direktori admin/vendors
	r.PathPrefix("/admin/vendors/fonts/").Handler(http.StripPrefix("/admin/vendors/fonts/", http.FileServer(http.Dir("static/admin/vendors/fonts"))))
	r.PathPrefix("/admin/vendors/images/").Handler(http.StripPrefix("/admin/vendors/images/", http.FileServer(http.Dir("static/admin/vendors/images"))))
	r.PathPrefix("/admin/vendors/scripts/").Handler(http.StripPrefix("/admin/vendors/scripts/", http.FileServer(http.Dir("static/admin/vendors/scripts"))))
	r.PathPrefix("/admin/vendors/styles/").Handler(http.StripPrefix("/admin/vendors/styles/", http.FileServer(http.Dir("static/admin/vendors/styles"))))

	// Menyajikan file statis dari direktori taruna/src
	r.PathPrefix("/taruna/src/fonts/").Handler(http.StripPrefix("/taruna/src/fonts/", http.FileServer(http.Dir("static/taruna/src/fonts"))))
	r.PathPrefix("/taruna/src/images/").Handler(http.StripPrefix("/taruna/src/images/", http.FileServer(http.Dir("static/taruna/src/images"))))
	r.PathPrefix("/taruna/src/plugins/").Handler(http.StripPrefix("/taruna/src/plugins/", http.FileServer(http.Dir("static/taruna/src/plugins"))))
	r.PathPrefix("/taruna/src/scripts/").Handler(http.StripPrefix("/taruna/src/scripts/", http.FileServer(http.Dir("static/taruna/src/scripts"))))
	r.PathPrefix("/taruna/src/styles/").Handler(http.StripPrefix("/taruna/src/styles/", http.FileServer(http.Dir("static/taruna/src/styles"))))

	// Menyajikan file statis dari direktori taruna/vendors
	r.PathPrefix("/taruna/vendors/fonts/").Handler(http.StripPrefix("/taruna/vendors/fonts/", http.FileServer(http.Dir("static/taruna/vendors/fonts"))))
	r.PathPrefix("/taruna/vendors/images/").Handler(http.StripPrefix("/taruna/vendors/images/", http.FileServer(http.Dir("static/taruna/vendors/images"))))
	r.PathPrefix("/taruna/vendors/scripts/").Handler(http.StripPrefix("/taruna/vendors/scripts/", http.FileServer(http.Dir("static/taruna/vendors/scripts"))))
	r.PathPrefix("/taruna/vendors/styles/").Handler(http.StripPrefix("/taruna/vendors/styles/", http.FileServer(http.Dir("static/taruna/vendors/styles"))))

	// Menyajikan file statis dari direktori dosen/src
	r.PathPrefix("/dosen/src/fonts/").Handler(http.StripPrefix("/dosen/src/fonts/", http.FileServer(http.Dir("static/dosen/src/fonts"))))
	r.PathPrefix("/dosen/src/images/").Handler(http.StripPrefix("/dosen/src/images/", http.FileServer(http.Dir("static/dosen/src/images"))))
	r.PathPrefix("/dosen/src/plugins/").Handler(http.StripPrefix("/dosen/src/plugins/", http.FileServer(http.Dir("static/dosen/src/plugins"))))
	r.PathPrefix("/dosen/src/scripts/").Handler(http.StripPrefix("/dosen/src/scripts/", http.FileServer(http.Dir("static/dosen/src/scripts"))))
	r.PathPrefix("/dosen/src/styles/").Handler(http.StripPrefix("/dosen/src/styles/", http.FileServer(http.Dir("static/dosen/src/styles"))))

	// Menyajikan file statis dari direktori dosen/vendors
	r.PathPrefix("/dosen/vendors/fonts/").Handler(http.StripPrefix("/dosen/vendors/fonts/", http.FileServer(http.Dir("static/dosen/vendors/fonts"))))
	r.PathPrefix("/dosen/vendors/images/").Handler(http.StripPrefix("/dosen/vendors/images/", http.FileServer(http.Dir("static/dosen/vendors/images"))))
	r.PathPrefix("/dosen/vendors/scripts/").Handler(http.StripPrefix("/dosen/vendors/scripts/", http.FileServer(http.Dir("static/dosen/vendors/scripts"))))
	r.PathPrefix("/dosen/vendors/styles/").Handler(http.StripPrefix("/dosen/vendors/styles/", http.FileServer(http.Dir("static/dosen/vendors/styles"))))

	// Public routes (tanpa middleware)
	r.HandleFunc("/loginusers", controllers.LoginUsers).Methods("GET")
	r.HandleFunc("/login", handlers.LoginHandler).Methods("POST", "OPTIONS")
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
	router := setupRoutes()

	// Create cert directory if it doesn't exist
	if err := os.MkdirAll("cert", os.ModePerm); err != nil {
		log.Fatal(err)
	}

	// Redirect HTTP to HTTPS
	go func() {
		log.Println("HTTP Service running on port 8080 (redirecting to HTTPS)")
		err := http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			target := "https://" + r.Host + r.URL.Path
			if len(r.URL.RawQuery) > 0 {
				target += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, target, http.StatusMovedPermanently)
		}))
		if err != nil {
			log.Fatal("HTTP Server Error: ", err)
		}
	}()

	// Start HTTPS server
	log.Println("HTTPS Service running on port 8443")
	server := &http.Server{
		Addr:    ":8443",
		Handler: router,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			},
		},
	}
	log.Fatal(server.ListenAndServeTLS("cert/server.crt", "cert/server.key"))
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0600)
}
