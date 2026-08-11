package prompts

import "testing"

func TestValidateComposeProjectName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{name: "lowercase letters", value: "bluemedia"},
		{name: "digits hyphens and underscores", value: "blue-media_2"},
		{name: "starts with digit", value: "2blue"},
		{name: "uppercase letters", value: "BlueMedia", wantError: true},
		{name: "starts with hyphen", value: "-bluemedia", wantError: true},
		{name: "contains whitespace", value: "blue media", wantError: true},
		{name: "empty", value: "", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateComposeProjectName(tt.value)
			if (err != nil) != tt.wantError {
				t.Fatalf("validateComposeProjectName(%q) error = %v, wantError = %v", tt.value, err, tt.wantError)
			}
		})
	}
}
