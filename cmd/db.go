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
}

var dbSurveysCmd = &cobra.Command{
	Use:   "surveys",
	Short: "List surveys from MySQL",
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
		fmt.Fprintln(w, "ID\tTITLE\tSHORTNAME\tPUBLISHED\tANSWERS\tCREATED")
		for _, s := range surveys {
			created := ""
			if s.Created.Valid {
				created = s.Created.String
			}
			fmt.Fprintf(w, "%d\t%s\t%s\t%v\t%d\t%s\n",
				s.SurveyID, s.Title, s.Shortname, s.Published, s.NumAnswers, created)
		}
		return w.Flush()
	},
}

var dbAnswersCmd = &cobra.Command{
	Use:   "answers",
	Short: "List answer sets for a survey",
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
		fmt.Fprintln(w, "ANSWER_SET_ID\tUNIQUECODE\tDATE\tEMAIL")
		for _, a := range answers {
			date := ""
			if a.Date.Valid {
				date = a.Date.String
			}
			email := ""
			if a.Email.Valid {
				email = a.Email.String
			}
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n",
				a.AnswerSetID, a.UniqueCode, date, email)
		}
		log.Infof("Total: %d answer sets", len(answers))
		return w.Flush()
	},
}

var dbLookupCmd = &cobra.Command{
	Use:   "lookup",
	Short: "Look up UNIQUECODE by email",
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

func init() {
	dbSurveysCmd.Flags().Bool("json", false, "JSON output")

	dbAnswersCmd.Flags().Int64("survey", 0, "Survey ID")
	dbAnswersCmd.Flags().Bool("json", false, "JSON output")
	dbAnswersCmd.MarkFlagRequired("survey")

	dbLookupCmd.Flags().String("email", "", "Email address to look up")
	dbLookupCmd.Flags().Int64("survey", 0, "Survey ID")
	dbLookupCmd.MarkFlagRequired("email")
	dbLookupCmd.MarkFlagRequired("survey")

	dbCmd.AddCommand(dbSurveysCmd)
	dbCmd.AddCommand(dbAnswersCmd)
	dbCmd.AddCommand(dbLookupCmd)
}