// Copyright 2026 Greg Mushen and contributors. Licensed under Apache-2.0. See LICENSE.
// Novel command implementation is in calls_task_run.go.
// pp:data-source local

package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gregmushen/pp-bland/internal/store"
	"github.com/spf13/cobra"
)

type blandTaskMatch struct {
	CallID      string          `json:"call_id"`
	Status      string          `json:"status"`
	Task        string          `json:"task"`
	PhoneNumber string          `json:"phone_number,omitempty"`
	Transcript  json.RawMessage `json:"transcripts,omitempty"`
	Summary     string          `json:"summary,omitempty"`
	Error       string          `json:"error_message,omitempty"`
}

func newNovelCallsSearchTaskCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "search-task <words>",
		Short:       "Find prior calls made through this CLI by task wording and review their status and summary.",
		Example:     "  bland-pp-cli calls search-task reservation --agent",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "local", "pp:no-error-path-probe": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "calls search-task")
			}
			if flags.dataSource == "live" {
				return usageErr(fmt.Errorf("calls search-task has no live equivalent; use --data-source local or omit --data-source"))
			}
			if len(args) == 0 {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("search words are required"))
			}
			query := strings.TrimSpace(strings.Join(args, " "))
			dbPath := defaultDBPath("bland-pp-cli")
			db, err := store.OpenReadOnly(dbPath)
			if err != nil {
				return fmt.Errorf("open local call history: %w; first run 'bland-pp-cli calls task run'", err)
			}
			defer db.Close()
			raw, err := db.Search(query, 50, "calls")
			if err != nil {
				return fmt.Errorf("search local call history: %w", err)
			}
			results := make([]blandTaskMatch, 0, len(raw))
			for _, item := range raw {
				var row blandTaskMatch
				if err := json.Unmarshal(item, &row); err != nil {
					continue
				}
				results = append(results, row)
			}
			if !wantsHumanTable(cmd.OutOrStdout(), flags) {
				return printJSONFiltered(cmd.OutOrStdout(), results, flags)
			}
			if len(results) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No locally cached calls match that task.")
				return nil
			}
			for _, row := range results {
				outcome := strings.TrimSpace(row.Summary)
				if outcome == "" {
					outcome = strings.TrimSpace(row.Error)
				}
				outcome = strings.Join(strings.Fields(outcome), " ")
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%s\n", row.CallID, row.Status, row.Task, outcome)
			}
			return nil
		},
	}
	return cmd
}
