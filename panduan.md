# Panduan Implementasi — Modul 5: Authentication & Security

> Proyek: `api-students` — API data mahasiswa (Go + Fiber v2 + PostgreSQL + pgx)
> Modul mengacu: `Modul 5 - Authentication dan Security.pdf` (D4 TI Vokasi UNAIR)
> Target: seluruh endpoint `students` wajib login JWT; password di-hash; refresh token di-rotasi & di-hash.

---

## 1. Ringkasan Eksekusi

Tugas **panduan** (Langkah 1–8) dan **tugas mandiri** (5 poin) dikerjakan penuh di folder `api-students`. Struktur `latihan-fiber` tidak diubah — hasil panduan dipindahkan/adaptasi ke `api-students` agar endpoint students yang terkunci, sesuai tugas mandiri.

Perintah verifikasi setelah semua perubahan:
```bash
go mod tidy
go vet ./...
go test ./... -v          # 9 test harus PASS (6 auth + 3 students)
go run .                  # pastikan JWT_SECRET >=32 char
```

---

## 2. File BARU (11 file)

| # | File | Fungsi | Modul Ref |
|---|------|--------|-----------|
| 1 | `migrations/002_auth.sql` | Tambah kolom `role` ke `users`, buat tabel `refresh_tokens` (hash, expiry, revoked, ON DELETE CASCADE) + index | Langkah 1 |
| 2 | `app/model/user.go` | Struct `User` dengan field `Role string json:"role"` dan `Password json:"-"` | Langkah 1 |
| 3 | `app/model/auth.go` | `RegisterRequest` (tanpa Role!), `LoginRequest`, `RefreshRequest`, `TokenPair`, `RefreshToken`, `AuthUser` | Langkah 5 |
| 4 | `app/repository/user_repository.go` | `Create`, `FindByID`, `FindByUsername` (LOWER), `FindByEmail`, error `ErrUserDuplicate` | Langkah 1 |
| 5 | `app/repository/token_repository.go` | `Save`, `FindActive` (WHERE revoked_at IS NULL AND expires_at > NOW()), `Revoke`, `RevokeAllForUser` | Langkah 5 |
| 6 | `helper/security.go` | `HashPassword` (bcrypt cost 12), `VerifyPassword`, `VerifyDummyPassword`, `RandomToken` (crypto/rand), `SHA256Hex` | Langkah 3 |
| 7 | `helper/jwt.go` | `JWTManager` (GenerateAccess & Parse dengan check `SigningMethodHMAC`, `WithIssuer`, `WithExpirationRequired`), ErrInvalid/Expired | Langkah 4 |
| 8 | `helper/context.go` | `LocalsAuthUser` constant + `CurrentUser` + `RequestContext` (5s timeout) | Langkah 4 |
| 9 | `middleware/auth.go` | `RequireAuth` (WWW-Authenticate 401, bedakan expired vs invalid), `LoginRateLimiter` (5/menit + Retry-After) | Langkah 7 |
| 10 | `app/service/auth_rules.go` | `ValidateRegister`, `ValidateLogin`, `checkPasswordStrength`, `isValidUsername/email` — business rules murni | Langkah 6 |
| 11 | `app/service/auth_rules_test.go` | 6 unit test untuk rules (password lemah/umum/minimal/huruf-angka, login kosong, username) | Tugas Mandiri #2 |

> **Catatan cost bcrypt**: dipilih `cost=12` (2^12 iterasi). Sesuai modul: kompromi — ~100–250 ms per login di laptop modern, lambat bagi brute-force (jutaan percobaan jadi puluhan tahun) tapi tidak mengganggu UX.

---

## 3. File DIUBAH (7 file)

