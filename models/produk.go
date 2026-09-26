package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"belajar-golang/config"
	"belajar-golang/entities"
)

var ErrTidakDitemukan = errors.New("produk tidak ditemukan")

type Filter struct {
	Cari     string
	Kategori string
	Urutkan  string
	Arah     string
}

func (f Filter) Aktif() bool {
	return f.Cari != "" || f.Kategori != ""
}

type Statistik struct {
	TotalProduk     int
	TotalStok       int
	TotalNilai      int64
	StokMenipis     int
	StokHabis       int
	JumlahKategori  int
}

type KategoriStat struct {
	Nama   string
	Jumlah int
}

func KategoriBesertaJumlah() ([]KategoriStat, error) {
	rows, err := config.DB.Query("SELECT kategori, COUNT(*) AS jumlah FROM produk GROUP BY kategori ORDER BY jumlah DESC, kategori ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var daftar []KategoriStat
	for rows.Next() {
		var k KategoriStat
		if err := rows.Scan(&k.Nama, &k.Jumlah); err != nil {
			return nil, err
		}
		daftar = append(daftar, k)
	}
	return daftar, rows.Err()
}

var kolomUrut = map[string]string{
	"":         "id",
	"nama":     "nama",
	"kategori": "kategori",
	"harga":    "harga",
	"stok":     "stok",
	"terbaru":  "id",
	"terlama":  "id",
}

func kolomFilter(f Filter) (string, string) {
	kolom, ok := kolomUrut[f.Urutkan]
	if !ok {
		kolom = "id"
	}
	arah := "DESC"
	switch {
	case f.Urutkan == "nama" || f.Urutkan == "kategori" || f.Arah == "asc":
		arah = "ASC"
	}
	if f.Urutkan == "terlama" {
		arah = "ASC"
	}
	return kolom, arah
}

func Migrasi() error {
	stmt := `CREATE TABLE IF NOT EXISTS produk (
		id INT UNSIGNED NOT NULL AUTO_INCREMENT,
		kode VARCHAR(32) NOT NULL,
		nama VARCHAR(150) NOT NULL,
		kategori VARCHAR(60) NOT NULL,
		harga BIGINT NOT NULL DEFAULT 0,
		stok INT NOT NULL DEFAULT 0,
		deskripsi TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		UNIQUE KEY uk_produk_kode (kode),
		KEY idx_produk_kategori (kategori),
		KEY idx_produk_nama (nama)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

	if _, err := config.DB.Exec(stmt); err != nil {
		return err
	}
	return nil
}

func Seed(jumlah int) error {
	var total int
	if err := config.DB.QueryRow("SELECT COUNT(*) FROM produk").Scan(&total); err != nil {
		return err
	}
	if total > 0 {
		return nil
	}

	contoh := []entities.Produk{
		{Nama: "Laptop Ultrabook 14 inci", Kategori: "Elektronik", Harga: 12500000, Stok: 12, Deskripsi: "Bodi magnesium, RAM 16GB, SSD 512GB, layar OLED 2.8K. Cocok untuk kerja desain dan coding."},
		{Nama: "TWS Wireless Earbuds", Kategori: "Elektronik", Harga: 349000, Stok: 4, Deskripsi: "Noise cancelling aktif, daya tahan 28 jam dengan casing pengisi daya."},
		{Nama: "Kopi Arabika Gayo 250g", Kategori: "Makanan", Harga: 78000, Stok: 65, Deskripsi: "Biji kopi single origin dari dataran tinggi Gayo, roasting medium."},
		{Nama: "Keripik Kentang Ori", Kategori: "Makanan", Harga: 25000, Stok: 120, Deskripsi: "Kentang pilihan, tanpa pengawet buatan, kemasan 250 gram."},
		{Nama: "Air Mineral 600ml", Kategori: "Minuman", Harga: 8000, Stok: 240, Deskripsi: "Kemasan botol PET, aman untuk minuman harian."},
		{Nama: "Matcha Latte Botol", Kategori: "Minuman", Harga: 32000, Stok: 0, Deskripsi: "Matcha premium dengan susu oat, tanpa gula tambahan."},
		{Nama: "Kemeja Flanel Oversize", Kategori: "Busana", Harga: 189000, Stok: 34, Deskripsi: "Bahan katun flanel tebal, potongan oversize, warna earth tone."},
		{Nama: "Sneakers Urban Runner", Kategori: "Busana", Harga: 649000, Stok: 3, Deskripsi: "Midsole busa, outsole karet, ringan dipakai harian."},
		{Nama: "Serum Vitamin C 10%", Kategori: "Skincare", Harga: 159000, Stok: 52, Deskripsi: "Mencerahkan dan menyamarkan noda, tekstur cepat meresap."},
		{Nama: "Buku Belajar Go_lang", Kategori: "Buku", Harga: 145000, Stok: 21, Deskripsi: "Panduan lengkap Go dari dasar sampai web service, lengkap dengan contoh kode."},
		{Nama: "Matras Yoga Premium", Kategori: "Olahraga", Harga: 225000, Stok: 17, Deskripsi: "TPE tebal 8mm, anti-slip, dilengkapi tas jinjing."},
		{Nama: "Rice Cooker 1.8L", Kategori: "Peralatan", Harga: 389000, Stok: 9, Deskripsi: "Penanak otomatis, panci anti lengket, tiga fungsi masak."},
		{Nama: "Lampu Meja Minimalis", Kategori: "Rumah", Harga: 129000, Stok: 27, Deskripsi: "LED touch dimmer, warna warm white, kabel USB 1.5 meter."},
	}

	kontrak := make([]entities.Produk, 0, len(contoh))
	for i := 0; i < jumlah && i < len(contoh); i++ {
		kontrak = append(kontrak, contoh[i])
	}

	for _, p := range kontrak {
		if _, err := buatProduk(p); err != nil {
			return err
		}
	}
	return nil
}

const selectProduk = `SELECT id, kode, nama, kategori, harga, stok, COALESCE(deskripsi, ''), created_at, updated_at FROM produk`

func pindaian(rows *sql.Rows) ([]entities.Produk, error) {
	defer rows.Close()
	var daftar []entities.Produk
	for rows.Next() {
		var p entities.Produk
		if err := rows.Scan(&p.ID, &p.Kode, &p.Nama, &p.Kategori, &p.Harga, &p.Stok, &p.Deskripsi, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		daftar = append(daftar, p)
	}
	return daftar, rows.Err()
}

func SemuaProduk(f Filter) ([]entities.Produk, error) {
	var syarat []string
	var args []any

	if f.Cari != "" {
		kunci := "%" + f.Cari + "%"
		syarat = append(syarat, "(nama LIKE ? OR kode LIKE ? OR kategori LIKE ? OR deskripsi LIKE ?)")
		args = append(args, kunci, kunci, kunci, kunci)
	}
	if f.Kategori != "" {
		syarat = append(syarat, "kategori = ?")
		args = append(args, f.Kategori)
	}

	where := ""
	if len(syarat) > 0 {
		where = " WHERE " + strings.Join(syarat, " AND ")
	}

	kolom, arah := kolomFilter(f)
	query := fmt.Sprintf("%s%s ORDER BY %s %s", selectProduk, where, kolom, arah)

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	return pindaian(rows)
}

func CariProduk(id int64) (entities.Produk, error) {
	var p entities.Produk
	query := selectProduk + " WHERE id = ?"
	err := config.DB.QueryRow(query, id).Scan(&p.ID, &p.Kode, &p.Nama, &p.Kategori, &p.Harga, &p.Stok, &p.Deskripsi, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return p, ErrTidakDitemukan
	}
	return p, err
}

func buatProduk(p entities.Produk) (int64, error) {
	tx, err := config.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		"INSERT INTO produk (kode, nama, kategori, harga, stok, deskripsi) VALUES (?, ?, ?, ?, ?, ?)",
		"tmp", p.Nama, p.Kategori, p.Harga, p.Stok, p.Deskripsi,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if _, err := tx.Exec("UPDATE produk SET kode = ? WHERE id = ?", entities.KodeProduk(id), id); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func TambahProduk(p entities.Produk) (int64, error) {
	return buatProduk(p)
}

func UbahProduk(id int64, p entities.Produk) error {
	res, err := config.DB.Exec(
		"UPDATE produk SET nama = ?, kategori = ?, harga = ?, stok = ?, deskripsi = ? WHERE id = ?",
		p.Nama, p.Kategori, p.Harga, p.Stok, p.Deskripsi, id,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return ErrTidakDitemukan
	}
	return nil
}

func HapusProduk(id int64) error {
	res, err := config.DB.Exec("DELETE FROM produk WHERE id = ?", id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return ErrTidakDitemukan
	}
	return nil
}

func SemuaKategori() ([]string, error) {
	rows, err := config.DB.Query("SELECT DISTINCT kategori FROM produk ORDER BY kategori ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var daftar []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, err
		}
		daftar = append(daftar, k)
	}
	return daftar, rows.Err()
}

func StatistikProduk() (Statistik, error) {
	var s Statistik
	query := `SELECT
		COUNT(*),
		COALESCE(SUM(stok), 0),
		COALESCE(SUM(harga * stok), 0),
		COALESCE(SUM(CASE WHEN stok BETWEEN 1 AND 5 THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN stok <= 0 THEN 1 ELSE 0 END), 0),
		COUNT(DISTINCT kategori)
	FROM produk`

	err := config.DB.QueryRow(query).Scan(&s.TotalProduk, &s.TotalStok, &s.TotalNilai, &s.StokMenipis, &s.StokHabis, &s.JumlahKategori)
	return s, err
}

func Inisialisasi() error {
	if err := Migrasi(); err != nil {
		return err
	}
	return Seed(13)
}
