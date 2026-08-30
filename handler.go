package main

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
)

func paramID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))

	if err != nil || id <= 0 {
		return 0, false
	}

	return id, true
}

func listStudents(repo *repository.StudentRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		q := parseListQuery(c)

		params := repository.ListParams{
			Page:     q.Page,
			Limit:    q.Limit,
			Search:   q.Search,
			Sort:     q.Sort,
			Order:    q.Order,
			IsActive: q.IsActive,
		}

		students, total, err := repo.List(c.Context(), params)
		if err != nil {
			return fail(
				c,
				fiber.StatusInternalServerError,
				"gagal mengambil daftar student",
			)
		}

		totalPages := 0
		if total > 0 {
			totalPages = (total + q.Limit - 1) / q.Limit
		}

		return okList(
			c,
			"daftar student berhasil diambil",
			students,
			&Meta{
				Page:       q.Page,
				Limit:      q.Limit,
				Total:      total,
				TotalPages: totalPages,
			},
		)
	}
}

func getStudent(repo *repository.StudentRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, valid := paramID(c)

		if !valid {
			return fail(
				c,
				fiber.StatusBadRequest,
				"id harus berupa angka positif",
			)
		}

		student, err := repo.GetByID(c.Context(), id)

		if errors.Is(err, repository.ErrNotFound) {
			return fail(
				c,
				fiber.StatusNotFound,
				"student tidak ditemukan",
			)
		}

		if err != nil {
			return fail(
				c,
				fiber.StatusInternalServerError,
				"gagal mengambil student",
			)
		}

		return ok(
			c,
			"student ditemukan",
			student,
		)
	}
}

func createStudent(repo *repository.StudentRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req model.CreateStudentRequest

		if err := c.BodyParser(&req); err != nil {
			return fail(
				c,
				fiber.StatusBadRequest,
				"JSON tidak valid",
			)
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

		student := model.Student{
			NIM:      strings.TrimSpace(req.NIM),
			Name:     strings.TrimSpace(req.Name),
			Grade:    req.Grade,
			IsActive: req.IsActive,
		}

		result, err := repo.Create(c.Context(), student)

		if errors.Is(err, repository.ErrDuplicate) {
			return fail(
				c,
				fiber.StatusConflict,
				"NIM sudah digunakan",
			)
		}

		if err != nil {
			return fail(
				c,
				fiber.StatusInternalServerError,
				"gagal membuat student",
			)
		}

		return created(
			c,
			"student berhasil dibuat",
			result,
			"/api/v1/students/"+strconv.Itoa(result.ID),
		)
	}
}

func replaceStudent(repo *repository.StudentRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, valid := paramID(c)

		if !valid {
			return fail(
				c,
				fiber.StatusBadRequest,
				"id harus berupa angka positif",
			)
		}

		var req model.ReplaceStudentRequest

		if err := c.BodyParser(&req); err != nil {
			return fail(
				c,
				fiber.StatusBadRequest,
				"JSON tidak valid",
			)
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

		student := model.Student{
			NIM:      strings.TrimSpace(req.NIM),
			Name:     strings.TrimSpace(req.Name),
			Grade:    req.Grade,
			IsActive: req.IsActive,
		}

		result, err := repo.Update(
			c.Context(),
			id,
			student,
		)

		if errors.Is(err, repository.ErrNotFound) {
			return fail(
				c,
				fiber.StatusNotFound,
				"student tidak ditemukan",
			)
		}

		if errors.Is(err, repository.ErrDuplicate) {
			return fail(
				c,
				fiber.StatusConflict,
				"NIM sudah digunakan",
			)
		}

		if err != nil {
			return fail(
				c,
				fiber.StatusInternalServerError,
				"gagal memperbarui student",
			)
		}

		return ok(
			c,
			"student berhasil diperbarui",
			result,
		)
	}
}

func patchStudent(repo *repository.StudentRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, valid := paramID(c)

		if !valid {
			return fail(
				c,
				fiber.StatusBadRequest,
				"id harus berupa angka positif",
			)
		}

		var req model.PatchStudentRequest

		if err := c.BodyParser(&req); err != nil {
			return fail(
				c,
				fiber.StatusBadRequest,
				"JSON tidak valid",
			)
		}

		if req.NIM == nil &&
			req.Name == nil &&
			req.Grade == nil &&
			req.IsActive == nil {
			return failValidation(c, map[string]string{
				"body": "minimal satu field harus dikirim",
			})
		}

		fields := make(map[string]any)

		if req.NIM != nil {
			if strings.TrimSpace(*req.NIM) == "" {
				return failValidation(c, map[string]string{
					"nim": "NIM tidak boleh kosong",
				})
			}

			fields["nim"] = strings.TrimSpace(*req.NIM)
		}

		if req.Name != nil {
			if strings.TrimSpace(*req.Name) == "" {
				return failValidation(c, map[string]string{
					"name": "nama tidak boleh kosong",
				})
			}

			fields["name"] = strings.TrimSpace(*req.Name)
		}

		if req.Grade != nil {
			if *req.Grade < 0 || *req.Grade > 100 {
				return failValidation(c, map[string]string{
					"grade": "grade harus berada di antara 0 dan 100",
				})
			}

			fields["grade"] = *req.Grade
		}

		if req.IsActive != nil {
			fields["is_active"] = *req.IsActive
		}

		result, err := repo.Patch(
			c.Context(),
			id,
			fields,
		)

		if errors.Is(err, repository.ErrNotFound) {
			return fail(
				c,
				fiber.StatusNotFound,
				"student tidak ditemukan",
			)
		}

		if errors.Is(err, repository.ErrDuplicate) {
			return fail(
				c,
				fiber.StatusConflict,
				"NIM sudah digunakan",
			)
		}

		if err != nil {
			return fail(
				c,
				fiber.StatusInternalServerError,
				"gagal mengubah student",
			)
		}

		return ok(
			c,
			"student berhasil diubah sebagian",
			result,
		)
	}
}

func deleteStudent(repo *repository.StudentRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, valid := paramID(c)

		if !valid {
			return fail(
				c,
				fiber.StatusBadRequest,
				"id harus berupa angka positif",
			)
		}

		err := repo.Delete(c.Context(), id)

		if errors.Is(err, repository.ErrNotFound) {
			return fail(
				c,
				fiber.StatusNotFound,
				"student tidak ditemukan",
			)
		}

		if err != nil {
			return fail(
				c,
				fiber.StatusInternalServerError,
				"gagal menghapus student",
			)
		}

		return noContent(c)
	}
}
