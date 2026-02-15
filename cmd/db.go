package cmd

import (
	"encoding/json"
	"eusurveymgr/db"
	"eusurveymgr/log"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Query the MySQL database directly",
	Long:  "Query the EUSurvey MySQL database directly for surveys, answer sets, and responses.",
}

var dbSurveysCmd = &cobra.Command{
	Use:   "surveys",
	Short: "List surveys from MySQL",
	Long:  "List all surveys from MySQL (latest version per SURVEY_UID, deduplicated).",
	Example: `  eusurveymgr db surveys
  eusurveymgr db surveys --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOut, _ := cmd.Flags().GetBool("json")

		dbconn, err := db.ConnectToMySQL(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
		if err != nil {
			return fmt.Errorf("connecting to MySQL: %w", err)
		}
		defer dbconn.Close()

		surveys, err := db.ListSurveys(dbconn)
		if err != nil {
			return err
		}

		if jsonOut {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(surveys)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tUID\tALIAS\tTITLE\tPUB\tANS\tCREATED")
		for _, s := range surveys {
			created := ""
			if s.Created.Valid {
				created = s.Created.String
			}
			title := s.Title
			if len(title) > 32 {
				title = title[:32] + "…"
			}
			fmt.Fprintf(w, "%d\t%.8s\t%s\t%s\t%v\t%d\t%s\n",
				s.SurveyID, s.SurveyUID, s.Alias, title, s.Published, s.NumAnswers, created)
		}
		return w.Flush()
	},
}

var dbAnswersCmd = &cobra.Command{
	Use:   "answers",
	Short: "List answer sets for a survey",
	Long:  "List all answer sets (respondents) for a survey, showing name and email.",
	Example: `  eusurveymgr db answers --survey 4578
  eusurveymgr db answers --survey 4609 --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		surveyID, _ := cmd.Flags().GetInt64("survey")
		jsonOut, _ := cmd.Flags().GetBool("json")

		dbconn, err := db.ConnectToMySQL(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
		if err != nil {
			return fmt.Errorf("connecting to MySQL: %w", err)
		}
		defer dbconn.Close()

		answers, err := db.ListAnswerSets(dbconn, surveyID)
		if err != nil {
			return err
		}

		if jsonOut {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(answers)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ANSWER_SET_ID\tUNIQUECODE\tDATE\tNAME\tEMAIL")
		for _, a := range answers {
			date := ""
			if a.Date.Valid {
				date = a.Date.String
			}
			name := ""
			if a.Name.Valid {
				name = a.Name.String
			}
			email := ""
			if a.Email.Valid {
				email = a.Email.String
			}
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
				a.AnswerSetID, a.UniqueCode, date, name, email)
		}
		log.Infof("Total: %d answer sets", len(answers))
		return w.Flush()
	},
}

var dbLookupCmd = &cobra.Command{
	Use:     "lookup",
	Short:   "Look up UNIQUECODE by email",
	Long:    "Look up the ANSWER_SET_ID and UNIQUECODE for a respondent by email address.",
	Example: "  eusurveymgr db lookup --email user@example.com --survey 4578",
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		surveyID, _ := cmd.Flags().GetInt64("survey")

		dbconn, err := db.ConnectToMySQL(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
		if err != nil {
			return fmt.Errorf("connecting to MySQL: %w", err)
		}
		defer dbconn.Close()

		answerSetID, uniqueCode, err := db.LookupUniqueCode(dbconn, email, surveyID)
		if err != nil {
			return err
		}

		fmt.Printf("ANSWER_SET_ID: %d\n", answerSetID)
		fmt.Printf("UNIQUECODE:    %s\n", uniqueCode)
		return nil
	},
}

var dbResponsesCmd = &cobra.Command{
	Use:   "responses",
	Short: "Show answers for a respondent",
	Long:  "Show all answer values for a respondent, identified by --email and --survey.",
	Example: `  eusurveymgr db responses --email user@example.com --survey 4578
  eusurveymgr db responses --email user@example.com --survey 4578 --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		surveyID, _ := cmd.Flags().GetInt64("survey")
		jsonOut, _ := cmd.Flags().GetBool("json")

		dbconn, err := db.ConnectToMySQL(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
		if err != nil {
			return fmt.Errorf("connecting to MySQL: %w", err)
		}
		defer dbconn.Close()

		answerSetID, _, err := db.LookupUniqueCode(dbconn, email, surveyID)
		if err != nil {
			return err
		}

		responses, err := db.GetResponses(dbconn, answerSetID)
		if err != nil {
			return err
		}

		if jsonOut {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(responses)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "PA_ID\tQUESTION\tVALUE")
		for _, r := range responses {
			question := ""
			if r.Question.Valid {
				question = r.Question.String
				if len(question) > 40 {
					question = question[:40] + "…"
				}
			}
			value := ""
			if r.Value.Valid {
				value = r.Value.String
			}
			fmt.Fprintf(w, "%d\t%s\t%s\n", r.PA_ID, question, value)
		}
		log.Infof("Total: %d answers (ANSWER_SET_ID=%d)", len(responses), answerSetID)
		return w.Flush()
	},
}

func init() {
	dbSurveysCmd.Flags().Bool("json", false, "JSON output")

	dbAnswersCmd.Flags().Int64("survey", 0, "Survey ID")
	dbAnswersCmd.Flags().Bool("json", false, "JSON output")
	dbAnswersCmd.MarkFlagRequired("survey")

	dbLookupCmd.Flags().String("email", "", "Email address to look up")
	dbLookupCmd.Flags().Int64("survey", 0, "Survey ID")
	dbLookupCmd.MarkFlagRequired("email")
	dbLookupCmd.MarkFlagRequired("survey")

	dbResponsesCmd.Flags().String("email", "", "Respondent email address")
	dbResponsesCmd.Flags().Int64("survey", 0, "Survey ID")
	dbResponsesCmd.Flags().Bool("json", false, "JSON output")
	dbResponsesCmd.MarkFlagRequired("email")
	dbResponsesCmd.MarkFlagRequired("survey")

	dbCmd.AddCommand(dbSurveysCmd)
	dbCmd.AddCommand(dbAnswersCmd)
	dbCmd.AddCommand(dbLookupCmd)
	dbCmd.AddCommand(dbResponsesCmd)
}