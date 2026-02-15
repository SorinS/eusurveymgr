package cmd

import (
	"eusurveymgr/config"
	"eusurveymgr/log"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
	cfg     *config.Configuration

	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "eusurveymgr",
	Short: "EUSurvey Management CLI",
	Long: `eusurveymgr — EUSurvey Management CLI for surveys, results, PDFs, and database queries.

Environment variables override config file values (avoids exposing credentials):
  EUSURVEYMGR_WEB_USER, EUSURVEYMGR_WEB_PASSWORD
  EUSURVEYMGR_DB_HOST, EUSURVEYMGR_DB_NAME, EUSURVEYMGR_DB_USER, EUSURVEYMGR_DB_PASSWORD`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if verbose {
			log.SetLogLevel(log.Debug)
		}
		// Skip config loading for version command
		if cmd.Name() == "version" {
			return nil
		}
		var err error
		cfg, err = config.LoadFromFile(cfgFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		if verbose {
			config.PrintConfig(cfg)
		}
		return nil
	},
	SilenceUsage: true,
}

var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Print version info",
	Example: "  eusurveymgr version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("eusurveymgr %s (commit %s, built %s)\n", version, commit, buildDate)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "eusurveymgr.json", "Path to config file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose (debug) output")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(surveysCmd)
	rootCmd.AddCommand(resultsCmd)
	rootCmd.AddCommand(pdfCmd)
	rootCmd.AddCommand(dbCmd)
}

func SetVersion(v, c, d string) {
	version = v
	commit = c
	buildDate = d
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}