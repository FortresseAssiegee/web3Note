package config

import "testing"

func TestConfigValidation(t *testing.T) {
	if c, err := Parse(nil); err != nil || c.Mode != "demo" {
		t.Fatal("bad defaults")
	}
	for _, args := range [][]string{{"-ttl=0s"}, {"-interval=-1s"}, {"-timeout=no"}, {"-mode=trade"}, {"unexpected"}, {"-unknown=x"}} {
		if _, err := Parse(args); err == nil {
			t.Errorf("accepted %v", args)
		}
	}
}
