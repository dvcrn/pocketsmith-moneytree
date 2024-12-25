package sanitizier

import (
	"testing"
)

func TestSanitize(t *testing.T) {

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"Alpha full width", "ＭｙＴｅｓｔＭｅｔｈｏｄ", "MyTestMethod"},
		{"Spaces in the end", "Test          Transaction         ", "Test Transaction"},
		{"Normal Japanese with iD", "ビ－・エフ・シ－ イベント/iD", "ビ－・エフ・シ－ イベント /iD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sanitize(tt.input); got != tt.want {
				t.Errorf("Sanitize() = %v, want %v", got, tt.want)
			}
		})
	}
}