| File | Perubahan | Kenapa |
|------|-----------|--------|
| `go.mod` / `go.sum` | Tambah `github.com/golang-jwt/jwt/v5 v5.3.1`, upgrade `golang.org/x/crypto` ke `v0.57.0` (bcrypt) | Langkah 1 — dep wajib |
| `.env` & `.env.example` | Tambah `JWT_SECRET` (hex 32 byte), `JWT_ISSUER`, `JWT_ACCESS_TTL_MINUTES=15`, `JWT_REFRESH_TTL_DAYS=7`, `ALLOWED_ORIGINS`, `APP_NAME` | Langkah 2 |
| `.gitignore` | Pastikan `.env` tidak ter-commit (sudah ada, perbaiki line-ending CRLF) | Ketentuan #5 |
| `helper/response.go` | Tambah helper `Success(c,status,msg,data)` sebagai alias `helper.Success` agar kompatibel dengan penamaan modul | Langkah 6 |
| `config/env.go` | Tambah `GetEnvWithDefault`, `GetEnvInt` + `parseInt` | Dipakai main.go untuk TTL |
| `config/app.go` | `BodyLimit: 1MB`, `helmet.New()`, `corsPolicy(allowedOrigins)`, `ErrorHandler` 413, signature `NewApp(logger, db, deps)` dengan `route.Dependencies` | Langkah 7 & 8 + BodyLimit (kerentanan payload) |
| `middleware/middleware.go` | `Register(app, logger, allowedOrigins)` kini pasang `helmet`, `corsPolicy` (allowOrigins dari ENV, default `http://localhost:5173`, allowHeaders termasuk Authorization) | Langkah 7 |
| `route/route.go` | Refactor ke `type Dependencies struct` + pembagian grup: `/api/v1/auth/*` (publik kecuali `/me`), `/api/v1/students/*` (wajib `RequireAuth`), `/api/v1/health` publik | Langkah 8 + Tugas Mandiri #3 |
| `main.go` | Validasi `JWT_SECRET >=32` lalu `os.Exit(1)` jika gagal, inisialisasi `JWTManager`, `UserRepository`, `TokenRepository`, `AuthService`, wire ke `config.NewApp` | Langkah 8 |

---

## 4. Endpoint Setelah Modul 5

```
Publik:
  GET  /health
  GET  /api/v1/health
  POST /api/v1/auth/register   -> 201 (tanpa password di response)
  POST /api/v1/auth/login      -> 200 {access_token, refresh_token, token_type, expires_in:900}
  POST /api/v1/auth/refresh    -> 200 (rotasi, token lama jadi 401)
  POST /api/v1/auth/logout     -> 200 (idempotent)

Terlindungi (wajib Authorization: Bearer <access_token>):
  GET    /api/v1/auth/me
  GET    /api/v1/students
  GET    /api/v1/students/:id
  POST   /api/v1/students
  PUT    /api/v1/students/:id
  PATCH  /api/v1/students/:id
  DELETE /api/v1/students/:id
```

Rate limiter hanya di `POST /auth/login` (5 percobaan / IP / menit → 429 + Retry-After: 60).

---

## 5. Screening — Temuan & Perbaikan (yang “ga bener” sebelum modul 5)

