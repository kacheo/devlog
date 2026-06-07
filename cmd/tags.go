package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kacheo/devlog/internal/config"
	"github.com/kacheo/devlog/internal/render"
	"github.com/kacheo/devlog/internal/store"
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Manage journal tags",
	Long: `Manage tags across all journal entries.

By default, lists all tags with usage counts.

Examples:
  devlog tags                          # list all tags
  devlog tags list                     # same as above
  devlog tags list --json              # machine-readable
  devlog tags rename auth oauth        # rename a tag across all entries`,
	Args: cobra.NoArgs,
	RunE: runTagsList,
}

var tagsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tags with usage counts",
	Args:  cobra.NoArgs,
	RunE:  runTagsList,
}

var tagsRenameCmd = &cobra.Command{
	Use:   "rename <old> <new>",
	Short: "Rename a tag across all journal entries",
	Args:  cobra.ExactArgs(2),
	RunE:  runTagsRename,
}

func init() {
	tagsCmd.AddCommand(tagsListCmd)
	tagsCmd.AddCommand(tagsRenameCmd)
	rootCmd.AddCommand(tagsCmd)
}

func openStore() (*store.Store, error) {
	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	st, err := store.New(cfg.Journal.Dir)
	if err != nil {
		return nil, fmt.Errorf("journal not configured: %w\nRun 'devlog init' to set up", err)
	}
	return st, nil
}

func runTagsList(cmd *cobra.Command, _ []string) error {
	st, err := openStore()
	if err != nil {
		return err
	}

	tags, err := st.ListTags()
	if err != nil {
		return fmt.Errorf("listing tags: %w", err)
	}

	w := cmd.OutOrStdout()

	if globalJSON {
		b, err := render.TagsJSON(tags)
		if err != nil {
			return err
		}
		fmt.Fprintln(w, string(b))
		return nil
	}

	render.TagsTerminal(tags, w)
	return nil
}

func runTagsRename(cmd *cobra.Command, args []string) error {
	oldTag, newTag := args[0], args[1]

	st, err := openStore()
	if err != nil {
		return err
	}

	n, err := st.RenameTag(oldTag, newTag)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	if n == 0 {
		fmt.Fprintf(w, "No entries use tag %q.\n", oldTag)
	} else {
		fmt.Fprintf(w, "Renamed tag %q → %q in %d file(s).\n", oldTag, newTag, n)
	}
	return nil
}
