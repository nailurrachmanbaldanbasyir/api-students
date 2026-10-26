package service

import (
	"strings"

	"api-students/app/model"
	"api-students/helper"
)

func ApplyPatch(current model.Student, req model.PatchStudentRequest) model.Student {
	if req.NIM != nil {
		current.NIM = strings.TrimSpace(*req.NIM)
	}
	if req.Name != nil {
		current.Name = strings.TrimSpace(*req.Name)
	}
	if req.Grade != nil {
		current.Grade = *req.Grade
	}
	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}
	return current
}

func IsEmptyPatch(req model.PatchStudentRequest) bool {
	return req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil
}

// Wrapper untuk kompatibilitas test lama (deprected, gunakan helper.ValidateStruct)
func ValidateCreateStudent(req model.CreateStudentRequest) map[string]string {
	return helper.ValidateStruct(req)
}
func ValidateReplaceStudent(req model.ReplaceStudentRequest) map[string]string {
	return helper.ValidateStruct(req)
}
func ApplyPatchStudent(req model.PatchStudentRequest) (map[string]any, map[string]string) {
	if IsEmptyPatch(req) {
		return nil, map[string]string{"body": "minimal satu field harus dikirim"}
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return nil, errs
	}
	fields := map[string]any{}
	if req.NIM != nil {
		fields["nim"] = strings.TrimSpace(*req.NIM)
	}
	if req.Name != nil {
		fields["name"] = strings.TrimSpace(*req.Name)
	}
	if req.Grade != nil {
		fields["grade"] = *req.Grade
	}
	if req.IsActive != nil {
		fields["is_active"] = *req.IsActive
	}
	return fields, nil
}
