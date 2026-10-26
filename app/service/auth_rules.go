package service

import "api-students/app/model"
import "api-students/helper"

// Deprecated: gunakan helper.ValidateStruct langsung
func ValidateRegister(req model.RegisterRequest) map[string]string {
	return helper.ValidateStruct(req)
}

func ValidateLogin(req model.LoginRequest) map[string]string {
	return helper.ValidateStruct(req)
}
