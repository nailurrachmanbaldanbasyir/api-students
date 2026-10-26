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
	return helper.ParamID(c)
}

func translateStudentError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("student tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Conflict("NIM sudah digunakan")
	default:
		return helper.Internal(err)
	}
}

func listStudents(repo *repository.StudentRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Content negotiation SEBELUM query
		format, err := helper.Negotiate(c, helper.FormatJSON, helper.FormatCSV)
		if err != nil {
			return err
		}
		q, err := helper.ParseCursorQuery(c)
		if err != nil {
			return err
		}
		rows, err := repo.FindAfterCursor(c.Context(), q)
		if err != nil {
			return helper.Internal(err)
		}
		hasMore := len(rows) > q.Limit
		if hasMore {
			rows = rows[:q.Limit]
		}
		meta := &model.CursorMeta{Limit: q.Limit, HasMore: hasMore}
		if hasMore && len(rows) > 0 {
			last := rows[len(rows)-1]
			meta.NextCursor = helper.EncodeCursor(last.CreatedAt, last.ID)
		}
		if format == helper.FormatCSV {
			return helper.WriteStudentsCSV(c, rows)
		}
		return helper.SuccessCursor(c, "daftar student berhasil diambil", rows, meta)
	}
}

func getStudent(repo *repository.StudentRepository, perms *helper.PermissionSet) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, valid := paramID(c)
		if !valid {
			return helper.BadRequest("id harus berupa angka positif")
		}
		student, err := repo.GetByID(c.Context(), id)
		if err != nil {
			return translateStudentError(err)
		}
		current, ok := helper.CurrentUser(c)
		if !ok {
			return helper.Unauthorized("belum terautentikasi")
		}
		if !service.CanAccessStudent(current, student.OwnerID, perms, "student:read:any") {
			return helper.Forbidden("tidak memiliki hak mengakses student ini")
		}
		return helper.Success(c, fiber.StatusOK, "student ditemukan", student)
	}
}

func createStudent(repo *repository.StudentRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req model.CreateStudentRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.BadRequest("body harus berupa JSON yang valid")
		}
		if errs := helper.ValidateStruct(req); errs != nil {
			return helper.Validation(errs)
		}
		current, ok := helper.CurrentUser(c)
		if !ok {
			return helper.Unauthorized("belum terautentikasi")
		}
		student := model.Student{
			NIM: strings.TrimSpace(req.NIM), Name: strings.TrimSpace(req.Name),
			Grade: req.Grade, IsActive: req.IsActive, OwnerID: current.UserID,
		}
		result, err := repo.Create(c.Context(), student)
		if err != nil {
			return translateStudentError(err)
		}
		return helper.Created(c, "student berhasil dibuat", result, "/api/v1/students/"+strconv.Itoa(result.ID))
	}
}

func replaceStudent(repo *repository.StudentRepository, perms *helper.PermissionSet) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, valid := paramID(c)
		if !valid {
			return helper.BadRequest("id harus berupa angka positif")
		}
		current, ok := helper.CurrentUser(c)
		if !ok {
			return helper.Unauthorized("belum terautentikasi")
		}
		existing, err := repo.GetByID(c.Context(), id)
		if err != nil {
			return translateStudentError(err)
		}
		if !service.CanAccessStudent(current, existing.OwnerID, perms, "student:update:any") {
			return helper.Forbidden("tidak memiliki hak mengubah student ini")
		}
		var req model.ReplaceStudentRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.BadRequest("body harus berupa JSON yang valid")
		}
		if errs := helper.ValidateStruct(req); errs != nil {
			return helper.Validation(errs)
		}
		student := model.Student{
			NIM: strings.TrimSpace(req.NIM), Name: strings.TrimSpace(req.Name),
			Grade: req.Grade, IsActive: req.IsActive,
		}
		result, err := repo.Update(c.Context(), id, student)
		if err != nil {
			return translateStudentError(err)
		}
		return helper.Success(c, fiber.StatusOK, "student berhasil diperbarui", result)
	}
}

func patchStudent(repo *repository.StudentRepository, perms *helper.PermissionSet) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, valid := paramID(c)
		if !valid {
			return helper.BadRequest("id harus berupa angka positif")
		}
		current, ok := helper.CurrentUser(c)
		if !ok {
			return helper.Unauthorized("belum terautentikasi")
		}
		existing, err := repo.GetByID(c.Context(), id)
		if err != nil {
			return translateStudentError(err)
		}
		if !service.CanAccessStudent(current, existing.OwnerID, perms, "student:update:any") {
			return helper.Forbidden("tidak memiliki hak mengubah student ini")
		}
		var req model.PatchStudentRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.BadRequest("body harus berupa JSON yang valid")
		}
		if service.IsEmptyPatch(req) {
			return helper.BadRequest("minimal satu field harus dikirim")
		}
		if errs := helper.ValidateStruct(req); errs != nil {
			return helper.Validation(errs)
		}
		merged := service.ApplyPatch(existing, req)
		// gunakan Update untuk menyimpan hasil merge
		result, err := repo.Update(c.Context(), id, merged)
		if err != nil {
			return translateStudentError(err)
		}
		return helper.Success(c, fiber.StatusOK, "student berhasil diubah sebagian", result)
	}
}

func deleteStudent(repo *repository.StudentRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, valid := paramID(c)
		if !valid {
			return helper.BadRequest("id harus berupa angka positif")
		}
		if err := repo.Delete(c.Context(), id); err != nil {
			return translateStudentError(err)
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}
