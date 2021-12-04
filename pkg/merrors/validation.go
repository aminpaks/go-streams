package merrors

import "github.com/aminpaks/go-streams/pkg/re"

func BuildFiledValidationError(field string, message string) re.JsonObj {
	return re.JsonObj{
		"field":   field,
		"message": message,
	}
}
