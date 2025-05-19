package controllers

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"ta_service/entities"
	"ta_service/utils"
	"time"
)

// Fungsi untuk menampilkan dashboard
func Index(w http.ResponseWriter, r *http.Request) {
	temp, _ := template.ParseFiles("static/dashboard.html")
	temp.Execute(w, nil)
}

func LoginUsers(w http.ResponseWriter, r *http.Request) {
	temp, _ := template.ParseFiles("static/login.html")
	temp.Execute(w, nil)
}

func AdminDashboard(w http.ResponseWriter, r *http.Request) {
	// Set header content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Ambil token dari cookie atau header
	var tokenString string
	cookie, err := r.Cookie("token")
	if err == nil {
		tokenString = cookie.Value
	} else {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenString = strings.Replace(authHeader, "Bearer ", "", 1)
		}
	}

	// Validasi token
	claims, err := utils.ParseJWT(tokenString)
	if err != nil || strings.ToLower(claims.Role) != "admin" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Serve the admin dashboard HTML file
	http.ServeFile(w, r, "static/admin/admin_dashboard.html")
}

func Calendar(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles("static/admin/calendar.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	temp.Execute(w, nil)
}

func ListUser(w http.ResponseWriter, r *http.Request) {
	// Set header content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Ambil token dari cookie atau header
	var tokenString string
	cookie, err := r.Cookie("token")
	if err == nil {
		tokenString = cookie.Value
	} else {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenString = strings.Replace(authHeader, "Bearer ", "", 1)
		}
	}

	// Validasi token
	claims, err := utils.ParseJWT(tokenString)
	if err != nil || strings.ToLower(claims.Role) != "admin" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Serve the admin dashboard HTML file
	http.ServeFile(w, r, "static/admin/listuser.html")
}

func ListDosen(w http.ResponseWriter, r *http.Request) {
	// Set header content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Ambil token dari cookie atau header
	var tokenString string
	cookie, err := r.Cookie("token")
	if err == nil {
		tokenString = cookie.Value
	} else {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenString = strings.Replace(authHeader, "Bearer ", "", 1)
		}
	}

	// Validasi token
	claims, err := utils.ParseJWT(tokenString)
	if err != nil || strings.ToLower(claims.Role) != "admin" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Serve the admin dashboard HTML file
	http.ServeFile(w, r, "static/admin/listdosen.html")
}

func AddUser(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles("static/admin/adduser.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	temp.Execute(w, nil)
}

func Profile(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles("static/admin/profile.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	temp.Execute(w, nil)
}

func EditUser(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles("static/admin/edituser.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	temp.Execute(w, nil)
}

// ... existing code ...
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Set header content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Ambil token dari localStorage atau sessionStorage
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		// Jika tidak ada di header, coba ambil dari cookie
		cookie, err := r.Cookie("token")
		if err == nil {
			tokenString = cookie.Value
		}
	}

	// Validasi token
	claims, err := utils.ParseJWT(tokenString)
	if err != nil || strings.ToLower(claims.Role) != "admin" {
		http.Redirect(w, r, "/loginusers", http.StatusSeeOther)
		return
	}

	if r.Method == "GET" {
		userID := r.URL.Query().Get("id")

		// Tambahkan timeout untuk request
		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		apiURL := "http://104.43.89.154:8081/users/detail?id=" + userID
		log.Printf("Mencoba mengakses API: %s", apiURL)

		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			log.Printf("Error membuat request: %v", err)
			http.Error(w, "Gagal membuat request", http.StatusInternalServerError)
			return
		}

		// Pastikan token diteruskan dengan benar
		if !strings.HasPrefix(tokenString, "Bearer ") {
			tokenString = "Bearer " + tokenString
		}
		req.Header.Set("Authorization", tokenString)

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error mengakses API: %v", err)
			// Tampilkan pesan error yang lebih informatif
			if strings.Contains(err.Error(), "connection refused") {
				http.Error(w, "API service tidak dapat diakses. Pastikan API service sudah berjalan di port 8081", http.StatusServiceUnavailable)
			} else {
				http.Error(w, "Gagal mengambil data user: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}
		defer resp.Body.Close()

		// Baca response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error membaca response: %v", err)
			http.Error(w, "Gagal membaca response dari API", http.StatusInternalServerError)
			return
		}

		// Log response untuk debugging
		log.Printf("API Response: %s", string(body))

		var user entities.User
		if err := json.Unmarshal(body, &user); err != nil {
			log.Printf("Error decode JSON: %v", err)
			http.Error(w, "Gagal memproses data user", http.StatusInternalServerError)
			return
		}

		// Parse template
		temp, err := template.ParseFiles("static/admin/deleteuser.html")
		if err != nil {
			log.Printf("Error parsing template: %v", err)
			http.Error(w, "Gagal memuat template", http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"user": user,
		}

		if err := temp.Execute(w, data); err != nil {
			log.Printf("Error executing template: %v", err)
			http.Error(w, "Gagal menampilkan halaman", http.StatusInternalServerError)
			return
		}
	}
}

// TARUNA WEB SERVICE

// TarunaDashboard menangani tampilan dashboard untuk taruna
func TarunaDashboard(w http.ResponseWriter, r *http.Request) {
	// Set header content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Ambil token dari cookie atau header
	var tokenString string
	cookie, err := r.Cookie("token")
	if err == nil {
		tokenString = cookie.Value
	} else {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenString = strings.Replace(authHeader, "Bearer ", "", 1)
		}
	}

	// Validasi token
	claims, err := utils.ParseJWT(tokenString)
	if err != nil || strings.ToLower(claims.Role) != "taruna" {
		http.Redirect(w, r, "/loginusers", http.StatusSeeOther)
		return
	}

	// Serve the taruna dashboard HTML file
	http.ServeFile(w, r, "static/taruna/taruna_dashboard.html")
}

func ICP(w http.ResponseWriter, r *http.Request) {
	// Set header content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Ambil token dari cookie atau header
	var tokenString string
	cookie, err := r.Cookie("token")
	if err == nil {
		tokenString = cookie.Value
	} else {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenString = strings.Replace(authHeader, "Bearer ", "", 1)
		}
	}

	// Validasi token
	claims, err := utils.ParseJWT(tokenString)
	if err != nil || strings.ToLower(claims.Role) != "taruna" { // Ubah dari "admin" menjadi "taruna"
		http.Redirect(w, r, "/loginusers", http.StatusSeeOther)
		return
	}

	// Serve the ICP HTML file
	http.ServeFile(w, r, "static/taruna/icp.html")
}

