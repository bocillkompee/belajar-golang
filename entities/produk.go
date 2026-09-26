package entities

import (
	"fmt"
	"strings"
	"time"
)

type Produk struct {
	ID        int64
	Kode      string
	Nama      string
	Kategori  string
	Harga     int64
	Stok      int
	Deskripsi string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p Produk) Nilai() int64 {
	return p.Harga * int64(p.Stok)
}

func (p Produk) StokKelas() string {
	switch {
	case p.Stok <= 0:
		return "habis"
	case p.Stok <= 5:
		return "menipis"
	default:
		return "aman"
	}
}

func (p Produk) StokLabel() string {
	switch p.StokKelas() {
	case "habis":
		return "Stok Habis"
	case "menipis":
		return "Stok Menipis"
	default:
		return "Stok Aman"
	}
}

func (p Produk) Inisial() string {
	parts := strings.Fields(p.Nama)
	if len(parts) == 0 {
		return "?"
	}
	hasil := strings.ToUpper(string([]rune(parts[0])[0]))
	if len(parts) > 1 {
		hasil += strings.ToUpper(string([]rune(parts[1])[0]))
	}
	return hasil
}

func (p Produk) PersentaseStok() int {
	if p.Stok <= 0 {
		return 0
	}
	persen := p.Stok * 100 / 50
	if persen > 100 {
		return 100
	}
	return persen
}

func (p Produk) Ringkasan() string {
	teks := strings.TrimSpace(p.Deskripsi)
	if len(teks) > 90 {
		return strings.TrimSpace(teks[:90]) + "..."
	}
	if teks == "" {
		return "Belum ada deskripsi untuk produk ini."
	}
	return teks
}

func (p Produk) Dibuat() string {
	return p.CreatedAt.Format("02 Jan 2006, 15:04")
}

var ikonKategori = map[string]string{
	"Elektronik": "\U0001F4F1",
	"Makanan":    "\U0001F354",
	"Minuman":    "\U0001F964",
	"Busana":     "\U0001F455",
	"Skincare":   "\U0001F484",
	"Buku":       "\U0001F4DA",
	"Olahraga":   "\U0001F3BE",
	"Rumah":      "\U0001F3E0",
	"Peralatan":  "\U0001F527",
}

func (p Produk) Emoji() string {
	if e, ok := ikonKategori[p.Kategori]; ok {
		return e
	}
	return "\U0001F4E6"
}

func (p Produk) Warna() string {
	warna := map[string]string{
		"Elektronik": "violet",
		"Makanan":    "amber",
		"Minuman":    "sky",
		"Busana":     "pink",
		"Skincare":   "rose",
		"Buku":       "indigo",
		"Olahraga":   "lime",
		"Rumah":      "emerald",
		"Peralatan":  "cyan",
	}
	if w, ok := warna[p.Kategori]; ok {
		return w
	}
	return "slate"
}

func KodeProduk(id int64) string {
	return fmt.Sprintf("PRD-%04d", id)
}

func SemuaKategori() []string {
	return []string{"Elektronik", "Makanan", "Minuman", "Busana", "Skincare", "Buku", "Olahraga", "Rumah", "Peralatan", "Lainnya"}
}
