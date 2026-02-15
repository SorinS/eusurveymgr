package cmd

import (
	"eusurveymgr/client"
	"eusurveymgr/log"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var resultsCmd = &cobra.Command{
	Use:   "results",
	Short: "Export survey results",
	Long:  "Export survey results via the WebService API (async server-side operation).",
}

var resultsExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export survey results to XML",
	Long: `Start an async results export on the server and poll until complete.
Accepts both numeric survey IDs and aliases. Can be slow for large surveys
as the server generates PDFs as a side-effect.`,
	Example: `  eusurveymgr results export --id Check4SkillsInRomana
  eusurveymgr results export --id 4578 --output results-ro.xml
  eusurveymgr results export --id Check4SkillsInEnglish --showids=false`,
	RunE: func(cmd *cobra.Command, args []string) error {
		formID, _ := cmd.Flags().GetString("id")
		outFile, _ := cmd.Flags().GetString("output")
		showIDs, _ := cmd.Flags().GetBool("showids")
		c := client.New(cfg)

		log.Infof("Preparing results export for survey %s...", formID)
		taskID, err := c.PrepareResults(formID, showIDs)
		if err != nil {
			return err
		}
		log.Infof("Export task ID: %s, polling for results...", taskID)

		data, err := c.GetResults(taskID, cfg.TimeoutSeconds)
		if err != nil {
			return err
		}

		output := outFile
		if output == "" {
			output = "results-" + formID + ".xml"
		}
		if err := os.WriteFile(output, data, 0644); err != nil {
			return fmt.Errorf("writing output file: %w", err)
		}

		log.Infof("Results saved to %s (%d bytes)", output, len(data))
		return nil
	},
}

func init() {
	resultsExportCmd.Flags().String("id", "", "Survey/form ID")
	resultsExportCmd.Flags().String("output", "", "Output file (default: results-<id>.xml)")
	resultsExportCmd.Flags().Bool("showids", true, "Include answer set IDs")
	resultsExportCmd.MarkFlagRequired("id")

	resultsCmd.AddCommand(resultsExportCmd)
}