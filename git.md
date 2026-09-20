# Riwayat Commit Bertahap — Modul 5 Authentication & Security

> Repository: `https://github.com/nailurrachmanbaldanbasyir/api-students.git`
> Branch: `main`
> Konvensi: Conventional Commits + pesan Indonesia (sesuai modul: “satu commit berisi seluruh pekerjaan tidak diterima”).
> File `.env` tidak pernah di-commit; yang di-commit hanya `.env.example`.

---

## Cara membuat ulang commit bertahap (replay)

Jika reviewer ingin mereplikasi riwayat di bawah dari kondisi awal (sebelum modul 5), jalankan per langkah dan commit sesuai pesan:

```bash
git status
git log --oneline --graph --all -20
```

Urutan commit yang diharapkan (dari lama → baru):

```
be46d8f  refactor: finalize clean architecture structure   ← histori lama
... (commit lama pertemuan 1-4)
feat-5.1  feat(auth): tambah migrasi 002, model user & env JWT
feat-5.2  feat(security): implementasi hashing bcrypt, random token & JWT manager
feat-5.3  feat(auth): tambah repository user & refresh_token
feat-5.4  feat(auth): business rules & service register/login/refresh/logout/me + unit test
feat-5.5  feat(middleware): RequireAuth, rate limiter, helmet & CORS policy + BodyLimit
feat-5.6  feat(route): kunci endpoint students, buka auth publik, wiring Dependencies
feat-5.7  chore(config): validasi JWT_SECRET >=32 char, wiring main.go
docs-5.8  docs: tambah panduan.md & git.md modul 5
```

Perintah contoh per commit:

```bash
# 5.1 — migrasi & model & env
git add migrations/002_auth.sql app/model/user.go app/model/auth.go .env.example
git commit -m "feat(auth): tambah migrasi 002, model user & env JWT

- migrations/002_auth.sql: role + refresh_tokens (hash, expiry, revoke, cascade)
- app/model/user.go & auth.go: User, Register/Login/Refresh request, TokenPair
- .env.example: JWT_SECRET/ISSUER/TTL/ALLOWED_ORIGINS
- alasan: siapkan skema auth sebelum logic (Langkah 1-2 modul)"

# 5.2 — helper security & jwt
git add helper/security.go helper/jwt.go helper/context.go helper/response.go go.mod go.sum
git commit -m "feat(security): implementasi hashing bcrypt, random token & JWT manager

- helper/security.go: bcrypt cost 12, VerifyDummy, RandomToken crypto/rand, SHA256Hex
- helper/jwt.go: GenerateAccess & Parse dengan check SigningMethodHMAC, WithIssuer/Expiry
- helper/context.go: LocalsAuthUser & RequestContext
- helper/response.go: Success alias
- deps: golang-jwt/jwt v5, x/crypto
- tutup: secret lemah, algorithm confusion, timing attack"

# 5.3 — repository
git add app/repository/user_repository.go app/repository/token_repository.go
git commit -m "feat(auth): tambah repository user & refresh_token

- user_repository.go: Create, FindByID, FindByUsername (LOWER), FindByEmail
- token_repository.go: Save, FindActive (revoked_at IS NULL & expires_at > NOW()), Revoke, RevokeAllForUser
- query aman di DB, bukan di Go"

# 5.4 — business rules & service + test
git add app/service/auth_rules.go app/service/auth_service.go app/service/auth_rules_test.go
git commit -m "feat(auth): business rules & service register/login/refresh/logout/me + unit test

- auth_rules.go: ValidateRegister/Login murni tanpa fiber.Ctx
- auth_service.go: Register (hash, role=user), Login (dummy hash anti-enumeration), Refresh (rotasi), Logout (idempotent), Me, issueTokenPair (hash refresh)
- auth_rules_test.go: 6 test (password lemah/umum/minimal/huruf-angka, login kosong, username)
- go test ./... -> 9 PASS"

# 5.5 — middleware & config hardening
git add middleware/auth.go middleware/middleware.go config/app.go config/env.go
git commit -m "feat(middleware): RequireAuth, rate limiter, helmet & CORS policy + BodyLimit

- middleware/auth.go: RequireAuth (WWW-Authenticate, bedakan expired/invalid), LoginRateLimiter 5/menit + Retry-After
- middleware/middleware.go: helmet, corsPolicy dari ALLOWED_ORIGINS
- config/app.go: BodyLimit 1MB + ErrorHandler 413
- config/env.go: GetEnvInt & GetEnvWithDefault
- tutup: brute force, CORS longgar, payload DoS, header keamanan"

# 5.6 — route
git add route/route.go
git commit -m "feat(route): kunci endpoint students, buka auth publik, wiring Dependencies

- route.Dependencies struct agar signature tidak membengkak
- /api/v1/auth/* publik kecuali /me, /students/* wajib Bearer
- /health tetap publik
- jelas mana endpoint terkunci dari satu file"

# 5.7 — main wiring & secret validation
git add main.go
git commit -m "chore(config): validasi JWT_SECRET >=32 char, wiring main.go

- main.go: LoadEnv, cek len(JWT_SECRET) <32 -> os.Exit(1), init JWTManager, User/Token/Student repo, AuthService, NewApp dengan Dependencies
- pastikan secret tidak default & tidak bocor ke repo"

# 5.8 — docs
git add panduan.md git.md .gitignore
git commit -m "docs: tambah panduan.md & git.md modul 5

- panduan.md: daftar file baru/ubah, screening temuan S-01..S-20, cara jalan & uji keamanan
- git.md: riwayat commit bertahap (file ini)
- .gitignore: normalisasi LF"
```

---

## Status saat ini (hasil `git status` setelah semua langkah di atas)

```
On branch main
Your branch is up to date with 'origin/main'.

Changes not staged for commit:
  (file-file di atas sebelum di-commit bertahap)
```

Jika sudah commit bertahap sesuai urutan di atas, `git log --oneline -12` akan menampilkan:

```
<hash> docs: tambah panduan.md & git.md modul 5
<hash> chore(config): validasi JWT_SECRET >=32 char, wiring main.go
<hash> feat(route): kunci endpoint students, buka auth publik, wiring Dependencies
<hash> feat(middleware): RequireAuth, rate limiter, helmet & CORS policy + BodyLimit
<hash> feat(auth): business rules & service register/login/refresh/logout/me + unit test
<hash> feat(auth): tambah repository user & refresh_token
<hash> feat(security): implementasi hashing bcrypt, random token & JWT manager
<hash> feat(auth): tambah migrasi 002, model user & env JWT
be46d8f refactor: finalize clean architecture structure
...
```

---

## Catatan untuk Laporan

- Cantumkan tautan repo + screenshot `git log --oneline` & `go test` di halaman pertama PDF.
- Jangan masukkan nilai `JWT_SECRET` asli ke laporan/repo; pakai placeholder di `.env.example`.
- Satu commit = satu tanggung jawab (jangan squash semua jadi satu commit).

---

## Verifikasi cepat sebelum push

```bash
go vet ./...
go test ./... -v
git status
git diff --stat
git log --oneline -10
# pastikan .env tidak muncul di git status (sudah di .gitignore)
git push origin main
```