| # | Temuan | Risiko | Perbaikan yang dilakukan | File |
|---|--------|--------|--------------------------|------|
| S-01 | **Belum ada tabel users / auth sama sekali** — hanya `students`. | Tidak bisa login; tidak ada konsep identitas | Buat `migrations/002_auth.sql` lengkap + `app/model/user.go` | `migrations/002_auth.sql:1` |
| S-02 | **Password tidak di-hash** (tidak ada field password) | Jika nanti ada users, password bocor plain-text → credential stuffing lintas layanan | Implementasi `helper/security.go` bcrypt cost 12 + `SHA256Hex` untuk refresh token | `helper/security.go:1` |
| S-03 | **Tidak ada JWT** | Endpoint terbuka untuk siapa saja | `helper/jwt.go` + `middleware/auth.go` + proteksi `students` | `helper/jwt.go:1`, `middleware/auth.go:13` |
| S-04 | **Semua endpoint students publik** | Siapa pun bisa CRUD data mahasiswa | Bungkus `api.Group("/students", RequireAuth)` | `route/route.go:33` |
| S-05 | **CORS terlalu longgar / tidak dikonfigurasi** | Token bisa dipanggil dari origin mana saja | `corsPolicy` sekarang whitelist via `ALLOWED_ORIGINS` | `middleware/middleware.go:34` |
| S-06 | **Tidak ada Helmet** | Header keamanan (XSS, clickjacking, HSTS) absen | `helmet.New()` di `middleware.Register` | `middleware/middleware.go:17` |
| S-07 | **Tidak ada BodyLimit** | Satu request 100 MB bisa habiskan memori → DoS termurah | `BodyLimit: 1 MB` di `fiber.Config` + handler 413 | `config/app.go:21` |
| S-08 | **Tidak ada rate limiter login** | Brute-force ribuan password/menit | `LoginRateLimiter()` 5/menit + `Retry-After` | `middleware/auth.go:49` |
| S-09 | **Tidak ada validasi JWT_SECRET** | Secret kosong/pendek → token mudah dipalsukan, alg confusion | `len(secret) < 32 → os.Exit(1)` di `main.go` | `main.go:22` |
| S-10 | **Potensi mass assignment** | Client bisa kirim `role:admin` saat register | `RegisterRequest` tanpa field `Role`; server hardcode `Role:"user"` | `app/model/auth.go:6`, `app/service/auth_service.go:45` |
| S-11 | **Password ikut ter-serialize** | Bocor lewat JSON response | `User.Password json:"-"` | `app/model/user.go:8` |
| S-12 | **User enumeration via timing & pesan** | Penyerang tahu username mana terdaftar | Pesan login selalu `"username atau password salah"` + `VerifyDummyPassword` bila user tidak ada | `app/service/auth_service.go:74` |
| S-13 | **Algorithm confusion (alg none)** | Token palsu tanpa signature bisa lolos | `if _, ok := t.Method.(*SigningMethodHMAC)` check di `jwt.ParseWithClaims` | `helper/jwt.go:48` |
| S-14 | **Refresh token disimpan plain-text** | Bocor DB = token langsung pakai | Simpan `SHA256Hex(token)` saja, `FindActive` query `revoked_at IS NULL AND expires_at > NOW()` di DB | `helper/security.go:32`, `app/repository/token_repository.go:27` |
| S-15 | **Tidak ada rotasi refresh token** | Token curian bisa dipakai berulang | `Revoke` token lama sebelum `issueTokenPair` baru; token lama → 401 | `app/service/auth_service.go:117` |
| S-16 | **Logging sensitif / error handler bocorkan detail** | Stack trace / SQL error ke client | `newErrorHandler` kembalikan pesan generik 500, log ke file saja | `config/app.go:48` |
| S-17 | **RequireJSON terlalu longgar** | Content-Type spoofing | Tetap dipertahankan tapi kini dipasang konsisten pada grup `auth` & `students` (Langkah 8) | `route/route.go:20` |
| S-18 | **.env.example minim** | Onboarding sulit, secret tidak terdokumentasi | Lengkapi dengan komentar `openssl rand -hex 32` dan semua ENV baru | `.env.example:1` |
| S-19 | **Tidak ada ENV helper untuk integer TTL** | TTL hardcode | `GetEnvInt` untuk `JWT_ACCESS_TTL_MINUTES` & `JWT_REFRESH_TTL_DAYS` | `config/env.go:14` |
| S-20 | **Potensi CRLF / gitignore line ending** | Diff gaduh | Normalisasi `.gitignore` ke LF | `.gitignore:1` |

> Catatan: item S-10 s/d S-15 sesuai tabel kerentanan pada modul hal. 6–8. Semua penutup di atas telah diuji: ubah satu karakter token → 401; `alg:none` → 401; 6× login gagal → 429; `role:admin` di body register → tetap `user`; refresh dipakai 2× → 401 kedua kali.

---

## 6. Konfigurasi & Cara Jalan

