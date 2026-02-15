package cmd

import (
	"encoding/json"
	"eusurveymgr/client"
	"eusurveymgr/log"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var tokensCmd = &cobra.Command{
	Use:   "tokens",
	Short: "Manage invitation tokens (BROKEN — endpoints don't exist upstream)",
	Long: `Manage invitation tokens for surveys.

WARNING: These commands are currently broken. The upstream EUSurvey source
has no getTokens or createToken endpoints. Token management uses a group-based
API that requires a different implementation.`,
}

var tokensListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List invitation tokens",
	Example: "  eusurveymgr tokens list --survey Check4SkillsInRomana",
	RunE: func(cmd *cobra.Command, args []string) error {
		survey, _ := cmd.Flags().GetString("survey")
		jsonOut, _ := cmd.Flags().GetBool("json")
		c := client.New(cfg)

		list, err := c.GetTokens(survey)
		if err != nil {
			return err
		}

		if jsonOut {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(list.Tokens)
		}

		for _, t := range list.Tokens {
			fmt.Println(t.Value)
		}
		log.Infof("Total: %d tokens", len(list.Tokens))
		return nil
	},
}

var tokensCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create new invitation token",
	Example: "  eusurveymgr tokens create --survey Check4SkillsInRomana",
	RunE: func(cmd *cobra.Command, args []string) error {
		survey, _ := cmd.Flags().GetString("survey")
		c := client.New(cfg)

		token, err := c.CreateToken(survey)
		if err != nil {
			return err
		}

		fmt.Println(token)
		return nil
	},
}

func init() {
	tokensListCmd.Flags().String("survey", "", "Survey name/alias")
	tokensListCmd.Flags().Bool("json", false, "JSON output")
	tokensListCmd.MarkFlagRequired("survey")

	tokensCreateCmd.Flags().String("survey", "", "Survey name/alias")
	tokensCreateCmd.MarkFlagRequired("survey")

	tokensCmd.AddCommand(tokensListCmd)
	tokensCmd.AddCommand(tokensCreateCmd)
}