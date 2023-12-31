package action

import (
	"github.com/urfave/cli/v2"
)

// flagScanDepth pass val to urfave flag.
func flagLogLevel(v *string) *cli.StringFlag {
	return &cli.StringFlag{
		Name:        "loglevel",
		Usage:       "configure log level",
		Value:       "info",
		EnvVars:     []string{"HOUSEKEEPER_LOG_LEVEL"},
		Destination: v,
	}
}

// flagScanDepth pass val to urfave flag.
func flagScanDepth(v *int) *cli.IntFlag {
	return &cli.IntFlag{
		Name:        "scandepth",
		Usage:       "configure git scan depth",
		Value:       10,
		EnvVars:     []string{"HOUSEKEEPER_SCAN_DEPTH"},
		Destination: v,
	}
}

// flagDryRun pass val to urfave flag.
func flagDryRun(v *bool) *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:        "dry-run",
		Usage:       "run without dangerous activity",
		Value:       false,
		EnvVars:     []string{"HOUSEKEEPER_DRY_RUN"},
		Destination: v,
	}
}

// flagDryRun pass val to urfave flag.
func flagProtected(v *bool) *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:        "protected",
		Usage:       "set image protected",
		Value:       false,
		EnvVars:     []string{"HOUSEKEEPER_SET_PROTECTED"},
		Destination: v,
	}
}

// flagDryRun pass val to urfave flag.
func flagHidden(v *bool) *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:        "hidden",
		Usage:       "set image hidden",
		Value:       false,
		EnvVars:     []string{"HOUSEKEEPER_SET_HIDDEN"},
		Destination: v,
	}
}
