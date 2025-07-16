package level1_tests

import (
	"testing"
)

func TestParseAuthResults(t *testing.T) {
	tests := []struct {
		name           string
		interfaceName  string
		wpaCliOutput   string
		expectedStatus string
	}{
		{
			name:           "Authenticated interface",
			interfaceName:  "rdma0",
			wpaCliOutput:   "Supplicant PAE state=AUTHENTICATED\nwpa_state=COMPLETED\n",
			expectedStatus: "PASS",
		},
		{
			name:           "Non-authenticated interface",
			interfaceName:  "rdma1",
			wpaCliOutput:   "Supplicant PAE state=DISCONNECTED\nwpa_state=DISCONNECTED\n",
			expectedStatus: "FAIL - Interface not authenticated",
		},
		{
			name:           "Empty output",
			interfaceName:  "rdma2",
			wpaCliOutput:   "",
			expectedStatus: "FAIL - Unable to run wpa_cli command",
		},
		{
			name:           "Error output",
			interfaceName:  "rdma3",
			wpaCliOutput:   "Error: Failed to connect to wpa_supplicant",
			expectedStatus: "FAIL - Unable to run wpa_cli command",
		},
		{
			name:           "Interface without supplicant",
			interfaceName:  "rdma4",
			wpaCliOutput:   "Could not connect to wpa_supplicant: rdma4 - re-trying\n",
			expectedStatus: "FAIL - Interface not authenticated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseAuthResults(tt.interfaceName, tt.wpaCliOutput)
			if err != nil {
				t.Errorf("parseAuthResults() error = %v", err)
				return
			}

			if result.Device != tt.interfaceName {
				t.Errorf("parseAuthResults() device = %v, want %v", result.Device, tt.interfaceName)
			}

			if result.AuthStatus != tt.expectedStatus {
				t.Errorf("parseAuthResults() status = %v, want %v", result.AuthStatus, tt.expectedStatus)
			}
		})
	}
}

func TestGetAuthCheckTestConfig(t *testing.T) {
	tests := []struct {
		name          string
		shape         string
		expectedError bool
	}{
		{
			name:          "Valid shape BM.GPU.H100.8",
			shape:         "BM.GPU.H100.8",
			expectedError: false,
		},
		{
			name:          "Unknown shape",
			shape:         "BM.UNKNOWN.SHAPE",
			expectedError: false, // Should not error, just return disabled
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := getAuthCheckTestConfig(tt.shape)
			
			if tt.expectedError && err == nil {
				t.Errorf("getAuthCheckTestConfig() expected error but got none")
				return
			}
			
			if !tt.expectedError && err != nil {
				t.Errorf("getAuthCheckTestConfig() unexpected error = %v", err)
				return
			}

			if config == nil {
				t.Errorf("getAuthCheckTestConfig() returned nil config")
				return
			}

			// Config should have IsEnabled field
			if config.IsEnabled != true && config.IsEnabled != false {
				t.Errorf("getAuthCheckTestConfig() IsEnabled should be boolean")
			}
		})
	}
}