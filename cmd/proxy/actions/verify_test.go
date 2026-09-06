package actions

import "testing"

func TestValidateVerifyFlags(t *testing.T) {
	cases := []struct {
		name          string
		verifyStorage bool
		purge         bool
		wantErr       bool
	}{
		{"neither", false, false, false},
		{"verify only", true, false, false},
		{"verify and purge", true, true, false},
		{"purge without verify", false, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateVerifyFlags(tc.verifyStorage, tc.purge)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
