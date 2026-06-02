package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/kacheo/devlog/internal/config"
	"github.com/kacheo/devlog/internal/git"
	"github.com/kacheo/devlog/internal/render"
	"github.com/kacheo/devlog/internal/store"
)

var syncQuiet bool

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Import today's commits and PRs from configured repos",
	Args:  cobra.NoArgs,
	RunE:  runSync,
}

func init() {
	syncCmd.Flags().BoolVar(&syncQuiet, "quiet", false, "suppress all output")
	rootCmd.AddCommand(syncCmd)
}

func runSyncWithConfig(cfg *config.Config, date time.Time, quiet bool) (*store.DayEntry, error) {
	st, err := store.New(cfg.Journal.Dir)
	if err != nil {
		return nil, fmt.Errorf("journal not configured: %w\nRun 'devlog init' to set up", err)
	}
	entry, err := st.LoadOrCreate(date)
	if err != nil {
		return nil, fmt.Errorf("loading day file: %w", err)
	}

	for _, ws := range cfg.Workspaces {
		if _, statErr := os.Stat(ws.Path); os.IsNotExist(statErr) {
			if !quiet {
				fmt.Fprintf(os.Stderr, "warning: workspace path %q not found, skipping\n", ws.Name)
			}
		}
	}

	repos, err := cfg.EffectiveRepos()
	if err != nil {
		return nil, fmt.Errorf("resolving repos: %w", err)
	}

	for _, repo := range repos {
		if _, err := os.Stat(repo.Path); os.IsNotExist(err) {
			if !quiet {
				fmt.Fprintf(os.Stderr, "warning: repo path %q not found, skipping\n", repo.Path)
			}
			continue
		}

		commits, err := git.ScanCommits(repo.Path, repo.Name, date)
		if err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "warning: scanning commits in %q: %v\n", repo.Name, err)
			}
		} else {
			entry.Commits = mergeCommits(entry.Commits, commits)
		}

		ownerRepo := repo.GitHubSlug
		if ownerRepo == "" {
			ownerRepo, _ = git.GetOriginSlug(repo.Path)
		}
		if ownerRepo != "" {
			prs, err := git.FetchPRs(ownerRepo, repo.Name, cfg.GitHub.Token, date)
			if err != nil {
				if !quiet {
					fmt.Fprintf(os.Stderr, "warning: fetching PRs for %q: %v\n", repo.Name, err)
				}
			} else {
				entry.PRs = mergePRs(entry.PRs, prs)
			}
		}
	}

	if err := st.Save(entry); err != nil {
		return nil, err
	}
	return entry, nil
}

func runSync(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	date, err := resolveDate(globalDate)
	if err != nil {
		return err
	}

	entry, err := runSyncWithConfig(cfg, date, syncQuiet)
	if err != nil {
		return err
	}

	if !syncQuiet {
		if globalJSON {
			out, err := render.ShowJSON(entry)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(out)) //nolint:errcheck
		} else {
			render.ShowTerminal(entry, cmd.OutOrStdout())
		}
	}
	return nil
}

// mergeCommits merges incoming commits into existing, deduplicating by SHA.
func mergeCommits(existing, incoming []store.Commit) []store.Commit {
	seen := make(map[string]bool, len(existing))
	for _, c := range existing {
		seen[c.SHA] = true
	}
	result := append([]store.Commit(nil), existing...)
	for _, c := range incoming {
		if !seen[c.SHA] {
			seen[c.SHA] = true
			result = append(result, c)
		}
	}
	return result
}

// mergePRs merges incoming PRs into existing, deduplicating by Number+Repo.
// If an incoming PR matches an existing one, the incoming state is used to update.
func mergePRs(existing, incoming []store.PR) []store.PR {
	type key struct {
		number int
		repo   string
	}
	seen := make(map[key]int, len(existing))
	result := append([]store.PR(nil), existing...)
	for i, pr := range result {
		seen[key{pr.Number, pr.Repo}] = i
	}
	for _, pr := range incoming {
		k := key{pr.Number, pr.Repo}
		if idx, ok := seen[k]; ok {
			result[idx] = pr // update state
		} else {
			seen[k] = len(result)
			result = append(result, pr)
		}
	}
	return result
}
