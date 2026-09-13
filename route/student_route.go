package route

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/app/service"
	"api-students/helper"
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
		q := helper.ParseListQuery(c)

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
			return helper.Fail(
				c,
				fiber.StatusInternalServerError,
				"gagal mengambil daftar student",
			)
		}

		totalPages := 0
		if total > 0 {
			totalPages = (total + q.Limit - 1) / q.Limit
		}

		return helper.OKList(
			c,
			"daftar student berhasil diambil",
			students,
			&helper.Meta{
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
			return helper.Fail(
				c,
				fiber.StatusBadRequest,
				"id harus berupa angka positif",
			)
		}

		student, err := repo.GetByID(c.Context(), id)

		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(
				c,
				fiber.StatusNotFound,
				"student tidak ditemukan",
			)
		}

		if err != nil {
			return helper.Fail(
				c,
				fiber.StatusInternalServerError,
				"gagal mengambil student",
			)
		}

		return helper.OK(
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
			return helper.Fail(
				c,
				fiber.StatusBadRequest,
				"JSON tidak valid",
			)
		}

		validationErrors := service.ValidateCreateStudent(req)
		if len(validationErrors) > 0 {
			return helper.FailValidation(c, validationErrors)
		}

		student := model.Student{
			NIM:      strings.TrimSpace(req.NIM),
			Name:     strings.TrimSpace(req.Name),
			Grade:    req.Grade,
			IsActive: req.IsActive,
		}

		result, err := repo.Create(c.Context(), student)

		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Fail(
				c,
				fiber.StatusConflict,
				"NIM sudah digunakan",
			)
		}

		if err != nil {
			return helper.Fail(
				c,
				fiber.StatusInternalServerError,
				"gagal membuat student",
			)
		}

		return helper.Created(
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
			return helper.Fail(
				c,
				fiber.StatusBadRequest,
				"id harus berupa angka positif",
			)
		}

		var req model.ReplaceStudentRequest

		if err := c.BodyParser(&req); err != nil {
			return helper.Fail(
				c,
				fiber.StatusBadRequest,
				"JSON tidak valid",
			)
		}

		validationErrors := service.ValidateReplaceStudent(req)
		if len(validationErrors) > 0 {
			return helper.FailValidation(c, validationErrors)
		}

		student := model.Student{
			NIM:      strings.TrimSpace(req.NIM),
			Name:     strings.TrimSpace(req.Name),
			Grade:    req.Grade,
			IsActive: req.IsActive,
		}

		result, err := repo.Update(c.Context(), id, student)

		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(
				c,
				fiber.StatusNotFound,
				"student tidak ditemukan",
			)
		}

		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Fail(
				c,
				fiber.StatusConflict,
				"NIM sudah digunakan",
			)
		}

		if err != nil {
			return helper.Fail(
				c,
				fiber.StatusInternalServerError,
				"gagal memperbarui student",
			)
		}

		return helper.OK(
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
			return helper.Fail(
				c,
				fiber.StatusBadRequest,
				"id harus berupa angka positif",
			)
		}

		var req model.PatchStudentRequest

		if err := c.BodyParser(&req); err != nil {
			return helper.Fail(
				c,
				fiber.StatusBadRequest,
				"JSON tidak valid",
			)
		}

		fields, validationErrors := service.ApplyPatchStudent(req)

		if len(validationErrors) > 0 {
			return helper.FailValidation(c, validationErrors)
		}

		result, err := repo.Patch(c.Context(), id, fields)

		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(
				c,
				fiber.StatusNotFound,
				"student tidak ditemukan",
			)
		}

		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Fail(
				c,
				fiber.StatusConflict,
				"NIM sudah digunakan",
			)
		}

		if err != nil {
			return helper.Fail(
				c,
				fiber.StatusInternalServerError,
				"gagal mengubah student",
			)
		}

		return helper.OK(
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
			return helper.Fail(
				c,
				fiber.StatusBadRequest,
				"id harus berupa angka positif",
			)
		}

		err := repo.Delete(c.Context(), id)

		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(
				c,
				fiber.StatusNotFound,
				"student tidak ditemukan",
			)
		}

		if err != nil {
			return helper.Fail(
				c,
				fiber.StatusInternalServerError,
				"gagal menghapus student",
			)
		}

		return helper.NoContent(c)
	}
}
