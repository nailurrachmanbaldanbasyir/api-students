package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type PortfolioService struct {
	students   *repository.StudentRepository
	portfolios *repository.PortfolioRepository
}

func NewPortfolioService(
	students *repository.StudentRepository,
	portfolios *repository.PortfolioRepository,
) *PortfolioService {
	return &PortfolioService{
		students:   students,
		portfolios: portfolios,
	}
}

func (s *PortfolioService) studentID(c *fiber.Ctx) (int, error) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return 0, helper.BadRequest("ID mahasiswa tidak valid")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	if _, err := s.students.GetByID(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return 0, helper.NotFound("mahasiswa tidak ditemukan")
		}
		return 0, helper.Internal(err)
	}

	return id, nil
}

func (s *PortfolioService) Get(c *fiber.Ctx) error {
	id, err := s.studentID(c)
	if err != nil {
		return err
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	portfolio, err := s.portfolios.Get(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return helper.NotFound("portfolio tidak ditemukan")
		}
		return helper.Internal(err)
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"portfolio berhasil diambil",
		portfolio,
	)
}

func (s *PortfolioService) Put(c *fiber.Ctx) error {
	id, err := s.studentID(c)
	if err != nil {
		return err
	}

	var req model.Portfolio

	decoder := json.NewDecoder(bytes.NewReader(c.Body()))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		return helper.BadRequest(
			"body berisi JSON yang tidak valid atau field tidak dikenal",
		)
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return helper.BadRequest("body harus berisi satu objek JSON")
	}

	if strings.TrimSpace(req.Headline) == "" {
		return helper.BadRequest("headline wajib diisi")
	}

	if len(req.Projects) > 20 {
		return helper.BadRequest("maksimal 20 proyek per portfolio")
	}

	req.StudentID = id

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	if err := s.portfolios.Upsert(ctx, req); err != nil {
		return helper.Internal(err)
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"portfolio berhasil disimpan",
		req,
	)
}

func (s *PortfolioService) AddProject(c *fiber.Ctx) error {
	id, err := s.studentID(c)
	if err != nil {
		return err
	}

	var project model.PortfolioProject

	decoder := json.NewDecoder(bytes.NewReader(c.Body()))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&project); err != nil {
		return helper.BadRequest(
			"body berisi JSON yang tidak valid atau field tidak dikenal",
		)
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return helper.BadRequest("body harus berisi satu objek JSON")
	}

	project.Title = strings.TrimSpace(project.Title)
	if project.Title == "" {
		return helper.BadRequest("judul proyek wajib diisi")
	}

	if project.ID == "" {
		project.ID, err = helper.RandomToken(16)
		if err != nil {
			return helper.Internal(err)
		}
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	if err := s.portfolios.AddProject(ctx, id, project); err != nil {
		switch {
		case errors.Is(err, repository.ErrPortfolioNotFound):
			return helper.NotFound("portfolio tidak ditemukan")
		case errors.Is(err, repository.ErrPortfolioLimit):
			return helper.Conflict("maksimal 20 proyek per portfolio")
		default:
			return helper.Internal(err)
		}
	}

	return helper.Created(
		c,
		"proyek berhasil ditambahkan",
		project,
		"",
	)
}

func (s *PortfolioService) DeleteProject(c *fiber.Ctx) error {
	id, err := s.studentID(c)
	if err != nil {
		return err
	}

	projectID := strings.TrimSpace(c.Params("projectId"))
	if projectID == "" {
		return helper.BadRequest("ID proyek wajib diisi")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	if err := s.portfolios.DeleteProject(ctx, id, projectID); err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return helper.NotFound("portfolio tidak ditemukan")
		}

		if strings.Contains(err.Error(), "proyek tidak ditemukan") {
			return helper.NotFound("proyek tidak ditemukan")
		}

		return helper.Internal(err)
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"proyek berhasil dihapus",
		nil,
	)
}

func (s *PortfolioService) TechnologyStats(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	stats, err := s.portfolios.TechnologyStats(ctx)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"statistik teknologi berhasil diambil",
		stats,
	)
}
