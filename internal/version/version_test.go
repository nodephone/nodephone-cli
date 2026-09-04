package version

import (
	"strings"
	"testing"
)

func TestGetInfo(t *testing.T) {
	info := GetInfo()
	if info.Version != "v1.0.0" {
		t.Errorf("expected Version v1.0.0, got %s", info.Version)
	}
	if info.GoVersion == "" {
		t.Error("expected non-empty GoVersion")
	}
	if info.OS == "" {
		t.Error("expected non-empty OS")
	}

	str := info.String()
	if !strings.Contains(str, "NodePhone CLI") {
		t.Errorf("expected String() to contain 'NodePhone CLI', got:\n%s", str)
	}
	if !strings.Contains(str, "Version : v1.0.0") {
		t.Errorf("expected String() to contain 'Version : v1.0.0', got:\n%s", str)
	}
}
