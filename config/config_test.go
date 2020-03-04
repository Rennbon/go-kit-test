package config

import "testing"

func TestDecodeConfig(t *testing.T) {
	cnf, err := DecodeConfig("./config.toml")
	if err != nil {
		t.Error(err)
	}
	t.Log(cnf)
}
