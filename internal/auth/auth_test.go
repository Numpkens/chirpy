package auth

import (
	"net/http"
	"testing"
)

func TestGetAPIKey(t *testing.T) {
	tests := []struct {
		name          string
		headers       http.Header
		expectedKey   string
		expectedError bool
	}{
		{
			name: "Valid ApiKey header",
			headers: http.Header{
				"Authorization": []string{"ApiKey f271c819202a4667a425332c02094c97"},
			},
			expectedKey:   "f271c819202a4667a425332c02094c97",
			expectedError: false,
		},
		{
			name: "Missing Authorization header",
			headers:       http.Header{},
			expectedKey:   "",
			expectedError: true,
		},
		{
			name: "Malformed header - wrong prefix",
			headers: http.Header{
				"Authorization": []string{"Bearer some-token"},
			},
			expectedKey:   "",
			expectedError: true,
		},
		{
			name: "Malformed header - missing key",
			headers: http.Header{
				"Authorization": []string{"ApiKey"},
			},
			expectedKey:   "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := GetAPIKey(tt.headers)
			if (err != nil) != tt.expectedError {
				t.Errorf("GetAPIKey() error = %v, expectedError %v", err, tt.expectedError)
				return
			}
			if key != tt.expectedKey {
				t.Errorf("GetAPIKey() = %v, expected %v", key, tt.expectedKey)
			}
		})
	}
}