func EditICP(w http.ResponseWriter, r *http.Request) {
	// Set header content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Validasi token dan role
	var tokenString string
	cookie, err := r.Cookie("token")
	if err == nil {
		tokenString = cookie.Value
	} else {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenString = strings.Replace(authHeader, "Bearer ", "", 1)
		}
	}

	claims, err := utils.ParseJWT(tokenString)
	if err != nil || strings.ToLower(claims.Role) != "taruna" {
		http.Redirect(w, r, "/loginusers", http.StatusSeeOther)
		return
	}

	http.ServeFile(w, r, "static/taruna/editicp.html")
}

// DOSEN WEB SERVICE

func DosenDashboard(w http.ResponseWriter, r *http.Request) {
	// Set header content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Ambil token dari cookie atau header
	var tokenString string
	cookie, err := r.Cookie("token")
	if err == nil {
		tokenString = cookie.Value
	} else {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenString = strings.Replace(authHeader, "Bearer ", "", 1)
		}
	}

	// Validasi token
	claims, err := utils.ParseJWT(tokenString)
	if err != nil || strings.ToLower(claims.Role) != "dosen" {
		http.Redirect(w, r, "/loginusers", http.StatusSeeOther)
		return
	}

	// Serve the dosen dashboard HTML file
	http.ServeFile(w, r, "static/dosen/dosen_dashboard.html")
}

func ReviewICP(w http.ResponseWriter, r *http.Request) {
	// Set header content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Validasi token dan role
	var tokenString string
	cookie, err := r.Cookie("token")
	if err == nil {
		tokenString = cookie.Value
	} else {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenString = strings.Replace(authHeader, "Bearer ", "", 1)
		}
	}

	claims, err := utils.ParseJWT(tokenString)
	if err != nil || strings.ToLower(claims.Role) != "dosen" {
		http.Redirect(w, r, "/loginusers", http.StatusSeeOther)
		return
	}

	// Serve the review ICP HTML file
	http.ServeFile(w, r, "static/dosen/reviewicp.html")
}

// Handler untuk profile dosen
func ProfileDosen(w http.ResponseWriter, r *http.Request) {
	// Set header content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Ambil token dari cookie atau header
	var tokenString string
	cookie, err := r.Cookie("token")
	if err == nil {
		tokenString = cookie.Value
	} else {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenString = strings.Replace(authHeader, "Bearer ", "", 1)
		}
	}

	// Validasi token
	claims, err := utils.ParseJWT(tokenString)
	if err != nil || strings.ToLower(claims.Role) != "dosen" {
		http.Redirect(w, r, "/loginusers", http.StatusSeeOther)
		return
	}

	http.ServeFile(w, r, "static/dosen/profile_dosen.html")
}

// Handler untuk profile taruna
func ProfileTaruna(w http.ResponseWriter, r *http.Request) {
	// Set header content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Ambil token dari cookie atau header
	var tokenString string
	cookie, err := r.Cookie("token")
	if err == nil {
		tokenString = cookie.Value
	} else {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenString = strings.Replace(authHeader, "Bearer ", "", 1)
		}
	}

	// Validasi token
	claims, err := utils.ParseJWT(tokenString)
	if err != nil || strings.ToLower(claims.Role) != "taruna" {
		http.Redirect(w, r, "/loginusers", http.StatusSeeOther)
		return
	}

	http.ServeFile(w, r, "static/taruna/profile_taruna.html")
}

// Handler untuk halaman edit profile taruna
func EditProfileTaruna(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Ambil token dari cookie atau header
	var tokenString string
	cookie, err := r.Cookie("token")
	if err == nil {
		tokenString = cookie.Value
	} else {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenString = strings.Replace(authHeader, "Bearer ", "", 1)
		}
	}

	// Validasi token
	claims, err := utils.ParseJWT(tokenString)
	if err != nil || strings.ToLower(claims.Role) != "taruna" {
		http.Redirect(w, r, "/loginusers", http.StatusSeeOther)
		return
	}

	http.ServeFile(w, r, "static/taruna/edituser_taruna.html")
}
