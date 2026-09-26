const root = document.documentElement;
const body = document.body;

const simpanTema = (tema) => {
  root.setAttribute("data-theme", tema);
  try { localStorage.setItem("tema", tema); } catch (e) {}
};

const temaTersimpan = () => {
  try {
    const simpanan = localStorage.getItem("tema");
    if (simpanan) return simpanan;
  } catch (e) {}
  return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
};

simpanTema(temaTersimpan());

document.querySelectorAll("[data-theme-toggle]").forEach((tombol) => {
  tombol.addEventListener("click", () => {
    simpanTema(root.getAttribute("data-theme") === "dark" ? "light" : "dark");
  });
});

const sidebar = document.getElementById("sidebar");
document.querySelectorAll("[data-sidebar-toggle]").forEach((tombol) => {
  tombol.addEventListener("click", () => sidebar.classList.toggle("open"));
});

document.addEventListener("click", (event) => {
  if (!sidebar) return;
  const DalamSidebar = sidebar.contains(event.target);
  const TombolSidebar = event.target.closest("[data-sidebar-toggle]");
  if (sidebar.classList.contains("open") && !DalamSidebar && !TombolSidebar) {
    sidebar.classList.remove("open");
  }
});

document.querySelectorAll("[data-autosubmit]").forEach((elemen) => {
  let tunda;
  const kirim = () => {
    if (elemen.form) elemen.form.submit();
  };
  elemen.addEventListener("input", () => {
    clearTimeout(tunda);
    tunda = setTimeout(kirim, 450);
  });
  elemen.addEventListener("change", kirim);
});

document.addEventListener("keydown", (event) => {
  if (event.key === "/" && !/^(INPUT|TEXTAREA|SELECT)$/.test(document.activeElement.tagName)) {
    const cari = document.querySelector('.topbar-search input[name="q"]');
    if (cari) {
      event.preventDefault();
      cari.focus();
    }
  }
  if (event.key === "Escape") {
    tutupModal();
    if (sidebar && sidebar.classList.contains("open")) sidebar.classList.remove("open");
  }
});

const formatRupiah = (angka) => String(angka).replace(/\B(?=(\d{3})+(?!\d))/g, ".");
const angkaDariTeks = (teks) => parseInt(String(teks).replace(/[^\d]/g, ""), 10) || 0;

document.querySelectorAll("[data-rupiah]").forEach((input) => {
  input.addEventListener("input", () => {
    const posisi = input.value.length - input.selectionStart;
    input.value = formatRupiah(angkaDariTeks(input.value));
    const baru = Math.max(0, input.value.length - posisi);
    input.setSelectionRange(baru, baru);
    perbaruiPratinjau();
  });
  input.addEventListener("focus", () => input.select());
});

const formProduk = document.querySelector(".form");
const target = (nama) => document.querySelector(`[data-preview="${nama}"]`);

const warnaKategori = {
  Elektronik: ["violet", "violet"], Makanan: ["amber", "amber"], Minuman: ["sky", "sky"],
  Busana: ["pink", "pink"], Skincare: ["rose", "rose"], Buku: ["indigo", "indigo"],
  Olahraga: ["lime", "lime"], Rumah: ["emerald", "emerald"], Peralatan: ["cyan", "cyan"],
};
const emojiKategori = {
  Elektronik: "\u{1F4F1}", Makanan: "\u{1F354}", Minuman: "\u{1F964}", Busana: "\u{1F455}",
  Skincare: "\u{1F484}", Buku: "\u{1F4DA}", Olahraga: "\u{1F3BE}", Rumah: "\u{1F3E0}",
  Peralatan: "\u{1F527}", Lainnya: "\u{1F4E6}",
};

function statusStok(stok) {
  if (stok <= 0) return { kelas: "habis", label: "Stok Habis" };
  if (stok <= 5) return { kelas: "menipis", label: "Stok Menipis" };
  return { kelas: "aman", label: "Stok Aman" };
}

