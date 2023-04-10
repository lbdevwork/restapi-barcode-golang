package utils

import (
	"fmt"
	"regexp"
)

func ConvertTo13DigitNumber(value string) string {
	for len(value) < 13 {
		value = "0" + value
	}

	regex, err := regexp.Compile(`^\d{13}$`)
	if err != nil {
		fmt.Printf("Error creating regular expression: %v\n", err)
		return "error"
	}

	if regex.MatchString(value) {
		return value
	}
	return "error"
}

func SafeString(value interface{}) string {
	if value != nil {
		return value.(string)
	}
	return ""
}

func SafeFloat64(value interface{}) float64 {
	if value == nil {
		return 0
	}
	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	default:
		return 0
	}
}
