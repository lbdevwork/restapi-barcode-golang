package utils

import (
	"fmt"
	"regexp"
	"strconv"
)

func ConvertTo13DigitNumber(value string) string {
	if value == "" {
		return "error"
	}

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
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			fmt.Printf("Error parsing float from string: %v\n", err)
			return 0
		}
		return f
	default:
		fmt.Printf("Unknown type: %T\n", value)
		return 0
	}
}
