package models

import (
	"database/sql"
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

func (u UserModel) Where(user *entities.User, fieldName, fieldValue string) error {
	row, err := u.db.Query(`
        SELECT id, nama_lengkap, email, username, password, role, jurusan, kelas 
        FROM users 
        WHERE `+fieldName+` = ? 
        LIMIT 1`, fieldValue)

	if err != nil {
		return err
	}

	defer row.Close()

	for row.Next() {
		err := row.Scan(
			&user.ID,
			&user.NamaLengkap,
			&user.Email,
			&user.Username,
			&user.Password,
			&user.Role,
			&user.Jurusan,
			&user.Kelas,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u UserModel) GetDosenIDByUserID(userID int64) (int64, error) {
	var dosenID int64
	err := u.db.QueryRow("SELECT id FROM dosen WHERE user_id = ?", userID).Scan(&dosenID)
	if err != nil {
		return 0, err
	}
	return dosenID, nil
}
