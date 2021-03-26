package cmd

import (
	"flag"

	"github.com/impochi/cloner/cli/config"
	"github.com/impochi/cloner/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	ignoreNamespaces string
)

func Execute() {
	flag.StringVar(&ignoreNamespaces, "ignore-namespaces", "kube-system", "Namespaces to ignore when cloning images")

	opts := zap.Options{}
	opts.BindFlags(flag.CommandLine)

	flag.Parse()

	logger := zap.New(zap.UseFlagOptions(&opts))

	// Create config
	cfg := &config.Config{}

	cfg.ParseIgnoreNamespaces(ignoreNamespaces)
	cfg.Logger = logger

	manager.Run(cfg)
}
