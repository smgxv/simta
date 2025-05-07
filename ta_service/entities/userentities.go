package entities

type User struct {
	ID          int64  `json:"id"`
	NamaLengkap string `json:"nama_lengkap"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Role        string `json:"role"`
	Jurusan     string `json:"jurusan"`
	Kelas       string `json:"kelas,omitempty"`
}
