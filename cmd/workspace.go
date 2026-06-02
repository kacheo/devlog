package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kacheo/devlog/internal/config"
)

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage workspace directories for auto-tracking",
}

var workspaceAddCmd = &cobra.Command{
	Use:   "add <path>",
	Short: "Register a workspace directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		abs, err := filepath.Abs(args[0])
		if err != nil {
			return err
		}
		added, err := addWorkspaceToConfig(config.DefaultPath(), abs)
		if err != nil {
			return err
		}
		if added {
			fmt.Fprintf(cmd.OutOrStdout(), "Workspace added: %s\n", abs)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "Workspace already registered: %s\n", abs)
		}
		return nil
	},
}

var workspaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List workspaces and their discovered repos",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(config.DefaultPath())
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		return printWorkspaceList(cfg, cmd.OutOrStdout())
	},
}

var workspaceExcludeCmd = &cobra.Command{
	Use:   "exclude <repo-path>",
	Short: "Exclude a repo path from its parent workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		abs, err := filepath.Abs(args[0])
		if err != nil {
			return err
		}
		excluded, err := excludeRepoFromWorkspace(config.DefaultPath(), abs)
		if err != nil {
			return err
		}
		if excluded {
			fmt.Fprintf(cmd.OutOrStdout(), "Excluded: %s\n", abs)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "Already excluded: %s\n", abs)
		}
		return nil
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceAddCmd, workspaceListCmd, workspaceExcludeCmd)
	rootCmd.AddCommand(workspaceCmd)
}

// addWorkspaceToConfig loads config at cfgPath, appends ws if not present, writes back.
// Returns true if a new workspace was added, false if it was already registered.
func addWorkspaceToConfig(cfgPath, wsPath string) (bool, error) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return false, err
	}
	for _, ws := range cfg.Workspaces {
		if ws.Path == wsPath {
			return false, nil
		}
	}
	cfg.Workspaces = append(cfg.Workspaces, config.WorkspaceConfig{
		Path: wsPath,
		Name: filepath.Base(wsPath),
	})
	return true, cfg.Write(cfgPath)
}

// excludeRepoFromWorkspace finds the workspace that owns repoPath (by path prefix)
// and adds repoPath to its exclude list.
// Returns true if the exclude entry was added, false if it was already present.
func excludeRepoFromWorkspace(cfgPath, repoPath string) (bool, error) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return false, err
	}
	for i, ws := range cfg.Workspaces {
		if strings.HasPrefix(repoPath, ws.Path+string(os.PathSeparator)) {
			for _, e := range ws.Exclude {
				if e == repoPath {
					return false, nil
				}
			}
			cfg.Workspaces[i].Exclude = append(cfg.Workspaces[i].Exclude, repoPath)
			return true, cfg.Write(cfgPath)
		}
	}
	return false, fmt.Errorf("no workspace found for repo %q — register the parent directory first with 'devlog workspace add'", repoPath)
}

// printWorkspaceList writes a human-readable summary of workspaces and discovered repos to w.
func printWorkspaceList(cfg *config.Config, w io.Writer) error {
	if len(cfg.Workspaces) == 0 {
		fmt.Fprintln(w, "No workspaces configured. Use 'devlog workspace add <path>' to add one.")
		return nil
	}
	for _, ws := range cfg.Workspaces {
		fmt.Fprintf(w, "workspace: %s (%s)\n", ws.Name, ws.Path)
		repos, err := config.DiscoverRepos(ws.Path, ws.Exclude)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: scanning workspace %q: %v\n", ws.Name, err)
			continue
		}
		if len(repos) == 0 {
			fmt.Fprintln(w, "  (no git repos found)")
		} else {
			for _, r := range repos {
				fmt.Fprintf(w, "  - %s (%s)\n", r.Name, r.Path)
			}
		}
		if len(ws.Exclude) > 0 {
			fmt.Fprintln(w, "  excluded:")
			for _, e := range ws.Exclude {
				fmt.Fprintf(w, "    - %s\n", e)
			}
		}
	}
	return nil
}
