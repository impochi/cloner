package config

import (
	"os"
	"strings"

	"github.com/go-logr/logr"
)

type Config struct {
	IgnoreNamespaces []string
	Logger           logr.Logger
}

func (c *Config) ParseIgnoreNamespaces(namespaces string) {
	ns := strings.Split(namespaces, ",")

	// Append controller namespace and kube-system namespace
	controllerNs := os.Getenv("CONTROLLER_NAMESPACE")
	ns = append(ns, controllerNs, "kube-system")
	ignoredNamespaces := []string{}

	// Remove duplicate names and empty strings from the namespaces input.
	dupl := map[string]string{}
	for _, value := range ns {
		value := strings.TrimSpace(value)

		if len(value) == 0 {
			continue
		}

		if _, ok := dupl[value]; !ok {
			dupl[value] = value
			ignoredNamespaces = append(ignoredNamespaces, value)
		}
	}

	c.IgnoreNamespaces = ignoredNamespaces
}
