package utils

import (
	"fmt"
	"regexp"
)

func convertTo13DigitNumber(value string) string {
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

func safeString(value interface{}) string {
	if value != nil {
		return value.(string)
	}
	return ""
}