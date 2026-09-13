package service

import (
	"strings"

	"api-students/app/model"
)

func ValidateCreateStudent(req model.CreateStudentRequest) map[string]string {
	errs := make(map[string]string)

	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "NIM wajib diisi"
	}

	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "nama wajib diisi"
	}

	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "grade harus berada di antara 0 dan 100"
	}

	return errs
}

func ValidateReplaceStudent(req model.ReplaceStudentRequest) map[string]string {
	errs := make(map[string]string)

	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "NIM wajib diisi"
	}

	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "nama wajib diisi"
	}

	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "grade harus berada di antara 0 dan 100"
	}

	return errs
}

func ApplyPatchStudent(req model.PatchStudentRequest) (map[string]any, map[string]string) {
	fields := make(map[string]any)
	errs := make(map[string]string)

	if req.NIM != nil {
		if strings.TrimSpace(*req.NIM) == "" {
			errs["nim"] = "NIM tidak boleh kosong"
		} else {
			fields["nim"] = strings.TrimSpace(*req.NIM)
		}
	}

	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			errs["name"] = "nama tidak boleh kosong"
		} else {
			fields["name"] = strings.TrimSpace(*req.Name)
		}
	}

	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 100 {
			errs["grade"] = "grade harus berada di antara 0 dan 100"
		} else {
			fields["grade"] = *req.Grade
		}
	}

	if req.IsActive != nil {
		fields["is_active"] = *req.IsActive
	}

	if len(fields) == 0 && len(errs) == 0 {
		errs["body"] = "minimal satu field harus dikirim"
	}

	if len(errs) > 0 {
		return nil, errs
	}

	return fields, nil
}
