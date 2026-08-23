# API Students

REST API sederhana untuk mengelola data mahasiswa menggunakan Go dan Fiber.

## Teknologi

- Go
- Fiber v2
- REST API
- JSON

## Base URL

http://localhost:3000/api/v1

---

# API Contract

## 1. GET List Students

### Endpoint

GET `/students`

### Query Parameters

| Parameter | Tipe | Default | Keterangan |
|---|---|---:|---|
| page | integer | 1 | Nomor halaman |
| limit | integer | 10 | Jumlah data per halaman |
| search | string | - | Pencarian berdasarkan nama |
| sort | string | id | Kolom untuk sorting |
| order | string | asc | Urutan asc atau desc |
| is_active | boolean | - | Filter berdasarkan status aktif |

### Contoh Request

GET `/api/v1/students?page=1&limit=10`

### Response

```json
{
  "success": true,
  "message": "daftar student berhasil diambil",
  "data": [
    {
      "id": 1,
      "nim": "434241121",
      "name": "Baldan",
      "grade": 90,
      "is_active": true
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 1,
    "total_pages": 1
  }
}