package models

import (
	"context"
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

func (u UserModel) Where(ctx context.Context, user *entities.User, fieldName, fieldValue string) error {
	// Validasi input ke himpunan kecil nilai yang diperbolehkan
	switch fieldName {
	case "email", "username":
	default:
		return errors.New("invalid field name")
	}

	const q = `
		SELECT id, nama_lengkap, email, username, password, role, jurusan, kelas
		FROM users
		WHERE
		  ( ? = 'email'    AND email    = ? )
		  OR
		  ( ? = 'username' AND username = ? )
		LIMIT 1`

	row := u.db.QueryRowContext(ctx, q, fieldName, fieldValue, fieldName, fieldValue)
	return row.Scan(
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
