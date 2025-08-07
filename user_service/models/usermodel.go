package models

import (
	"database/sql"
	"fmt"
	"strings"
	"user_service/config"
	"user_service/entities"
)

type UserModel struct {
	db *sql.DB
}

func NewUserModel() (*UserModel, error) {
	db, err := config.ConnectDB()
	if err != nil {
		return nil, err
	}
	return &UserModel{db: db}, nil
}

func (u UserModel) Where(user *entities.User, fieldName, fieldValue string) error {
	row, err := u.db.Query("SELECT id, nama_lengkap, email, username, password, role, jurusan, kelas, npm FROM users WHERE "+fieldName+" = ? LIMIT 1", fieldValue)

	if err != nil {
		return err
	}

	defer row.Close()

	for row.Next() {
		row.Scan(&user.ID, &user.NamaLengkap, &user.Email, &user.Username, &user.Password, &user.Role, &user.Jurusan, &user.Kelas, &user.NPM)
	}

	return nil
}

func (m *UserModel) FindAll() ([]entities.User, error) {
	rows, err := m.db.Query(`
        SELECT id, nama_lengkap, username, email, role, jurusan, kelas, npm
        FROM users
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []entities.User
	for rows.Next() {
		var user entities.User
		err := rows.Scan(
			&user.ID,
			&user.NamaLengkap,
			&user.Username,
			&user.Email,
			&user.Role,
			&user.Jurusan,
			&user.Kelas,
			&user.NPM,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (u UserModel) CreateUser(fullName, email, username, role, password, jurusan, kelas, npm string) (int64, error) {
	// Mulai transaksi
	tx, err := u.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("error memulai transaksi: %v", err)
	}

	// Insert ke tabel users
	result, err := tx.Exec("INSERT INTO users (nama_lengkap, email, username, role, password, jurusan, kelas, npm) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		fullName, email, username, role, password, jurusan, kelas, npm)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("error inserting user: %v", err)
	}

	// Dapatkan ID user yang baru dibuat
	userID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("error getting last insert ID: %v", err)
	}

	// Insert ke tabel sesuai role
	switch strings.ToLower(role) {
	case "dosen":
		_, err = tx.Exec("INSERT INTO dosen (user_id, nama_lengkap, email, jurusan) VALUES (?, ?, ?, ?)",
			userID, fullName, email, jurusan)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("error inserting dosen: %v", err)
		}

	case "taruna":
		_, err = tx.Exec(`INSERT INTO taruna (user_id, nama_lengkap, email, jurusan, kelas, npm) VALUES (?, ?, ?, ?, ?, ?)`,
			userID, fullName, email, jurusan, kelas, npm)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("error inserting taruna: %v", err)
		}
	}

	// Commit transaksi
	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("error committing transaction: %v", err)
	}

	return userID, nil
}

func (u UserModel) UpdateUser(userID int, fullName, email, username, role, jurusan, kelas string, npm *int) error {
	// Normalisasi role ke lowercase
	role = strings.ToLower(role)

	// Mulai transaksi
	tx, err := u.db.Begin()
	if err != nil {
		return fmt.Errorf("error memulai transaksi: %v", err)
	}

	// Dapatkan role lama user
	var oldRole string
	err = tx.QueryRow("SELECT role FROM users WHERE id = ?", userID).Scan(&oldRole)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error mendapatkan role lama: %v", err)
	}
	oldRole = strings.ToLower(oldRole)

	// Update tabel users
	_, err = tx.Exec(`UPDATE users SET nama_lengkap = ?, email = ?, username = ?, role = ?, jurusan = ?, kelas = ?, npm = ? WHERE id = ?`,
		fullName, email, username, role, jurusan, kelas, npm, userID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error updating user: %v", err)
	}

	// Jika role berubah
	if oldRole != role {
		// Hapus data dari tabel role lama
		switch oldRole {
		case "dosen":
			_, err = tx.Exec("DELETE FROM dosen WHERE user_id = ?", userID)
		case "taruna":
			_, err = tx.Exec("DELETE FROM taruna WHERE user_id = ?", userID)
		}
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error menghapus data role lama: %v", err)
		}

		// Masukkan data ke tabel role baru
		switch role {
		case "dosen":
			_, err = tx.Exec("INSERT INTO dosen (user_id, nama_lengkap, email, jurusan) VALUES (?, ?, ?, ?)",
				userID, fullName, email, jurusan)
		case "taruna":
			_, err = tx.Exec(`INSERT INTO taruna (user_id, nama_lengkap, email, jurusan, kelas, npm) VALUES (?, ?, ?, ?, ?, ?)`,
				userID, fullName, email, jurusan, kelas, npm)
		}
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error menambah data role baru: %v", err)
		}

	} else {
		// Role tidak berubah, cukup update data tabel role-nya
		switch role {
		case "dosen":
			_, err = tx.Exec("UPDATE dosen SET nama_lengkap = ?, email = ?, jurusan = ? WHERE user_id = ?",
				fullName, email, jurusan, userID)
		case "taruna":
			_, err = tx.Exec("UPDATE taruna SET nama_lengkap = ?, email = ?, jurusan = ?, kelas = ?, npm = ? WHERE user_id = ?",
				fullName, email, jurusan, kelas, npm, userID)
		}
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error mengupdate data role: %v", err)
		}
	}

	// Commit transaksi
	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}

func (u UserModel) GetUserByID(userID int) (*entities.User, error) {
	user := &entities.User{}
	err := u.db.QueryRow(`
		SELECT id, nama_lengkap, email, username, role, jurusan, kelas, npm
		FROM users WHERE id = ?`, userID).
		Scan(&user.ID, &user.NamaLengkap, &user.Email, &user.Username,
			&user.Role, &user.Jurusan, &user.Kelas, &user.NPM)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u UserModel) DeleteUser(userID int) error {
	_, err := u.db.Exec("DELETE FROM users WHERE id = ?", userID)
	return err
}
