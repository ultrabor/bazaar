package validator

import "testing"

func TestPhoneValid(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		want  bool
	}{
		{name: "valid", phone: "+998901234567", want: true},
		{name: "missing plus", phone: "998901234567", want: false},
		{name: "letters", phone: "+99890abcdefg", want: false},
		{name: "non ASCII digits", phone: "+99890123456٧", want: false},
		{name: "too short", phone: "+99890123456", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PhoneValid(tt.phone); got != tt.want {
				t.Fatalf("PhoneValid(%q) = %v, want %v", tt.phone, got, tt.want)
			}
		})
	}
}
