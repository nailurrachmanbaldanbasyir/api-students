package service

import (
	"testing"

	"api-students/app/model"
)

func TestValidateCreateStudent(t *testing.T) {
	req := model.CreateStudentRequest{
		NIM:      "",
		Name:     "",
		Grade:    120,
		IsActive: true,
	}

	errs := ValidateCreateStudent(req)

	if len(errs) != 3 {
		t.Fatalf("expected 3 validation errors, got %d", len(errs))
	}

	if errs["nim"] != "NIM wajib diisi" {
		t.Errorf("unexpected nim error: %v", errs["nim"])
	}

	if errs["name"] != "nama wajib diisi" {
		t.Errorf("unexpected name error: %v", errs["name"])
	}

	if errs["grade"] != "grade harus berada di antara 0 dan 100" {
		t.Errorf("unexpected grade error: %v", errs["grade"])
	}
}

func TestValidateReplaceStudent(t *testing.T) {
	req := model.ReplaceStudentRequest{
		NIM:      "123",
		Name:     "Baldan",
		Grade:    95,
		IsActive: true,
	}

	errs := ValidateReplaceStudent(req)

	if len(errs) != 0 {
		t.Fatalf("expected no validation errors, got %v", errs)
	}
}

func TestApplyPatchStudent(t *testing.T) {
	name := "Baldan Baru"
	grade := 95.0

	req := model.PatchStudentRequest{
		Name:  &name,
		Grade: &grade,
	}

	fields, errs := ApplyPatchStudent(req)

	if len(errs) != 0 {
		t.Fatalf("expected no validation errors, got %v", errs)
	}

	if fields["name"] != "Baldan Baru" {
		t.Errorf("unexpected name field: %v", fields["name"])
	}

	if fields["grade"] != 95.0 {
		t.Errorf("unexpected grade field: %v", fields["grade"])
	}
}
