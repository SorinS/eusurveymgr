package cmd

import (
	"encoding/json"
	"eusurveymgr/client"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var surveysCmd = &cobra.Command{
	Use:   "surveys",
	Short: "Manage surveys via WebService API",
	Long:  "Query the EUSurvey WebService API to list surveys and retrieve metadata.",
}

var surveysListCmd = &cobra.Command{
	Use:   "list",
	Short: "List surveys",
	Long:  "List all surveys for the authenticated user via the WebService API.",
	Example: `  eusurveymgr surveys list
  eusurveymgr surveys list --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOut, _ := cmd.Flags().GetBool("json")
		c := client.New(cfg)

		list, err := c.GetSurveys()
		if err != nil {
			return err
		}

		if jsonOut {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(list.Surveys)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ALIAS\tTITLE")
		for _, s := range list.Surveys {
			fmt.Fprintf(w, "%s\t%s\n", s.Alias, s.Title)
		}
		return w.Flush()
	},
}

var surveysInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get survey metadata",
	Long:  "Retrieve detailed metadata for a survey by its alias (shortname).",
	Example: `  eusurveymgr surveys info --alias Check4SkillsInRomana
  eusurveymgr surveys info --alias Check4SkillsInEnglish --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		alias, _ := cmd.Flags().GetString("alias")
		jsonOut, _ := cmd.Flags().GetBool("json")
		c := client.New(cfg)

		meta, err := c.GetSurveyMetadata(alias)
		if err != nil {
			return err
		}

		if jsonOut {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(meta)
		}

		fmt.Printf("ID:         %s\n", meta.ID)
		fmt.Printf("Alias:      %s\n", meta.Alias)
		fmt.Printf("Title:      %s\n", meta.Title)
		fmt.Printf("Type:       %s\n", meta.SurveyType)
		fmt.Printf("Status:     %s\n", meta.Status)
		fmt.Printf("Language:   %s\n", meta.Language)
		fmt.Printf("Security:   %s\n", meta.Security)
		fmt.Printf("Visibility: %s\n", meta.Visibility)
		fmt.Printf("Results:    %d\n", meta.Results)
		fmt.Printf("Contact:    %s\n", meta.Contact)
		fmt.Printf("Start:      %s\n", meta.Start)
		fmt.Printf("End:        %s\n", meta.End)
		return nil
	},
}

func init() {
	surveysListCmd.Flags().Bool("json", false, "JSON output")

	surveysInfoCmd.Flags().String("alias", "", "Survey alias/shortname")
	surveysInfoCmd.Flags().Bool("json", false, "JSON output")
	surveysInfoCmd.MarkFlagRequired("alias")

	surveysCmd.AddCommand(surveysListCmd)
	surveysCmd.AddCommand(surveysInfoCmd)
}