package validator

func PhoneValid(phone string) bool {
	if phone[0] != '+' || len(phone) != 12 {
		return false
	}

	return true
}
