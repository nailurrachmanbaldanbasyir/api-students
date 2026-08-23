package main

import (
	"sort"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

var students = []Student{
	{
		ID:       1,
		NIM:      "434241121",
		Name:     "Baldan",
		Grade:    90,
		IsActive: true,
	},
	{
		ID:       2,
		NIM:      "434241118",
		Name:     "Azzam",
		Grade:    85,
		IsActive: true,
	},
	{
		ID:       3,
		NIM:      "434241074",
		Name:     "Diaul",
		Grade:    78,
		IsActive: false,
	},
}

var nextID = 4

func paramID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))

	if err != nil || id <= 0 {
		return 0, false
	}

	return id, true
}

func findStudentIndex(id int) int {
	for i, student := range students {
		if student.ID == id {
			return i
		}
	}

	return -1
}

func listStudents(c *fiber.Ctx) error {
	q := parseListQuery(c)

	hasil := make([]Student, 0)

	for _, student := range students {
		if q.Search != "" &&
			!strings.Contains(
				strings.ToLower(student.Name),
				strings.ToLower(q.Search),
			) {
			continue
		}

		if q.IsActive != nil &&
			student.IsActive != *q.IsActive {
			continue
		}

		hasil = append(hasil, student)
	}

	sort.Slice(hasil, func(i, j int) bool {
		var less bool

		switch q.Sort {
		case "nim":
			less = hasil[i].NIM < hasil[j].NIM
		case "name":
			less = hasil[i].Name < hasil[j].Name
		case "grade":
			less = hasil[i].Grade < hasil[j].Grade
		case "is_active":
			less = !hasil[i].IsActive && hasil[j].IsActive
		default:
			less = hasil[i].ID < hasil[j].ID
		}

		if q.Order == "desc" {
			return !less
		}

		return less
	})

	total := len(hasil)

	totalPages := 0
	if total > 0 {
		totalPages = (total + q.Limit - 1) / q.Limit
	}

	mulai := (q.Page - 1) * q.Limit

	if mulai > total {
		mulai = total
	}

	akhir := mulai + q.Limit

	if akhir > total {
		akhir = total
	}

	return okList(
		c,
		"daftar student berhasil diambil",
		hasil[mulai:akhir],
		&Meta{
			Page:       q.Page,
			Limit:      q.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	)
}

func getStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)

	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	i := findStudentIndex(id)

	if i == -1 {
		return fail(c, fiber.StatusNotFound, "student tidak ditemukan")
	}

	return ok(c, "student ditemukan", students[i])
}

func createStudent(c *fiber.Ctx) error {
	contentType := strings.ToLower(c.Get("Content-Type"))

	if !strings.HasPrefix(contentType, "application/json") {
		return fail(c, fiber.StatusUnsupportedMediaType, "Content-Type harus application/json")
	}

	var req CreateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "JSON tidak valid")
	}

	if strings.TrimSpace(req.NIM) == "" {
		return failValidation(c, map[string]string{
			"nim": "NIM wajib diisi",
		})
	}

	if strings.TrimSpace(req.Name) == "" {
		return failValidation(c, map[string]string{
			"name": "nama wajib diisi",
		})
	}

	if req.Grade < 0 || req.Grade > 100 {
		return failValidation(c, map[string]string{
			"grade": "grade harus berada di antara 0 dan 100",
		})
	}

	for _, student := range students {
		if student.NIM == req.NIM {
			return fail(c, fiber.StatusConflict, "NIM sudah digunakan")
		}
	}

	student := Student{
		ID:       nextID,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
	}

	students = append(students, student)
	nextID++

	return created(
		c,
		"student berhasil dibuat",
		student,
		"/api/v1/students/"+strconv.Itoa(student.ID),
	)
}

func replaceStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)

	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	i := findStudentIndex(id)

	if i == -1 {
		return fail(c, fiber.StatusNotFound, "student tidak ditemukan")
	}

	contentType := strings.ToLower(c.Get("Content-Type"))

	if !strings.HasPrefix(contentType, "application/json") {
		return fail(c, fiber.StatusUnsupportedMediaType, "Content-Type harus application/json")
	}

	var req ReplaceStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "JSON tidak valid")
	}

	if strings.TrimSpace(req.NIM) == "" {
		return failValidation(c, map[string]string{
			"nim": "NIM wajib diisi",
		})
	}

	if strings.TrimSpace(req.Name) == "" {
		return failValidation(c, map[string]string{
			"name": "nama wajib diisi",
		})
	}

	if req.Grade < 0 || req.Grade > 100 {
		return failValidation(c, map[string]string{
			"grade": "grade harus berada di antara 0 dan 100",
		})
	}

	for _, student := range students {
		if student.ID != id && student.NIM == req.NIM {
			return fail(c, fiber.StatusConflict, "NIM sudah digunakan")
		}
	}

	students[i] = Student{
		ID:       id,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
	}

	return ok(c, "student berhasil diperbarui", students[i])
}

func patchStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)

	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	i := findStudentIndex(id)

	if i == -1 {
		return fail(c, fiber.StatusNotFound, "student tidak ditemukan")
	}

	contentType := strings.ToLower(c.Get("Content-Type"))

	if !strings.HasPrefix(contentType, "application/json") {
		return fail(c, fiber.StatusUnsupportedMediaType, "Content-Type harus application/json")
	}

	var req PatchStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "JSON tidak valid")
	}

	if req.NIM == nil &&
		req.Name == nil &&
		req.Grade == nil &&
		req.IsActive == nil {
		return failValidation(c, map[string]string{
			"body": "minimal satu field harus dikirim",
		})
	}

	if req.NIM != nil {
		if strings.TrimSpace(*req.NIM) == "" {
			return failValidation(c, map[string]string{
				"nim": "NIM tidak boleh kosong",
			})
		}

		for _, student := range students {
			if student.ID != id && student.NIM == *req.NIM {
				return fail(c, fiber.StatusConflict, "NIM sudah digunakan")
			}
		}

		students[i].NIM = *req.NIM
	}

	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return failValidation(c, map[string]string{
				"name": "nama tidak boleh kosong",
			})
		}

		students[i].Name = *req.Name
	}

	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 100 {
			return failValidation(c, map[string]string{
				"grade": "grade harus berada di antara 0 dan 100",
			})
		}

		students[i].Grade = *req.Grade
	}

	if req.IsActive != nil {
		students[i].IsActive = *req.IsActive
	}

	return ok(c, "student berhasil diubah sebagian", students[i])
}

func deleteStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)

	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	i := findStudentIndex(id)

	if i == -1 {
		return fail(c, fiber.StatusNotFound, "student tidak ditemukan")
	}

	students = append(students[:i], students[i+1:]...)

	return noContent(c)
}
