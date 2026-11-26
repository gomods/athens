package module

import "testing"

func TestMatchesModifierWithoutVPrefix(t *testing.T) {
	for _, tt := range []struct {
		version, qualifier string
		want               bool
	}{
		{"v17.0.0", "^17.0.0", true},
		{"v7.0.1", "^17.0.0", false},
		{"v1.2.4", "~1.2.3", true},
		{"v1.3.0", "~1.2.3", false},
		{"v4.4.0", "<4.5.2", true},
		{"v5.0.0", "<4.5.2", false},
		{"v1.2.4", "~v1.2.3", true},
		{"v2.0.0", "^v1.2.3", false},
	} {
		t.Run(tt.version+" "+tt.qualifier, func(t *testing.T) {
			if got := matches(tt.version, tt.qualifier); got != tt.want {
				t.Errorf("matches(%q, %q) = %v, want %v", tt.version, tt.qualifier, got, tt.want)
			}
		})
	}
}
