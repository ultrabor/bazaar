package validator

import "unicode"

func PhoneValid(phone string) bool {
	if len(phone) != 13 || phone[0] != '+' {
		return false
	}
	for _, char := range phone[1:] {
		if !unicode.IsDigit(char) || char > unicode.MaxASCII {
			return false
		}
	}

	return true
}
