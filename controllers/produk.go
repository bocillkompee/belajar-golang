package controllers

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"belajar-golang/entities"
	"belajar-golang/models"
	"belajar-golang/views"
)

type ProdukController struct{}

var pesanSukses = map[string]string{
	"tambah": "Produk baru berhasil ditambahkan.",
	"edit":   "Perubahan produk berhasil disimpan.",
	"hapus":  "Produk berhasil dihapus.",
}

func Router() http.Handler {
	c := &ProdukController{}

	mux := http.NewServeMux()

	mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(views.AssetsFS()))))
	mux.HandleFunc("GET /", c.dashboard)
	mux.HandleFunc("GET /produk", c.daftarProduk)
	mux.HandleFunc("GET /produk/tambah", c.formTambah)
	mux.HandleFunc("POST /produk/tambah", c.simpan)
	mux.HandleFunc("GET /produk/{id}", c.detail)
	mux.HandleFunc("GET /produk/{id}/edit", c.formEdit)
	mux.HandleFunc("POST /produk/{id}/edit", c.perbarui)
	mux.HandleFunc("POST /produk/{id}/hapus", c.hapus)

	return Timing(mux)
}

func Timing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mulai := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(mulai).Round(time.Microsecond))
	})
}

func ambilID(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func ambilFilter(r *http.Request) models.Filter {
	q := r.URL.Query()
	return models.Filter{
		Cari:     strings.TrimSpace(q.Get("q")),
		Kategori: strings.TrimSpace(q.Get("kategori")),
		Urutkan:  q.Get("urut"),
		Arah:     q.Get("arah"),
	}
}

func (c *ProdukController) dataDasar(r *http.Request) views.Data {
	stats, err := models.StatistikProduk()
	if err != nil {
		log.Println("gagal menghitung statistik:", err)
	}
	kategori, err := models.KategoriBesertaJumlah()
	if err != nil {
		log.Println("gagal memuat kategori:", err)
	}

	return views.Data{
		Flash:        ambilFlash(r),
		Stats:        stats,
		KategoriStat: kategori,
		KategoriOpsi: entities.SemuaKategori(),
	}
}

func ambilFlash(r *http.Request) *views.Flash {
	q := r.URL.Query()
	if pesan, ok := pesanSukses[q.Get("ok")]; ok {
		return &views.Flash{Jenis: "sukses", Pesan: pesan}
	}
	if pesan := q.Get("galat"); pesan != "" {
		return &views.Flash{Jenis: "error", Pesan: pesan}
	}
	return nil
}

func (c *ProdukController) dashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/dashboard" {
		http.NotFound(w, r)
		return
	}
	c.daftar(w, r, "dashboard")
}

func (c *ProdukController) daftarProduk(w http.ResponseWriter, r *http.Request) {
	c.daftar(w, r, "produk")
}

func (c *ProdukController) daftar(w http.ResponseWriter, r *http.Request, menu string) {
	filter := ambilFilter(r)

	produk, err := models.SemuaProduk(filter)
	if err != nil {
		http.Error(w, "Gagal memuat produk: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := c.dataDasar(r)
	data.Produk = produk
	data.Total = len(produk)
	data.Filter = filter

	if menu == "dashboard" {
		data.Judul = "Dashboard"
		data.Subjudul = "Ringkasan toko dan produk kamu"
		data.Menu = "dashboard"
	} else {
		data.Menu = "produk"
		if filter.Aktif() {
			data.Judul = "Hasil Pencarian"
			data.Subjudul = "Produk yang cocok dengan filter"
		} else {
			data.Judul = "Manajemen Produk"
			data.Subjudul = "Kelola seluruh produk dalam satu tempat"
		}
	}

	views.Render(w, "list", data)
}

func (c *ProdukController) formTambah(w http.ResponseWriter, r *http.Request) {
	data := c.dataDasar(r)
	data.Judul = "Tambah Produk"
	data.Subjudul = "Buat produk baru untuk katalog kamu"
	data.Menu = "tambah"
	data.Aksi = "tambah"
	data.FormAksi = "/produk/tambah"
	data.Form.Kategori = "Elektronik"

	views.Render(w, "form", data)
}

func (c *ProdukController) formEdit(w http.ResponseWriter, r *http.Request) {
	id, ok := ambilID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}

	produk, err := models.CariProduk(id)
	if err != nil {
		c.tidakDitemukan(w, r, err)
		return
	}

	data := c.dataDasar(r)
	data.Judul = "Edit Produk"
	data.Subjudul = "Perbarui data " + produk.Nama
	data.Menu = "produk"
	data.Aksi = "edit"
	data.FormAksi = "/produk/" + strconv.FormatInt(id, 10) + "/edit"
	data.Form = produk

	views.Render(w, "form", data)
}

func (c *ProdukController) detail(w http.ResponseWriter, r *http.Request) {
	id, ok := ambilID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}

	produk, err := models.CariProduk(id)
	if err != nil {
		c.tidakDitemukan(w, r, err)
		return
	}

	data := c.dataDasar(r)
	data.Judul = produk.Nama
	data.Subjudul = "Detail produk " + produk.Kode
	data.Menu = "produk"
	data.Detail = &produk

	views.Render(w, "detail", data)
}

