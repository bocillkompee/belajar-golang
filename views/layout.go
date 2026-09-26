package views

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"strings"

	"belajar-golang/entities"
	"belajar-golang/models"
)

//go:embed templates/*.html templates/assets/**
var Assets embed.FS

func AssetsFS() fs.FS {
	sub, err := fs.Sub(Assets, "templates/assets")
	if err != nil {
		panic(err)
	}
	return sub
}

type Flash struct {
	Jenis string
	Pesan string
}

type Data struct {
	Judul    string
	Subjudul string
	Menu     string
	Flash    *Flash
	Stats    models.Statistik

	Produk   []entities.Produk
	Total    int
	Filter   models.Filter
	Kategori []string
	KategoriStat []models.KategoriStat

	Form         entities.Produk
	Errors       map[string]string
	Aksi         string
	FormAksi     string
	KategoriOpsi []string
	Detail       *entities.Produk
}

var funcMap = template.FuncMap{
	"rp": func(n any) string { return "Rp " + angkaDari(n) },
	"rpSingkat": func(n any) string {
		v := angkaDari64(n)
		if v >= 1_000_000_000 {
			return fmt.Sprintf("%.1f M", float64(v)/1_000_000_000)
		}
		if v >= 1_000_000 {
			return fmt.Sprintf("%.1f jt", float64(v)/1_000_000)
		}
		if v >= 1_000 {
			return fmt.Sprintf("%.1f rb", float64(v)/1_000)
		}
		return formatRupiah(v)
	},
	"angka":  angkaDari,
	"rapi":   angkaDari,
	"tambah": func(a, b int) int { return a + b },
	"persen": func(total int) int {
		p := total * 5
		if p > 100 {
			return 100
		}
		return p
	},
	"urutURL": func(f models.Filter, kolom string) template.URL {
		q := url.Values{}
		if f.Cari != "" {
			q.Set("q", f.Cari)
		}
		if f.Kategori != "" {
			q.Set("kategori", f.Kategori)
		}
		q.Set("urut", kolom)
		arah := f.Arah
		if f.Urutkan == kolom && arah == "asc" {
			arah = "desc"
		} else {
			arah = "asc"
		}
		q.Set("arah", arah)
		return template.URL("/produk?" + q.Encode())
	},
	"truncate": func(n int, s string) string {
		if len(s) <= n {
			return s
		}
		return s[:n] + "..."
	},
}

func angkaDari64(v any) int64 {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int32:
		return int64(n)
	case int64:
		return n
	case float64:
		return int64(n)
	}
	return 0
}

func angkaDari(v any) string {
	return formatRupiah(angkaDari64(v))
}

func formatRupiah(n int64) string {
	negatif := n < 0
	if negatif {
		n = -n
	}
	teks := fmt.Sprintf("%d", n)
	var parts []string
	for len(teks) > 3 {
		parts = append([]string{teks[len(teks)-3:]}, parts...)
		teks = teks[:len(teks)-3]
	}
	parts = append([]string{teks}, parts...)
	hasil := strings.Join(parts, ".")
	if negatif {
		return "-" + hasil
	}
	return hasil
}

var (
	tmplList   = template.Must(template.New("base").Funcs(funcMap).ParseFS(Assets, "templates/base.html", "templates/produk_list.html"))
	tmplForm   = template.Must(template.New("base").Funcs(funcMap).ParseFS(Assets, "templates/base.html", "templates/produk_form.html"))
	tmplDetail = template.Must(template.New("base").Funcs(funcMap).ParseFS(Assets, "templates/base.html", "templates/produk_detail.html"))
)

func Render(w http.ResponseWriter, halaman string, data Data) {
	set := tmplList
	nama := "halaman-list"
	switch halaman {
	case "form":
		set = tmplForm
		nama = "halaman-form"
	case "detail":
		set = tmplDetail
		nama = "halaman-detail"
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := set.ExecuteTemplate(w, nama, data); err != nil {
		http.Error(w, "Terjadi kesalahan saat merender halaman: "+err.Error(), http.StatusInternalServerError)
	}
}
