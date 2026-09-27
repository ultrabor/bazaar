package validator

func PhoneValid(phone string) bool {
	if len(phone) != 13 || phone[0] != '+' {
		return false
	}

	return true
}
