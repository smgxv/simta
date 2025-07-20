package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"ta_service/utils"
)

func RoleRedirectMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("🔒 Memeriksa otorisasi...")

		// Cek token dari header Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// Coba ambil dari cookie
			cookie, err := r.Cookie("token")
			if err != nil {
				log.Println("🚫 ERROR: Token tidak ditemukan")
				http.Redirect(w, r, "/loginusers", http.StatusSeeOther)
				return
			}
			authHeader = "Bearer " + cookie.Value
		}

		// Extract token dari header
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		// Parse dan validasi token
		claims, err := utils.ParseJWT(tokenString)
		if err != nil {
			log.Printf("🚫 ERROR: Token tidak valid: %v", err)
			http.Redirect(w, r, "/loginusers", http.StatusSeeOther)
			return
		}

		// Cek role (case insensitive) dan path
		userRole := strings.ToLower(claims.Role)
		path := r.URL.Path

		// Validasi akses berdasarkan role
		switch userRole {
		case "taruna":
			// Taruna hanya boleh mengakses path /taruna/*
			if !strings.HasPrefix(path, "/taruna/") {
				log.Printf("🚫 Akses ditolak: Taruna mencoba mengakses %s", path)
				// Redirect ke taruna dashboard
				http.Redirect(w, r, "/taruna/dashboard", http.StatusSeeOther)
				return
			}
		case "admin":
			// Admin hanya boleh mengakses path /admin/*
			if !strings.HasPrefix(path, "/admin/") {
				log.Printf("🚫 Akses ditolak: Admin mencoba mengakses %s", path)
				// Redirect ke admin dashboard
				http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
				return
			}
		case "dosen":
			// Dosen hanya boleh mengakses path /dosen/*
			if !strings.HasPrefix(path, "/dosen/") {
				log.Printf("🚫 Akses ditolak: Dosen mencoba mengakses %s", path)
				// Redirect ke dosen dashboard
				http.Redirect(w, r, "/dosen/dashboard", http.StatusSeeOther)
				return
			}
		default:
			log.Printf("🚫 ERROR: Role tidak valid: %s", claims.Role)
			http.Redirect(w, r, "/loginusers", http.StatusSeeOther)
			return
		}

		// Set role di context untuk digunakan di handler
		ctx := context.WithValue(r.Context(), "userRole", userRole)
		r = r.WithContext(ctx)

		log.Printf("✅ Otorisasi berhasil untuk %s mengakses %s", userRole, path)
		next.ServeHTTP(w, r)
	})
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("CORS: Processing request from %s to %s", r.RemoteAddr, r.URL.Path)

		// Allow requests from any origin
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			log.Printf("CORS: Handling preflight request from %s", r.RemoteAddr)
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
		log.Printf("Response sent for: %s %s", r.Method, r.URL.Path)
	})
}
