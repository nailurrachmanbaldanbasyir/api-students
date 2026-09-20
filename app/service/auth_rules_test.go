package service

import (
	"testing"

	"api-students/app/model"
)

func TestValidateRegister_PasswordLemah(t *testing.T) {
	req := model.RegisterRequest{Username: "sari", Email: "sari@example.com", Password: "password1"}
	errs := ValidateRegister(req)
	if errs["password"] != "password terlalu umum" {
		t.Fatalf("expected password terlalu umum, got %v", errs["password"])
	}
}

func TestValidateRegister_PasswordMinimal(t *testing.T) {
	req := model.RegisterRequest{Username: "budi", Email: "budi@example.com", Password: "abc1"}
	errs := ValidateRegister(req)
	if errs["password"] != "minimal 8 karakter" {
		t.Fatalf("expected minimal 8 karakter, got %v", errs["password"])
	}
}

func TestValidateRegister_HarusHurufDanAngka(t *testing.T) {
	req := model.RegisterRequest{Username: "ani", Email: "ani@example.com", Password: "abcdefgh"}
	errs := ValidateRegister(req)
	if errs["password"] != "harus memuat huruf dan angka" {
		t.Fatalf("expected harus memuat huruf dan angka, got %v", errs["password"])
	}
}

func TestValidateRegister_Valid(t *testing.T) {
	req := model.RegisterRequest{Username: "sari", Email: "sari@example.com", Password: "rahasia123"}
	errs := ValidateRegister(req)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
}

func TestValidateLogin_Kosong(t *testing.T) {
	req := model.LoginRequest{Username: "", Password: ""}
	errs := ValidateLogin(req)
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %v", errs)
	}
}

func TestIsValidUsername(t *testing.T) {
	if !isValidUsername("sari_123.test") {
		t.Error("expected valid username")
	}
	if isValidUsername("sari@test!") {
		t.Error("expected invalid username")
	}
}
