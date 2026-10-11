# Playbook: Design System

Dimuat oleh: Design System Engineer, Head of UI, Head of Design, UI Designer, Builder FE/Android/Desktop.
Tujuan: mengubah keputusan visual menjadi aset yang bisa dipakai kode: token, komponen dasar, dan aturan pemakaian.

## Daftar isi
1. Struktur berkas · 2. Skema tokens.json · 3. Penamaan token · 4. Tema · 5. Pemetaan ke stack · 6. Komponen · 7. Verifikasi · 8. Tata kelola · 9. Anti-pola

## 1. Struktur berkas
```
design-system/
  MASTER.md          keluaran design engine + keputusan (satu sumber; hasil --persist)
  tokens.json        sumber kebenaran token (dibekukan di G2)
  tokens.css         variabel CSS (web, ekstensi, game, Wails/Tauri/Electron)
  components.md      kontrak komponen
  DESIGN_SYSTEM.md   panduan pemakaian ringkas
```
Ekspor platform (Compose, XAML, Dart, RN) dibuat Builder platform dari `tokens.json`, bukan ditulis tangan di tempat lain.

## 2. Skema `tokens.json` (dibaca `scripts/contrast.py`)
```json
{
  "version": "1.0.0",
  "color": {
    "light": {"canvas": "#FFFFFF", "surface": "#F5F5F7", "content": "#1D1D1F", "content-muted": "#6E6E73", "accent": "#0A66C2", "on-accent": "#FFFFFF"},
    "dark":  {"canvas": "#000000", "surface": "#1C1C1E", "content": "#F5F5F7", "content-muted": "#A1A1A6", "accent": "#4DA3FF", "on-accent": "#001B33"}
  },
  "space": {"1": 4, "2": 8, "3": 12, "4": 16, "6": 24, "8": 32},
  "radius": {"sm": 6, "md": 10, "lg": 16, "pill": 999},
  "font": {"sans": "system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif", "size": {"caption": 12, "body": 16, "title": 22}},
  "motion": {"fast": 150, "base": 200, "slow": 300},
  "pairs": [
    {"fg": "content", "bg": "canvas", "min": 4.5, "label": "teks isi"},
    {"fg": "content-muted", "bg": "surface", "min": 4.5, "label": "teks sekunder"},
    {"fg": "on-accent", "bg": "accent", "min": 4.5, "label": "teks di atas aksen"}
  ]
}
```
Nilai di atas hanya contoh bentuk; nilai nyata berasal dari arah visual yang dipilih (MASTER.md).

## 3. Penamaan token (semantik, bukan nilai)
- Warna: `canvas`, `surface`, `raised`, `overlay`; `content`, `content-muted`, `content-subtle`, `content-inverse`, `on-accent`; `line`, `line-strong`, `focus`; `accent`, `accent-hover`, `accent-soft`; `success`, `warning`, `danger`, `info` (+ varian `-soft`).
- Spasi `space-1..12` (4, 8, 12, 16, 24, 32, 48, 64 ...); radius `sm/md/lg/pill`; tipografi `font-sans`, `font-mono`, ukuran `caption/footnote/body/callout/headline/title/display`; bayangan `shadow-sm/md/lg`; motion `fast/base/slow`; z-index `dropdown/sticky/modal/toast`.
- Nama menyatakan **peran**, bukan warna ("accent", bukan "blue-500").

## 4. Tema
Terang dan gelap sejajar (paritas), semua pasangan kontras lulus di keduanya. CSS: variabel di `:root` (terang), blok `@media (prefers-color-scheme: dark)` untuk `:root:not([data-theme="light"])`, dan `:root[data-theme="dark"]`. Tema kontras tinggi: hormati `forced-colors` di web dan High Contrast di Windows.

## 5. Pemetaan ke stack
- **Web + Tailwind**: lihat versi di `package.json`. v4: definisikan lewat `@theme` (variabel CSS; pakai `@theme inline` bila merujuk variabel runtime); v3: `theme.extend` dengan `var(--...)`. Baca dokumentasi versi terpasang, jangan menebak sintaks.
- **Android**: Compose `MaterialTheme(colorScheme, typography, shapes)` dari token; Flutter `ThemeData/ColorScheme`; React Native modul tema JS.
- **Windows**: WinUI `ResourceDictionary` dengan `ThemeDictionaries` Light/Dark; Wails/Tauri/Electron memakai variabel CSS yang sama dengan web.
- Pemetaan konsep ke Flutter/RN/Tauri/Electron/SwiftUI: `references/cross-platform.md` di Full of Design.

## 6. Komponen
Kontrak tiap komponen di `components.md`: nama, tujuan, varian, ukuran, state, API (props), aksesibilitas (role, label, keyboard, fokus), token yang dipakai, contoh. Minimum: Button, IconButton, Input/Textarea, Select, Checkbox/Radio/Switch, Card, Modal/Sheet, Toast, Tabs, Tabel/List, Navigasi, Badge, Tooltip, Skeleton/Spinner, EmptyState, FormField (dengan galat).

## 7. Verifikasi
- `python "<SKILL_DIR>/scripts/contrast.py" --tokens design-system/tokens.json` hasilnya 0 gagal.
- Lint token: tidak ada hex, warna bernama, atau px lepas di luar token (cari dengan grep di kode UI).
- Galeri komponen (halaman statis) difoto terang dan gelap; tes a11y dasar pada komponen interaktif.

## 8. Tata kelola
Token dan kontrak komponen dibekukan di G2. Hanya Design System Engineer yang menulis `design-system/*` setelah itu, lewat CR. Versi semver di `tokens.json`. Perubahan token wajib menjalankan ulang `contrast.py` dan memberi tahu Builder terdampak.

## 9. Anti-pola
Token bernama menurut nilai; nilai warna di kode komponen; dua sumber token; menambah komponen tanpa kontrak; tema gelap sebagai pembalikan warna otomatis tanpa pengecekan kontras; mencampur dua `MASTER.md`.
