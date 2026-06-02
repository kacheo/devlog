package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/kacheo/devlog/internal/config"
	"github.com/kacheo/devlog/internal/store"
)

var editCmd = &cobra.Command{
	Use:   "edit [DATE]",
	Short: "Open a day file in $EDITOR",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runEdit,
}

func init() {
	rootCmd.AddCommand(editCmd)
}

func runEdit(cmd *cobra.Command, args []string) error {
	// Mutual exclusion: positional arg and --date flag
	if len(args) > 0 && globalDate != "" {
		return fmt.Errorf("--date and positional date argument are mutually exclusive; use one or the other")
	}

	// Determine date
	dateStr := globalDate
	if len(args) > 0 {
		dateStr = args[0]
	}
	date, err := resolveDate(dateStr)
	if err != nil {
		return err
	}

	// Load config
	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Open store
	st, err := store.New(cfg.Journal.Dir)
	if err != nil {
		return fmt.Errorf("journal not configured: %w\nRun 'devlog init' to set up", err)
	}

	// Ensure the file exists
	entry, err := st.LoadOrCreate(date)
	if err != nil {
		return fmt.Errorf("loading day file: %w", err)
	}

	path := st.DayFilePath(date)

	// Create the file if it doesn't exist yet
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := st.Save(entry); err != nil {
			return fmt.Errorf("creating day file: %w", err)
		}
	}

	// Open in editor
	editor := cfg.ResolvedEditor()
	c := exec.Command(editor, path)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
