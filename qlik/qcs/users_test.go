package qcs

import (
	"testing"
)

func TestRoleType_Normalize(t *testing.T) {
	tests := []struct {
		name string
		r    RoleType
	}{
		{"consumer", "consumer"},
		{"consumer", "Consumer"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.Normalize()
			if tt.name != string(tt.r) {
				t.Errorf("output: %s, expected: %s", tt.r, tt.name)
			}
		})
	}
}

func TestRoleType_IsValid(t *testing.T) {
	tests := []struct {
		name string
		r    RoleType
		want bool
	}{
		{"1", "Consumer", true},
		{"2", "consUmer", true},
		{"3", "consume", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
