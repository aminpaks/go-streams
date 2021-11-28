package merrors

func BuildFiledValidationError(field string, message string) *map[string]interface{} {
	return &map[string]interface{}{
		"field":   field,
		"message": message,
	}
}
