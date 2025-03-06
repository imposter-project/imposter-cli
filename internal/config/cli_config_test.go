package config

import (
	"gatehill.io/imposter/internal/logging"
	"testing"
)

func Test_checkCliVersion(t *testing.T) {
	logger = logging.GetLogger()

	tests := []struct {
		name          string
		configVersion string
		required      string
		wantErr       bool
	}{
		{
			name:          "dev version",
			configVersion: DevCliVersion,
			required:      "1.0.0",
			wantErr:       false,
		},
		{
			name:          "version meets requirement",
			configVersion: "1.2.0",
			required:      "1.0.0",
			wantErr:       false,
		},
		{
			name:          "version does not meet requirement",
			configVersion: "0.9.0",
			required:      "1.0.0",
			wantErr:       true,
		},
		{
			name:          "invalid config version",
			configVersion: "invalid",
			required:      "1.0.0",
			wantErr:       true,
		},
		{
			name:          "invalid required version",
			configVersion: "1.0.0",
			required:      "invalid",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Config.Version = tt.configVersion
			err := checkCliVersion(tt.required)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkCliVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
