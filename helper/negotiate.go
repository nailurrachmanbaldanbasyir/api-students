package helper

import (
	"encoding/csv"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
)

const (
	FormatJSON = fiber.MIMEApplicationJSON
	FormatCSV  = "text/csv"
)

func Negotiate(c *fiber.Ctx, offered ...string) (string, error) {
	accept := strings.TrimSpace(c.Get(fiber.HeaderAccept))
	if accept == "" || accept == "*/*" {
		return offered[0], nil
	}
	// handle Accept: */* contained within list like "text/html, */*"
	if strings.Contains(accept, "*/*") {
		return offered[0], nil
	}
	chosen := c.Accepts(offered...)
	if chosen == "" {
		return "", NotAcceptable("format yang diminta tidak tersedia, pilih salah satu dari: " + strings.Join(offered, ", "))
	}
	return chosen, nil
}

func WriteStudentsCSV(c *fiber.Ctx, students []model.Student) error {
	c.Set(fiber.HeaderContentType, FormatCSV+"; charset=utf-8")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="students.csv"`)
	var buf strings.Builder
	w := csv.NewWriter(&buf)
	header := []string{"id", "nim", "name", "grade", "is_active", "owner_id", "created_at"}
	if err := w.Write(header); err != nil {
		return Internal(err)
	}
	for _, s := range students {
		row := []string{
			strconv.Itoa(s.ID), s.NIM, s.Name,
			strconv.FormatFloat(s.Grade, 'f', 2, 64),
			strconv.FormatBool(s.IsActive),
			strconv.Itoa(s.OwnerID),
			s.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		}
		if err := w.Write(row); err != nil {
			return Internal(err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return Internal(err)
	}
	return c.SendString(buf.String())
}

func SuccessCursor(c *fiber.Ctx, message string, data any, meta *model.CursorMeta) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
		"meta":    meta,
	})
}
