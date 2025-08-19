package produkmanager

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

const (
	// Maksimal 2GB
	MaxProdukSize = 2 << 30 // 2 GB
	MinProdukSize = 1 << 10 // 1 KB
)

// Daftar ekstensi/sufiks kompresi/arsip yang diizinkan.
// Gunakan suffix agar mendukung ekstensi majemuk (mis. .tar.gz)
var AllowedCompressedSuffixes = []string{
	".zip",
	".rar",
}

// IsAllowedCompressed memeriksa apakah nama file berakhiran salah satu sufiks kompresi yang diizinkan
func IsAllowedCompressed(filename string) bool {
	lower := strings.ToLower(filename)
	for _, suf := range AllowedCompressedSuffixes {
		if strings.HasSuffix(lower, suf) {
			return true
		}
	}
	return false
}

// ValidateProdukFileType memastikan hanya file arsip/kompresi yang diterima dan menolak tipe berbahaya
func ValidateProdukFileType(filename string) error {
	if !IsAllowedCompressed(filename) {
		return fmt.Errorf("hanya file arsip/kompresi yang diizinkan (ZIP dan RAR)")
	}

	// Opsional: blokir tipe eksekusi umum berdasarkan ekstensi terakhir
	ext := strings.ToLower(filepath.Ext(filename))
	blocked := []string{".exe", ".bat", ".sh", ".msi", ".js", ".php", ".py", ".dll", ".so"}
	for _, b := range blocked {
		if ext == b {
			return fmt.Errorf("file executable tidak diizinkan")
		}
	}

	return nil
}

// ValidateProdukFileName melakukan sanitasi nama file produk TA
func ValidateProdukFileName(filename string) string {
	filename = filepath.Base(filename)
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '-' || r == '_' || r == '.':
			return r
		default:
			return '_'
		}
	}, filename)
}

// SanitizeProdukPath memastikan path valid (no traversal)
func SanitizeProdukPath(baseDir, filename string) (string, error) {
	fullPath := filepath.Join(baseDir, filename)
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("gagal membaca path file: %v", err)
	}

	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return "", fmt.Errorf("gagal membaca path base: %v", err)
	}

	if !strings.HasPrefix(absPath, absBase) {
		return "", fmt.Errorf("path tidak valid: traversal terdeteksi")
	}
	return fullPath, nil
}

// SaveProdukFile menyimpan file upload dengan aman
func SaveProdukFile(file multipart.File, handler *multipart.FileHeader, uploadDir, newFilename string) (string, error) {
	// Validasi ukuran (menggunakan size dari header jika tersedia)
	if handler.Size > MaxProdukSize {
		return "", fmt.Errorf("ukuran file melebihi 2GB")
	}
	if handler.Size >= 0 && handler.Size < MinProdukSize {
		return "", fmt.Errorf("ukuran file terlalu kecil (<1KB)")
	}

	// Validasi tipe file (berdasarkan nama/ekstensi)
	if err := ValidateProdukFileType(handler.Filename); err != nil {
		return "", err
	}

	if err := os.MkdirAll(uploadDir, 0o750); err != nil {
		return "", fmt.Errorf("gagal membuat direktori upload: %v", err)
	}

	finalPath, err := SanitizeProdukPath(uploadDir, newFilename)
	if err != nil {
		return "", err
	}

	// Gunakan O_EXCL untuk mencegah overwrite
	dst, err := os.OpenFile(finalPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o640)
	if err != nil {
		return "", fmt.Errorf("gagal membuat file: %v", err)
	}
	defer dst.Close()

	// Batasi copy hingga MaxProdukSize+1 untuk deteksi oversize saat streaming
	written, err := io.Copy(dst, io.LimitReader(file, MaxProdukSize+1))
	if err != nil {
		_ = os.Remove(finalPath)
		return "", fmt.Errorf("gagal menyimpan file: %v", err)
	}
	if written > MaxProdukSize {
		_ = os.Remove(finalPath)
		return "", fmt.Errorf("file terlalu besar (limit 2GB)")
	}
	if written < MinProdukSize {
		_ = os.Remove(finalPath)
		return "", fmt.Errorf("ukuran file terlalu kecil (<1KB)")
	}

	return finalPath, nil
}
