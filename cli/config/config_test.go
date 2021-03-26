package config_test

import (
	"os"
	"testing"

	"github.com/impochi/cloner/cli/config"
)

const testNamespace = "test-ns"

func TestParseIgnoreNamespaces(t *testing.T) {
	cfg := &config.Config{}

	os.Setenv("CONTROLLER_NAMESPACE", testNamespace)

	cases := []struct {
		namespaces string
		wanted     []string
	}{
		{
			namespaces: "",
			wanted:     []string{"kube-system", testNamespace},
		}, {
			namespaces: "default,kube-system,default",
			wanted:     []string{"default", "kube-system", testNamespace},
		},
		{
			namespaces: "test,,",
			wanted:     []string{"test", "kube-system", testNamespace},
		},
	}

	for _, test := range cases {
		cfg.ParseIgnoreNamespaces(test.namespaces)

		if len(cfg.IgnoreNamespaces) != len(test.wanted) {
			t.Errorf("Expected the length of wanted and IgnoreNamespaces list to be same")
		}

		equal := areEqual(cfg.IgnoreNamespaces, test.wanted)
		if !equal {
			t.Errorf("expected to be equal; got %v, wanted %v", cfg.IgnoreNamespaces, test.wanted)
		}
	}
}

func areEqual(first, second []string) bool {
	if len(first) != len(second) {
		return false
	}

	exists := make(map[string]bool)
	for _, value := range first {
		exists[value] = true
	}

	for _, value := range second {
		if !exists[value] {
			return false
		}
	}

	return true
}