func bacaForm(r *http.Request) (entities.Produk, map[string]string) {
	parse := func(nama string) int {
		angka := strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' {
				return r
			}
			return -1
		}, r.FormValue(nama))
		n, _ := strconv.Atoi(angka)
		return n
	}

	produk := entities.Produk{
		Nama:      strings.TrimSpace(r.FormValue("nama")),
		Kategori:  strings.TrimSpace(r.FormValue("kategori")),
		Harga:     int64(parse("harga")),
		Stok:      parse("stok"),
		Deskripsi: strings.TrimSpace(r.FormValue("deskripsi")),
	}

	validasi := map[string]string{}

	if produk.Nama == "" {
		validasi["nama"] = "Nama produk tidak boleh kosong."
	} else if len([]rune(produk.Nama)) < 3 {
		validasi["nama"] = "Nama produk minimal 3 karakter."
	}

	if produk.Kategori == "" {
		validasi["kategori"] = "Kategori wajib dipilih."
	}

	if produk.Harga <= 0 {
		validasi["harga"] = "Harga harus lebih besar dari 0."
	} else if produk.Harga > 10_000_000_000 {
		validasi["harga"] = "Harga terlalu besar, maksimal Rp 10.000.000.000."
	}

	if produk.Stok < 0 {
		validasi["stok"] = "Stok tidak boleh negatif."
	} else if produk.Stok > 1_000_000 {
		validasi["stok"] = "Stok terlalu banyak, maksimal 1.000.000 unit."
	}

	if len([]rune(produk.Deskripsi)) > 1000 {
		validasi["deskripsi"] = "Deskripsi maksimal 1.000 karakter."
	}

	return produk, validasi
}

func (c *ProdukController) simpan(w http.ResponseWriter, r *http.Request) {
	produk, errs := bacaForm(r)

	if len(errs) > 0 {
		data := c.dataDasar(r)
		data.Judul = "Tambah Produk"
		data.Subjudul = "Perbaiki dulu isian yang ditandai"
		data.Menu = "tambah"
		data.Aksi = "tambah"
		data.FormAksi = "/produk/tambah"
		data.Form = produk
		data.Errors = errs
		w.WriteHeader(http.StatusBadRequest)
		views.Render(w, "form", data)
		return
	}

	if _, err := models.TambahProduk(produk); err != nil {
		http.Error(w, "Gagal menyimpan produk: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/produk?ok=tambah", http.StatusSeeOther)
}

func (c *ProdukController) perbarui(w http.ResponseWriter, r *http.Request) {
	id, ok := ambilID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}

	produk, errs := bacaForm(r)

	if len(errs) > 0 {
		produk.ID = id
		produk.Kode = entities.KodeProduk(id)
		data := c.dataDasar(r)
		data.Judul = "Edit Produk"
		data.Subjudul = "Perbaiki dulu isian yang ditandai"
		data.Menu = "produk"
		data.Aksi = "edit"
		data.FormAksi = "/produk/" + strconv.FormatInt(id, 10) + "/edit"
		data.Form = produk
		data.Errors = errs
		w.WriteHeader(http.StatusBadRequest)
		views.Render(w, "form", data)
		return
	}

	if err := models.UbahProduk(id, produk); err != nil {
		c.tidakDitemukan(w, r, err)
		return
	}

	http.Redirect(w, r, "/produk?ok=edit", http.StatusSeeOther)
}

func (c *ProdukController) hapus(w http.ResponseWriter, r *http.Request) {
	id, ok := ambilID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}

	if err := models.HapusProduk(id); err != nil {
		c.tidakDitemukan(w, r, err)
		return
	}

	http.Redirect(w, r, "/produk?ok=hapus", http.StatusSeeOther)
}

func (c *ProdukController) tidakDitemukan(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, models.ErrTidakDitemukan) {
		http.Redirect(w, r, "/produk?galat=Produk+tidak+ditemukan.", http.StatusSeeOther)
		return
	}
	log.Println("gagal memproses produk:", err)
	http.Error(w, "Terjadi kesalahan: "+err.Error(), http.StatusInternalServerError)
}
