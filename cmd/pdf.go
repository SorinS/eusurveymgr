package cmd

import (
	"eusurveymgr/client"
	"eusurveymgr/db"
	"eusurveymgr/log"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var pdfCmd = &cobra.Command{
	Use:   "pdf",
	Short: "Download PDF documents",
	Long:  "Download survey form PDFs or generate and download individual answer PDFs.",
}

var pdfSurveyCmd = &cobra.Command{
	Use:   "survey",
	Short: "Download survey form PDF",
	Long:  "Download the survey form as a PDF file via the WebService API.",
	Example: `  eusurveymgr pdf survey --alias Check4SkillsInRomana
  eusurveymgr pdf survey --alias Check4SkillsInEnglish --output survey-en.pdf`,
	RunE: func(cmd *cobra.Command, args []string) error {
		alias, _ := cmd.Flags().GetString("alias")
		outFile, _ := cmd.Flags().GetString("output")
		c := client.New(cfg)

		log.Infof("Downloading survey PDF for %s...", alias)
		data, err := c.GetSurveyPDF(alias)
		if err != nil {
			return err
		}

		output := outFile
		if output == "" {
			output = alias + ".pdf"
		}
		if err := os.WriteFile(output, data, 0644); err != nil {
			return fmt.Errorf("writing PDF: %w", err)
		}

		log.Infof("Survey PDF saved to %s (%d bytes)", output, len(data))
		return nil
	},
}

var pdfAnswerCmd = &cobra.Command{
	Use:   "answer",
	Short: "Generate and download answer PDF",
	Long: `Generate and download an answer PDF for a specific respondent.

Provide either --code for a known UNIQUECODE, or --email and --survey
to look up the code from the database. Skips generation if the PDF
already exists on the server.`,
	Example: `  eusurveymgr pdf answer --code ae8d5fec-daaf-4aba-b860-544d1f717d8a
  eusurveymgr pdf answer --email user@example.com --survey 4578
  eusurveymgr pdf answer --email user@example.com --survey 4578 --output ./pdfs`,
	RunE: func(cmd *cobra.Command, args []string) error {
		code, _ := cmd.Flags().GetString("code")
		email, _ := cmd.Flags().GetString("email")
		surveyID, _ := cmd.Flags().GetInt64("survey")
		outDir, _ := cmd.Flags().GetString("output")
		c := client.New(cfg)

		var uniqueCode string
		var answerSetID int64
		var emailAddr string

		if code != "" {
			uniqueCode = code
		} else if email != "" {
			if surveyID == 0 {
				return fmt.Errorf("--survey is required when using --email")
			}
			dbconn, err := db.ConnectToMySQL(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
			if err != nil {
				return fmt.Errorf("connecting to DB for UNIQUECODE lookup: %w", err)
			}
			defer dbconn.Close()

			emailAddr = email
			answerSetID, uniqueCode, err = db.LookupUniqueCode(dbconn, email, surveyID)
			if err != nil {
				return err
			}
			log.Infof("Found ANSWER_SET_ID=%d UNIQUECODE=%s", answerSetID, uniqueCode)
		} else {
			return fmt.Errorf("provide either --code or --email (with --survey)")
		}

		// Check if PDF already exists before triggering generation
		ready, err := c.IsAnswerPDFReady(uniqueCode)
		if err != nil {
			return fmt.Errorf("checking PDF readiness: %w", err)
		}

		if !ready {
			log.Infof("Triggering PDF generation for %s...", uniqueCode)
			if err := c.CreateAnswerPDF(uniqueCode); err != nil {
				return err
			}

			log.Infof("Waiting for PDF to be ready...")
			deadline := time.Now().Add(time.Duration(cfg.TimeoutSeconds) * time.Second)
			delay := time.Second
			for {
				ready, err = c.IsAnswerPDFReady(uniqueCode)
				if err != nil {
					return fmt.Errorf("checking PDF readiness: %w", err)
				}
				if ready {
					break
				}
				if time.Now().After(deadline) {
					return fmt.Errorf("PDF generation timed out after %ds", cfg.TimeoutSeconds)
				}
				log.Debugf("PDF not ready yet, retrying in %v...", delay)
				time.Sleep(delay)
				if delay < 5*time.Second {
					delay += time.Second
				}
			}
		} else {
			log.Infof("PDF already exists for %s", uniqueCode)
		}

		log.Infof("Downloading PDF...")
		data, err := c.DownloadAnswerPDF(uniqueCode)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}

		var filename string
		if emailAddr != "" {
			filename = fmt.Sprintf("%d--%s.pdf", answerSetID, emailAddr)
		} else {
			filename = uniqueCode + ".pdf"
		}
		outPath := filepath.Join(outDir, filename)

		if err := os.WriteFile(outPath, data, 0644); err != nil {
			return fmt.Errorf("writing PDF: %w", err)
		}

		log.Infof("Answer PDF saved to %s (%d bytes)", outPath, len(data))
		return nil
	},
}

func init() {
	pdfSurveyCmd.Flags().String("alias", "", "Survey alias/shortname")
	pdfSurveyCmd.Flags().String("output", "", "Output file (default: <alias>.pdf)")
	pdfSurveyCmd.MarkFlagRequired("alias")

	pdfAnswerCmd.Flags().String("code", "", "Answer UNIQUECODE")
	pdfAnswerCmd.Flags().String("email", "", "Respondent email address")
	pdfAnswerCmd.Flags().Int64("survey", 0, "Survey ID (required with --email)")
	pdfAnswerCmd.Flags().String("output", ".", "Output directory")

	pdfCmd.AddCommand(pdfSurveyCmd)
	pdfCmd.AddCommand(pdfAnswerCmd)
}