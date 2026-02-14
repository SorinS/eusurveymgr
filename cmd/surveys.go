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
}

var surveysListCmd = &cobra.Command{
	Use:   "list",
	Short: "List surveys",
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
		fmt.Fprintln(w, "ALIAS\tTITLE\tSTATE\tANSWERS\tSTART\tEND")
		for _, s := range list.Surveys {
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n",
				s.Alias, s.Title, s.State, s.NumAnswers, s.Start, s.End)
		}
		return w.Flush()
	},
}

var surveysInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get survey metadata",
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

		fmt.Printf("Alias:      %s\n", meta.Alias)
		fmt.Printf("Title:      %s\n", meta.Title)
		fmt.Printf("State:      %s\n", meta.State)
		fmt.Printf("Language:   %s\n", meta.Language)
		fmt.Printf("Security:   %s\n", meta.Security)
		fmt.Printf("Answers:    %d\n", meta.NumAnswers)
		fmt.Printf("Contact:    %s\n", meta.Contact)
		fmt.Printf("Created:    %s\n", meta.Created)
		fmt.Printf("Published:  %s\n", meta.Published)
		fmt.Printf("Updated:    %s\n", meta.Updated)
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