package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"user_service/entities"
	"user_service/models"
	"user_service/utils"

	"golang.org/x/crypto/bcrypt"
)

// Fungsi Get user
func UserHandler(w http.ResponseWriter, r *http.Request) {
	// Header CORS
	w.Header().Set("Access-Control-Allow-Origin", "https://securesimta.my.id")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "No token provided", http.StatusUnauthorized)
		return
	}

	userModel, err := models.NewUserModel()
	if err != nil {
		log.Printf("Database connection error: %v", err)
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	// Cek apakah ada parameter id
	userID := r.URL.Query().Get("id")
	if userID != "" {
		// Jika ada id, ambil data user spesifik
		id, err := strconv.Atoi(userID)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		user, err := userModel.GetUserByID(id)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(user)
		return
	}

	// Jika tidak ada id, ambil semua user
	users, err := userModel.FindAll()
	if err != nil {
		log.Printf("Failed to fetch users: %v", err)
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(users)
}

// Fungsi Add user
func AddUser(w http.ResponseWriter, r *http.Request) {
	// Header CORS
	w.Header().Set("Access-Control-Allow-Origin", "https://securesimta.my.id")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "No token provided", http.StatusUnauthorized)
		return
	}

	// Inisialisasi userModel
	userModel, err := models.NewUserModel()
	if err != nil {
		log.Printf("Database connection error: %v", err)
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		// Parse request body
		var userData struct {
			FullName        string `json:"nama_lengkap"`
			Email           string `json:"email"`
			Username        string `json:"username"`
			Role            string `json:"role"`
			Password        string `json:"password"`
			ConfirmPassword string `json:"confirm_password"`
			Jurusan         string `json:"jurusan"`
			Kelas           string `json:"kelas"`
			NPM             string `json:"npm"`
		}

		if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validasi password
		if userData.Password != userData.ConfirmPassword {
			http.Error(w, "Password dan Confirm Password tidak cocok!", http.StatusBadRequest)
			return
		}

		if !utils.IsValidPassword(userData.Password) {
			http.Error(w, "Password harus minimal 8 karakter, mengandung huruf besar, huruf kecil, angka, dan simbol", http.StatusBadRequest)
			return
		}
		// Cek email duplikat
		var existingUser entities.User
		userModel.Where(&existingUser, "email", userData.Email)
		if existingUser.Email != "" {
			http.Error(w, "Email sudah terdaftar!", http.StatusConflict)
			return
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}

		// Simpan user baru
		_, err = userModel.CreateUser(userData.FullName, userData.Email, userData.Username, userData.Role, string(hashedPassword), userData.Jurusan, userData.Kelas, userData.NPM)
		if err != nil {
			log.Printf("Error creating user: %v", err) // Tambahkan logging ini
			http.Error(w, "Gagal menambahkan user", http.StatusInternalServerError)
			return
		}

		// Kirim response sukses
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "User berhasil ditambahkan",
		})
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Fungsi Edit user
func EditUser(w http.ResponseWriter, r *http.Request) {
	// Header CORS
	w.Header().Set("Access-Control-Allow-Origin", "https://securesimta.my.id")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Origin, Accept")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "No token provided", http.StatusUnauthorized)
		return
	}

	// Inisialisasi userModel
	userModel, err := models.NewUserModel()
	if err != nil {
		log.Printf("Database connection error: %v", err)
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodGet {
		// Ambil ID user dari parameter URL
		userIDStr := r.URL.Query().Get("id")
		if userIDStr == "" {
			http.Error(w, "ID User tidak ditemukan", http.StatusBadRequest)
			return
		}

		// Konversi ID string ke int
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "ID User tidak valid", http.StatusBadRequest)
			return
		}

		// Ambil data user yang akan diedit
		user, err := userModel.GetUserByID(userID)
		if err != nil {
			http.Error(w, "User tidak ditemukan", http.StatusNotFound)
			return
		}

		// Kirim response dalam format JSON
		json.NewEncoder(w).Encode(user)

	} else if r.Method == http.MethodPut {
		// Parse request body
		var userData struct {
			UserID   int    `json:"id"`
			FullName string `json:"nama_lengkap"`
			Email    string `json:"email"`
			Username string `json:"username"`
			Role     string `json:"role"`
			Jurusan  string `json:"jurusan"`
			Kelas    string `json:"kelas"`
			NPM      string `json:"npm"`
		}

		if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
			log.Printf("Error decoding request body: %v", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validasi data yang diperlukan
		if userData.UserID == 0 || userData.FullName == "" || userData.Email == "" || userData.Username == "" || userData.Role == "" {
			http.Error(w, "Semua field harus diisi", http.StatusBadRequest)
			return
		}

		// Cek apakah email sudah digunakan oleh user lain
		var existingUser entities.User
		err = userModel.Where(&existingUser, "email", userData.Email)
		if err == nil && existingUser.ID != userData.UserID && existingUser.Email != "" {
			http.Error(w, "Email sudah digunakan", http.StatusConflict)
			return
		}

		// Parse NPM jika tidak kosong
		var npm *int
		if userData.NPM != "" {
			npmInt, err := strconv.Atoi(userData.NPM)
			if err != nil {
				http.Error(w, "NPM harus berupa angka", http.StatusBadRequest)
				return
			}
			npm = &npmInt
		}

		// Update data user
		err := userModel.UpdateUser(userData.UserID, userData.FullName, userData.Email, userData.Username, userData.Role, userData.Jurusan, userData.Kelas, npm)
		if err != nil {
			log.Printf("Failed to update user: %v", err)
			http.Error(w, "Gagal mengupdate user", http.StatusInternalServerError)
			return
		}

		// Kirim response sukses
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "User berhasil diupdate",
		})
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Fungsi Get user detail
func GetUserDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "https://securesimta.my.id")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  false,
			"message": "Metode tidak diizinkan",
		})
		return
	}

	userID := r.URL.Query().Get("id")
	if userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  false,
			"message": "ID User tidak ditemukan",
		})
		return
	}

	id, err := strconv.Atoi(userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  false,
			"message": "ID User tidak valid",
		})
		return
	}

	userModel, err := models.NewUserModel()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  false,
			"message": "Gagal menghubungkan ke database",
		})
		return
	}

	user, err := userModel.GetUserByID(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  false,
			"message": "User tidak ditemukan",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// Fungsi Delete user
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Header CORS
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "https://securesimta.my.id")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  false,
			"message": "Metode tidak diizinkan",
		})
		return
	}

	// Ambil ID user dari parameter URL
	userID := r.URL.Query().Get("id")
	if userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  false,
			"message": "ID User tidak ditemukan",
		})
		return
	}

	// Konversi userID string ke int
	id, err := strconv.Atoi(userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  false,
			"message": "ID User tidak valid",
		})
		return
	}

	// Inisialisasi user model
	userModel, err := models.NewUserModel()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  false,
			"message": "Gagal menghubungkan ke database",
		})
		return
	}

	// Cek apakah user ada di database
	user, err := userModel.GetUserByID(id)
	if err != nil || user == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  false,
			"message": "User tidak ditemukan",
		})
		return
	}

	// Hapus user dari database
	err = userModel.DeleteUser(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  false,
			"message": "Gagal menghapus user: " + err.Error(),
		})
		return
	}

	// Kirim response sukses
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  true,
		"message": "User berhasil dihapus",
	})
}
