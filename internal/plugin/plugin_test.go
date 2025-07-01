package plugin

import (
	"gatehill.io/imposter/internal/engine"
	"github.com/sirupsen/logrus"
	"testing"
)

func init() {
	logger.SetLevel(logrus.TraceLevel)
}

func TestEnsurePlugin(t *testing.T) {
	type args struct {
		pluginName string
		version    string
		engineType engine.EngineType
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "fetch plugin", args: args{pluginName: "store-redis", engineType: engine.EngineTypeDockerCore, version: "4.2.2"}, wantErr: false},
		{name: "fetch nonexistent plugin version", args: args{pluginName: "store-redis", engineType: engine.EngineTypeDockerCore, version: "0.0.0"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := EnsurePlugin(tt.args.pluginName, tt.args.engineType, tt.args.version); (err != nil) != tt.wantErr {
				t.Errorf("EnsurePlugin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnsurePlugins(t *testing.T) {
	type args struct {
		version    string
		engineType engine.EngineType
	}
	tests := []struct {
		name    string
		args    args
		plugins []string
		wantErr bool
	}{
		{name: "no op if no plugins configured", args: args{engineType: engine.EngineTypeDockerCore, version: "4.2.2"}, plugins: nil, wantErr: false},
		{name: "fetch configured plugins", args: args{engineType: engine.EngineTypeDockerCore, version: "4.2.2"}, plugins: []string{"store-redis"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ensured, err := EnsurePlugins(tt.plugins, tt.args.engineType, tt.args.version, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsurePlugins() error = %v, wantErr %v", err, tt.wantErr)
			}
			if ensured != len(tt.plugins) {
				t.Errorf("EnsurePlugins() wanted %d plugins, ensured: %d", len(tt.plugins), ensured)
			}
		})
	}
}
