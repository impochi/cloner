// Package config handles the configuration of the controller.
package config

import (
	"os"
	"strings"

	"github.com/go-logr/logr"
)

// Config represents the configuration for the controller.
type Config struct {
	IgnoreNamespaces []string
	Logger           logr.Logger
}

// ParseIgnoreNamespaces parses the namespaces string provided by the user
// into a list of strings based on the comma separator.
// Appends the `kube-system` and CONTROLLER_NAMESPACEto the string and removes
// the duplicates.
func (c *Config) ParseIgnoreNamespaces(namespaces string) {
	ns := strings.Split(namespaces, ",")

	// Append controller namespace and kube-system namespace.
	controllerNs := os.Getenv("CONTROLLER_NAMESPACE")
	ns = append(ns, controllerNs, "kube-system")
	ignoredNamespaces := []string{}

	dupl := map[string]string{}

	// Remove duplicate names and empty strings from the namespaces input.
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
