package service

import (
	"testing"

	"api-students/app/model"
	"api-students/helper"
)

func TestValidateCreateStudent_Declarative(t *testing.T) {
	req := model.CreateStudentRequest{NIM: "", Name: "", Grade: 120}
	errs := helper.ValidateStruct(req)
	if len(errs) < 3 {
		t.Fatalf("expected >=3 validation errors, got %d: %v", len(errs), errs)
	}
	if errs["nim"] != "wajib diisi" {
		t.Errorf("unexpected nim error: %v", errs["nim"])
	}
	if errs["name"] != "wajib diisi" {
		t.Errorf("unexpected name error: %v", errs["name"])
	}
}

func TestValidatePatch_OmitNil(t *testing.T) {
	empty := ""
	req := model.PatchStudentRequest{Name: &empty}
	errs := helper.ValidateStruct(req)
	if errs["name"] != "minimal 3 karakter" {
		t.Fatalf("expected minimal 3 karakter, got %v", errs["name"])
	}
	// nil should be skipped
	req2 := model.PatchStudentRequest{}
	errs2 := helper.ValidateStruct(req2)
	if len(errs2) != 0 {
		t.Fatalf("expected no errors for nil patch, got %v", errs2)
	}
}

func TestValidateNIM_Custom(t *testing.T) {
	req := model.CreateStudentRequest{NIM: "ab!", Name: "Baldan", Grade: 90}
	errs := helper.ValidateStruct(req)
	if errs["nim"] == "" {
		t.Fatalf("expected nim validation error for ab!, got %v", errs)
	}
}

func TestIsEmptyPatch(t *testing.T) {
	if !IsEmptyPatch(model.PatchStudentRequest{}) {
		t.Error("expected empty patch true")
	}
	name := "x"
	if IsEmptyPatch(model.PatchStudentRequest{Name: &name}) {
		t.Error("expected not empty")
	}
}

func TestApplyPatch(t *testing.T) {
	cur := model.Student{NIM: "12345", Name: "Old", Grade: 80, IsActive: true}
	newName := "Baldan Baru"
	newGrade := 95.0
	req := model.PatchStudentRequest{Name: &newName, Grade: &newGrade}
	merged := ApplyPatch(cur, req)
	if merged.Name != "Baldan Baru" || merged.Grade != 95.0 || merged.NIM != "12345" {
		t.Fatalf("unexpected merge result: %+v", merged)
	}
}
