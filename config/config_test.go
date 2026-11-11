package config

import (
	"fmt"
	"os"
	"tcp-server/logger"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {

	config, err := LoadConfig()
	if err != nil {
		t.Fail()
	}
	if config.MaxConnections != 16 {
		t.Fatalf("default max conn should be 16")
	}
	logger.LOG.Info(fmt.Sprintf("%v", config))

}

func TestLoadConfigOverrideDefault(t *testing.T) {
	os.Setenv("MAX_CONNS", "10")
	config, err := LoadConfig()
	if err != nil {
		t.Fail()
	}
	if config.MaxConnections != 10 {
		t.Fatalf("overriden max conn should be 10")
	}
	logger.LOG.Info(fmt.Sprintf("%v", config))
}