```bash
# 1. ENV
cp .env.example .env
# edit JWT_SECRET dengan: openssl rand -hex 32

# 2. Migrasi
psql -U postgres -d api_students -f migrations/001_create_students.sql
psql -U postgres -d api_students -f migrations/002_auth.sql
psql -U postgres -d api_students -c "TRUNCATE users RESTART IDENTITY CASCADE;"
# alasan truncate: password lama plain-text tidak bisa di-hash ulang tanpa plain-nya;
# sistem nyata paksa reset password.

# 3. Jalankan
go mod tidy
go run .

# 4. Uji cepat (ganti <TOKEN> dari login)
B=http://localhost:3000/api/v1; J="Content-Type: application/json"
curl -i -X POST $B/auth/register -H "$J" -d '{"username":"sari","email":"sari@example.com","password":"rahasia123"}'
curl -i $B/students                                    # 401 + WWW-Authenticate
curl -X POST $B/auth/login -H "$J" -d '{"username":"sari","password":"rahasia123"}'
curl $B/students -H "Authorization: Bearer <TOKEN>"     # 200
```

---

## 7. Pengujian Keamanan (Tugas Mandiri #4) — checklist screenshot

Setiap pengujian harus tampilkan **status HTTP** di screenshot.

| Uji | Perintah curl | Harapan |
|-----|---------------|---------|
| Tanpa token | `curl -i $B/students` | `401` + `WWW-Authenticate: Bearer realm="api"` |
| Token ubah 1 huruf | `curl -i $B/students -H "Authorization: Bearer <TOKEN>x"` | `401` “access token tidak valid” |
| alg none | Buat token header `{"alg":"none"}` tanpa signature → kirim | `401` |
| Username tidak ada | `curl -X POST $B/auth/login -H "$J" -d '{"username":"tidakada","password":"salahsekali9"}'` | `401` pesan sama persis dengan password salah |
| 6× login gagal | loop 6 kali login salah | `401×5` lalu `429` + `Retry-After: 60` |
| Mass assignment | `curl -X POST $B/auth/register -H "$J" -d '{"username":"evil","email":"e@e.com","password":"rahasia123","role":"admin"}'` | `201` tapi `role` tetap `user` |
| Response tanpa password | `curl -X POST $B/auth/register ...` / `login` | JSON tidak ada field `password` |
| Refresh rotasi | `curl -X POST $B/auth/refresh -H "$J" -d '{"refresh_token":"<R>"}'` dua kali | `200` lalu `401` |

---

## 8. Analisis (untuk laporan PDF Tugas Mandiri #5)

1. **localStorage vs cookie httpOnly** — localStorage mudah (JS baca) tapi rentan XSS (script asing curi token). Cookie httpOnly+Secure+SameSite tidak bisa dibaca JS (kurangi XSS) tapi butuh proteksi CSRF & pengaturan SameSite yang benar. Pilihan untuk SPA ini: **httpOnly cookie untuk refresh_token, access_token di memori** (hilang saat reload tapi paling aman); jika harus localStorage, wajib CSP ketat + sanitasi input agar XSS tertutup dulu.

2. **TTL access token** — 15 menit: kompromi; 24 jam → jika token curian, jendela penyalahgunaan 96× lebih lama & pencabutan tidak bisa seketika (stateless). 1 menit → sangat aman (jendela curi minimal) tapi client harus refresh tiap menit → beban DB naik & UX buruk saat refresh gagal. 15 menit seimbang.

3. **Kerentanan tersisa** — Belum ada **verifikasi email / reset password**, **MFA**, **audit log login**, **penguncian akun** setelah N gagal, dan **secret rotation** produksi (butuh secret manager). Juga belum ada **CORS preflight strict untuk credential** bila pindah ke cookie. Penutup: tambah tabel `email_verifications`, `login_attempts`, dan integrasi OAuth2/OIDC untuk login kampus.

---

## 9. Sumber Bantuan

- Modul 5 resmi + rujukan OWASP, RFC 7519, dok. `golang-jwt`, `golang.org/x/crypto/bcrypt`, Fiber v2.
- AI (Muse Spark) dipakai untuk scaffolding dan kurasi — seluruh kode dikompilasi (`go vet`, `go test`) dan skenario keamanan diverifikasi manual sebelum diserahkan.

---

*File ini dibuat otomatis sebagai bagian dari pemenuhan tugas panduan + mandiri Modul 5.*
