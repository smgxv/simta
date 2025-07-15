package models

import (
	"database/sql"
	"errors"
	"ta_service/config"
	"ta_service/entities"
)

type UserModel struct {
	db *sql.DB
}

func NewUserModel() *UserModel {
	conn, err := config.DBConn()
	if err != nil {
		panic(err)
	}

	return &UserModel{
		db: conn,
	}
}

// Valid field whitelist untuk mencegah SQL injection via nama kolom
var allowedFields = map[string]bool{
	"email":    true,
	"username": true,
}

// Ambil user berdasarkan nilai dari field tertentu (email / username)
func (u UserModel) Where(user *entities.User, fieldName, fieldValue string) error {
	if !allowedFields[fieldName] {
		return errors.New("invalid field name")
	}

	query := `
		SELECT id, nama_lengkap, email, username, password, role, jurusan, kelas
		FROM users
		WHERE ` + fieldName + ` = ?
		LIMIT 1
	`

	return u.db.QueryRow(query, fieldValue).Scan(
		&user.ID,
		&user.NamaLengkap,
		&user.Email,
		&user.Username,
		&user.Password,
		&user.Role,
		&user.Jurusan,
		&user.Kelas,
	)
}

// Ambil ID dosen berdasarkan ID user
func (u UserModel) GetDosenIDByUserID(userID int64) (int64, error) {
	var dosenID int64
	err := u.db.QueryRow("SELECT id FROM dosen WHERE user_id = ?", userID).Scan(&dosenID)
	if err != nil {
		return 0, err
	}
	return dosenID, nil
}