function perbaruiPratinjau() {
  if (!formProduk) return;

  const nama = target("nama");
  const harga = target("harga");
  const stok = target("stok");
  const kategori = target("kategori");
  const deskripsi = target("deskripsi");
  if (!nama || !harga) return;

  const nilaiHarga = angkaDariTeks(harga.value);
  const nilaiStok = angkaDariTeks(stok ? stok.value : 0);
  const nilaiKategori = kategori ? kategori.value : "";
  const warna = warnaKategori[nilaiKategori] || ["slate", "slate"];

  nama.textContent = formProduk.nama.value.trim() || "Nama Produk";
  harga.textContent = "Rp " + formatRupiah(nilaiHarga);
  if (stok) stok.textContent = formatRupiah(nilaiStok);

  if (kategori) {
    kategori.textContent = nilaiKategori || "-";
    kategori.className = "badge badge-" + warna[1];
  }

  const status = statusStok(nilaiStok);
  const pil = document.querySelector("[data-preview-pill]");
  const bar = document.querySelector("[data-preview-bar]");
  if (pil) {
    pil.textContent = status.label;
    pil.className = "pill pill-" + status.kelas;
  }
  if (bar) {
    bar.parentElement.className = "progress progress-" + status.kelas;
    bar.style.width = Math.min(100, nilaiStok * 2) + "%";
  }

  const nilai = document.querySelector("[data-preview-nilai]");
  if (nilai) nilai.textContent = "Rp " + formatRupiah(nilaiHarga * nilaiStok);

  const thumb = document.querySelector("[data-preview-emoji]");
  if (thumb) {
    thumb.textContent = emojiKategori[nilaiKategori] || "\u{1F4E6}";
    thumb.className = "thumb thumb-lg thumb-" + warna[0];
  }

  if (deskripsi) {
    deskripsi.textContent = formProduk.deskripsi.value.trim() || "Deskripsi produk akan tampil di sini.";
  }
}

if (formProduk) {
  formProduk.addEventListener("input", perbaruiPratinjau);
  formProduk.addEventListener("change", perbaruiPratinjau);
  perbaruiPratinjau();
}

document.querySelectorAll("[data-counter-for]").forEach((elemen) => {
  const field = document.getElementById(elemen.dataset.counterFor);
  if (!field) return;
  const perbarui = () => { elemen.textContent = field.value.length; };
  field.addEventListener("input", perbarui);
  perbarui();
});

const modal = document.getElementById("modal-hapus");
let formHapus = null;

const bukaModal = (nama, form) => {
  if (!modal) return;
  formHapus = form;
  const judul = document.getElementById("modal-hapus-nama");
  if (judul) judul.textContent = nama;
  modal.hidden = false;
  document.body.style.overflow = "hidden";
};

const tutupModal = () => {
  if (!modal || modal.hidden) return;
  modal.hidden = true;
  formHapus = null;
  document.body.style.overflow = "";
};

document.querySelectorAll("form[data-hapus]").forEach((form) => {
  form.addEventListener("submit", (event) => {
    if (form.dataset.confirmed === "1") return;
    event.preventDefault();
    bukaModal(form.dataset.nama || "produk ini", form);
  });
});

document.querySelectorAll("[data-modal-close]").forEach((elemen) => {
  elemen.addEventListener("click", tutupModal);
});

const tombolYa = document.getElementById("modal-hapus-ya");
if (tombolYa) {
  tombolYa.addEventListener("click", () => {
    if (!formHapus) return;
    formHapus.dataset.confirmed = "1";
    formHapus.submit();
  });
}

const toast = document.querySelector(".toast");
if (toast) {
  const tutup = () => {
    toast.style.transition = "opacity .3s ease, transform .3s ease";
    toast.style.opacity = "0";
    toast.style.transform = "translateX(28px)";
    setTimeout(() => toast.remove(), 320);
  };
  const tombolTutup = toast.querySelector("[data-toast-close]");
  if (tombolTutup) tombolTutup.addEventListener("click", tutup);
  setTimeout(tutup, 4200);
}

body.classList.add("siap");
