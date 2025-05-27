package main

import (
	"log"
	"net/http"
	"ta_service/controllers"
	"ta_service/handlers"
	"ta_service/middleware"

	"github.com/gorilla/mux"
)

// Tambahkan middleware CORS
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// Menyajikan file statis dari direktori style
	http.Handle("/style/images/", http.StripPrefix("/style/images/", http.FileServer(http.Dir("static/style/images"))))
	http.Handle("/style/css/", http.StripPrefix("/style/css/", http.FileServer(http.Dir("static/style/css"))))
	http.Handle("/style/fonts/", http.StripPrefix("/style/fonts/", http.FileServer(http.Dir("static/style/fonts"))))
	http.Handle("/style/js/", http.StripPrefix("/style/js/", http.FileServer(http.Dir("static/style/js"))))
	http.Handle("/style/includes/", http.StripPrefix("/style/includes/", http.FileServer(http.Dir("static/style/includes"))))
	http.Handle("/style/vendor/", http.StripPrefix("/style/vendor/", http.FileServer(http.Dir("static/style/vendor"))))

	// Menyajikan file statis dari direktori admin/src
	http.Handle("/admin/src/fonts/", http.StripPrefix("/admin/src/fonts/", http.FileServer(http.Dir("static/admin/src/fonts"))))
	http.Handle("/admin/src/images/", http.StripPrefix("/admin/src/images/", http.FileServer(http.Dir("static/admin/src/images"))))
	http.Handle("/admin/src/plugins/", http.StripPrefix("/admin/src/plugins/", http.FileServer(http.Dir("static/admin/src/plugins"))))
	http.Handle("/admin/src/scripts/", http.StripPrefix("/admin/src/scripts/", http.FileServer(http.Dir("static/admin/src/scripts"))))
	http.Handle("/admin/src/styles/", http.StripPrefix("/admin/src/styles/", http.FileServer(http.Dir("static/admin/src/styles"))))

	// Menyajikan file statis dari direktori admin/vendors
	http.Handle("/admin/vendors/fonts/", http.StripPrefix("/admin/vendors/fonts/", http.FileServer(http.Dir("static/admin/vendors/fonts"))))
	http.Handle("/admin/vendors/images/", http.StripPrefix("/admin/vendors/images/", http.FileServer(http.Dir("static/admin/vendors/images"))))
	http.Handle("/admin/vendors/scripts/", http.StripPrefix("/admin/vendors/scripts/", http.FileServer(http.Dir("static/admin/vendors/scripts"))))
	http.Handle("/admin/vendors/styles/", http.StripPrefix("/admin/vendors/styles/", http.FileServer(http.Dir("static/admin/vendors/styles"))))

	// Menyajikan file statis dari direktori taruna/src
	http.Handle("/taruna/src/fonts/", http.StripPrefix("/taruna/src/fonts/", http.FileServer(http.Dir("static/taruna/src/fonts"))))
	http.Handle("/taruna/src/images/", http.StripPrefix("/taruna/src/images/", http.FileServer(http.Dir("static/taruna/src/images"))))
	http.Handle("/taruna/src/plugins/", http.StripPrefix("/taruna/src/plugins/", http.FileServer(http.Dir("static/taruna/src/plugins"))))
	http.Handle("/taruna/src/scripts/", http.StripPrefix("/taruna/src/scripts/", http.FileServer(http.Dir("static/taruna/src/scripts"))))
	http.Handle("/taruna/src/styles/", http.StripPrefix("/taruna/src/styles/", http.FileServer(http.Dir("static/taruna/src/styles"))))

	// Menyajikan file statis dari direktori taruna/vendors
	http.Handle("/taruna/vendors/fonts/", http.StripPrefix("/taruna/vendors/fonts/", http.FileServer(http.Dir("static/taruna/vendors/fonts"))))
	http.Handle("/taruna/vendors/images/", http.StripPrefix("/taruna/vendors/images/", http.FileServer(http.Dir("static/taruna/vendors/images"))))
	http.Handle("/taruna/vendors/scripts/", http.StripPrefix("/taruna/vendors/scripts/", http.FileServer(http.Dir("static/taruna/vendors/scripts"))))
	http.Handle("/taruna/vendors/styles/", http.StripPrefix("/taruna/vendors/styles/", http.FileServer(http.Dir("static/taruna/vendors/styles"))))

	// Menyajikan file statis dari direktori dosen/src
	http.Handle("/dosen/src/fonts/", http.StripPrefix("/dosen/src/fonts/", http.FileServer(http.Dir("static/dosen/src/fonts"))))
	http.Handle("/dosen/src/images/", http.StripPrefix("/dosen/src/images/", http.FileServer(http.Dir("static/dosen/src/images"))))
	http.Handle("/dosen/src/plugins/", http.StripPrefix("/dosen/src/plugins/", http.FileServer(http.Dir("static/dosen/src/plugins"))))
	http.Handle("/dosen/src/scripts/", http.StripPrefix("/dosen/src/scripts/", http.FileServer(http.Dir("static/dosen/src/scripts"))))
	http.Handle("/dosen/src/styles/", http.StripPrefix("/dosen/src/styles/", http.FileServer(http.Dir("static/dosen/src/styles"))))

	// Menyajikan file statis dari direktori dosen/vendors
	http.Handle("/dosen/vendors/fonts/", http.StripPrefix("/dosen/vendors/fonts/", http.FileServer(http.Dir("static/dosen/vendors/fonts"))))
	http.Handle("/dosen/vendors/images/", http.StripPrefix("/dosen/vendors/images/", http.FileServer(http.Dir("static/dosen/vendors/images"))))
	http.Handle("/dosen/vendors/scripts/", http.StripPrefix("/dosen/vendors/scripts/", http.FileServer(http.Dir("static/dosen/vendors/scripts"))))
	http.Handle("/dosen/vendors/styles/", http.StripPrefix("/dosen/vendors/styles/", http.FileServer(http.Dir("static/dosen/vendors/styles"))))

	// Membuat router baru
	router := mux.NewRouter()

	// Tambahkan middleware CORS ke router
	router.Use(corsMiddleware)

	// // API endpoints
	// http.HandleFunc("/login", handlers.LoginHandler)
	// http.HandleFunc("/refresh-token", handlers.RefreshTokenHandler)
	// router.HandleFunc("/logout", handlers.LogoutHandler).Methods("GET", "POST", "OPTIONS")

	// Public routes (tanpa middleware)
	router.HandleFunc("/loginusers", controllers.LoginUsers).Methods("GET")
	router.HandleFunc("/login", handlers.LoginHandler).Methods("POST", "OPTIONS")
	router.HandleFunc("/logout", handlers.LogoutHandler).Methods("POST", "OPTIONS")

	// Web endpoints
	http.HandleFunc("/dashboard", controllers.Index)
	// http.HandleFunc("/loginusers", controllers.LoginUsers)

	// Routes dengan middleware
	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middleware.RoleRedirectMiddleware)

	// Perbaiki routing untuk admin dashboard
	adminRouter.HandleFunc("/dashboard", controllers.AdminDashboard).Methods("GET", "OPTIONS") // Perhatikan path berubah dari /admin/dashboard menjadi /dashboard
	adminRouter.HandleFunc("/calendar", controllers.Calendar).Methods("GET")
	adminRouter.HandleFunc("/listuser", controllers.ListUser).Methods("GET")
	adminRouter.HandleFunc("/adduser", controllers.AddUser).Methods("GET", "POST")
	adminRouter.HandleFunc("/profile", controllers.Profile).Methods("GET")
	adminRouter.HandleFunc("/edituser", controllers.EditUser).Methods("GET")
	adminRouter.HandleFunc("/deleteuser", controllers.DeleteUser).Methods("GET", "POST")
	adminRouter.HandleFunc("/listdosen", controllers.ListDosen).Methods("GET")
	adminRouter.HandleFunc("/listicp", controllers.ListICP).Methods("GET", "OPTIONS") // Tambahkan route untuk ICP admin

	// Tambahkan routes untuk taruna
	tarunaRoutes := router.PathPrefix("/taruna").Subrouter()
	tarunaRoutes.Use(middleware.RoleRedirectMiddleware)

	// Route dashboard taruna
	tarunaRoutes.HandleFunc("/dashboard", controllers.TarunaDashboard).Methods("GET", "OPTIONS")
	tarunaRoutes.HandleFunc("/icp", controllers.ICP).Methods("GET", "OPTIONS")
	tarunaRoutes.HandleFunc("/editicp", controllers.EditICP).Methods("GET", "OPTIONS")
	tarunaRoutes.HandleFunc("/viewicp", controllers.ViewICPTaruna).Methods("GET", "OPTIONS")
	// Tambahkan route profile taruna
	tarunaRoutes.HandleFunc("/profile", controllers.ProfileTaruna).Methods("GET")
	// Tambahkan route edit profile taruna
	tarunaRoutes.HandleFunc("/editprofile", controllers.EditProfileTaruna).Methods("GET")
	tarunaRoutes.HandleFunc("/proposal", controllers.Proposal).Methods("GET", "OPTIONS")

	// Tambahkan routes untuk dosen
	dosenRoutes := router.PathPrefix("/dosen").Subrouter()
	dosenRoutes.Use(middleware.RoleRedirectMiddleware)

	// Route dashboard dosen
	dosenRoutes.HandleFunc("/dashboard", controllers.DosenDashboard).Methods("GET", "OPTIONS")
	dosenRoutes.HandleFunc("/review_icp", controllers.ReviewICP).Methods("GET", "OPTIONS")
	dosenRoutes.HandleFunc("/profile", controllers.ProfileDosen).Methods("GET", "OPTIONS")
	dosenRoutes.HandleFunc("/editprofile", controllers.EditProfileDosen).Methods("GET", "OPTIONS")
	dosenRoutes.HandleFunc("/viewicp", controllers.ViewICPDosen).Methods("GET", "OPTIONS")
	dosenRoutes.HandleFunc("/viewicp_review", controllers.ViewICPReviewDosen).Methods("GET", "OPTIONS")
	dosenRoutes.HandleFunc("/viewicp_revisi", controllers.ViewICPRevisiDosen).Methods("GET", "OPTIONS")

	// Tambahkan router ke http.Handle
	http.Handle("/", router) // Tambahkan ini untuk menggunakan router mux

	// Tambahkan route untuk userlist
	// http.HandleFunc("/userlist", func(w http.ResponseWriter, r *http.Request) {
	// 	log.Println("Mengakses halaman userlist")
	// 	http.ServeFile(w, r, "static/userlist.html")
	// })

	// Default redirect ke login
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/loginusers", http.StatusSeeOther)
	}).Methods("GET")

	log.Println("Auth Service running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
