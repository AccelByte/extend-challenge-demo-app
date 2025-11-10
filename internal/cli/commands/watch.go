// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/api"
	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/cli"
	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/cli/output"
	"github.com/spf13/cobra"
)

// NewWatchCommand creates the watch command
func NewWatchCommand() *cobra.Command {
	var interval time.Duration
	var challengeID string
	var once bool

	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Continuously monitor challenges",
		Long:  "Watch challenges and output updates at regular intervals.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get format flag
			format, _ := cmd.Flags().GetString("format")

			// Create container
			container := cli.GetContainerFromFlags(cmd)

			ctx := context.Background()
			formatter := output.NewFormatter(format)

			// Setup signal handling for Ctrl+C
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			var prevChallenges []api.Challenge

			// Helper to fetch and print
			fetchAndPrint := func() error {
				challenges, err := container.APIClient.ListChallenges(ctx)
				if err != nil {
					return err
				}

				// Filter if specific challenge requested
				if challengeID != "" {
					filtered := []api.Challenge{}
					for _, c := range challenges {
						if c.ID == challengeID {
							filtered = append(filtered, c)
						}
					}
					challenges = filtered
				}

				// Detect changes (simple comparison)
				changeCount := 0
				if len(prevChallenges) > 0 {
					changeCount = detectChangeCount(prevChallenges, challenges)
				}

				// Format and print
				result, err := formatter.FormatChallenges(challenges)
				if err != nil {
					return err
				}

				// Print timestamp and change info (text mode only)
				if format == "text" || format == "" {
					fmt.Printf("[%s] ", time.Now().Format("2006-01-02 15:04:05"))
					if len(prevChallenges) > 0 {
						if changeCount > 0 {
							fmt.Printf("%d change(s) detected\n", changeCount)
						} else {
							fmt.Println("No changes")
						}
					} else {
						fmt.Println("Initial fetch")
					}
				}

				fmt.Println(result)

				prevChallenges = challenges
				return nil
			}

			// Initial fetch
			if err := fetchAndPrint(); err != nil {
				return err
			}

			// If --once, exit
			if once {
				return nil
			}

			// Continuous watching
			for {
				select {
				case <-ticker.C:
					if err := fetchAndPrint(); err != nil {
						fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					}

				case <-sigChan:
					fmt.Println("\nStopping watch...")
					return nil
				}
			}
		},
	}

	cmd.Flags().DurationVar(&interval, "interval", 5*time.Second, "Refresh interval")
	cmd.Flags().StringVar(&challengeID, "challenge", "", "Watch specific challenge only")
	cmd.Flags().BoolVar(&once, "once", false, "Print once and exit")

	return cmd
}

// detectChangeCount counts the number of goals that have changed
func detectChangeCount(prev, curr []api.Challenge) int {
	changes := 0

	// Create map of prev challenges for quick lookup
	prevMap := make(map[string]api.Challenge)
	for _, c := range prev {
		prevMap[c.ID] = c
	}

	// Check for changes
	for _, currChallenge := range curr {
		prevChallenge, exists := prevMap[currChallenge.ID]
		if !exists {
			continue
		}

		// Create goal maps
		prevGoals := make(map[string]api.Goal)
		for _, g := range prevChallenge.Goals {
			prevGoals[g.ID] = g
		}

		// Compare goals
		for _, currGoal := range currChallenge.Goals {
			prevGoal, exists := prevGoals[currGoal.ID]
			if !exists {
				continue
			}

			if currGoal.Progress != prevGoal.Progress || currGoal.Status != prevGoal.Status {
				changes++
			}
		}
	}

	return changes
}
