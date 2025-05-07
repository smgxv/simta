package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"ta_service/entities"
	"ta_service/models"
	"ta_service/utils"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Struktur data untuk request login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Struktur data untuk response login
type LoginResponse struct {
	Email       string          `json:"email"`
	ID          int64           `json:"id"`
	Token       string          `json:"token"`
	Users       []entities.User `json:"users"` // Tambahkan field Users
	Role        string          `json:"role"`
	Success     bool            `json:"success"`
	RedirectURL string          `json:"redirect_url"` // Tambahkan field ini
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("=== Memulai proses login ===")

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight request
	if r.Method == "OPTIONS" {
		log.Println("Handling OPTIONS request")
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		log.Println("❌ ERROR: Method tidak diizinkan:", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("❌ ERROR: Gagal decode request body:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Println("👤 Mencoba login dengan email:", req.Email)

	// Validasi user dari database
	userModel := models.NewUserModel()
	var user entities.User

	if err := userModel.Where(&user, "email", req.Email); err != nil {
		if err == sql.ErrNoRows {
			json.NewEncoder(w).Encode(map[string]string{"error": "Email atau password salah"})
			return
		}
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Verifikasi password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Email atau password salah"})
		return
	}

	log.Println("✅ Login berhasil untuk user:", req.Email)

	// Generate token dengan role
	token, err := utils.GenerateJWT(user.Email, user.Role)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	log.Println("🔑 Token berhasil dibuat untuk user:", req.Email)

	// Buat request ke API users
	client := &http.Client{}
	apiURL := getEnv("API_SERVICE_URL", "http://104.43.89.154:8081")
	reqAPI, err := http.NewRequest("GET", apiURL+"/users", nil)
	if err != nil {
		log.Println("❌ ERROR: Gagal membuat request ke API users:", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	reqAPI.Header.Set("Authorization", "Bearer "+token)
	log.Println("📡 Mengirim request ke API users")

	// Kirim request
	resp, err := client.Do(reqAPI)
	if err != nil {
		log.Println("❌ ERROR: Gagal mengambil data users:", err)
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Baca response
	var users []entities.User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		log.Println("❌ ERROR: Gagal decode response users:", err)
		http.Error(w, "Error processing users data", http.StatusInternalServerError)
		return
	}

	log.Printf("📊 Berhasil mengambil data %d users", len(users))

	// Buat response
	response := LoginResponse{
		Email:       user.Email,
		ID:          user.ID,
		Token:       token,
		Role:        user.Role,
		Success:     true,
		RedirectURL: getDashboardURL(user.Role), // Tambahkan URL redirect berdasarkan role
	}
	// Setelah login berhasil dan response dikirim
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("❌ ERROR: Gagal encode response:", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Proses login selesai untuk user: %s dengan role: %s", req.Email, user.Role)

	log.Println("=== Akhir proses login ===")
}

// Tambahkan fungsi helper untuk menentukan URL redirect
func getDashboardURL(role string) string {
	role = strings.ToLower(role)
	if role == "taruna" {
		return "/taruna/dashboard"
	}
	return "/admin/dashboard"
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("🔄 Memproses logout...")

	// Set header
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// Handle preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Daftar semua cookie yang mungkin ada
	cookies := []string{
		"token",
		"role",
		"session",
		"userId",
		"username",
		"auth",
		"user_session",
		"remember_token",
	}

	// Hapus cookie untuk semua path yang mungkin
	paths := []string{"/", "/admin", "/taruna", "/admin/", "/taruna/"}

	for _, cookieName := range cookies {
		for _, path := range paths {
			cookie := &http.Cookie{
				Name:     cookieName,
				Value:    "",
				Path:     path,
				Domain:   "",
				MaxAge:   -1,
				Expires:  time.Now().Add(-24 * time.Hour),
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
			}
			http.SetCookie(w, cookie)
		}
	}

	// Kirim response JSON
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "success",
		"message":  "Logout berhasil",
		"redirect": "/loginusers",
	})

	log.Println("✅ Logout berhasil")
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
