package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) entriesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "entries <slug>",
		Short: "List entries (table of contents) for a doc set",
		Long: `List all documentation entries for a given doc set.

The slug identifies the doc set (e.g., python~3.11, go, react). Use
'devdocs list' to discover available slugs.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]
			a.progressf("fetching entries for %s...", slug)
			entries, err := a.client.DocEntries(cmd.Context(), slug)
			if err != nil {
				return mapFetchErr(err)
			}
			n := a.effectiveLimit(0)
			if n > 0 && n < len(entries) {
				entries = entries[:n]
			}
			return a.renderOrEmpty(entries, len(entries))
		},
	}
}

func (a *App) typesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "types <slug>",
		Short: "List entry types for a doc set",
		Long: `List the entry type summary for a given doc set.

Shows each entry type (e.g., function, class, method) and how many entries
of that type exist. Use 'devdocs list' to discover available slugs.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]
			a.progressf("fetching types for %s...", slug)
			types, err := a.client.DocTypes(cmd.Context(), slug)
			if err != nil {
				return mapFetchErr(err)
			}
			n := a.effectiveLimit(0)
			if n > 0 && n < len(types) {
				types = types[:n]
			}
			return a.renderOrEmpty(types, len(types))
		},
	}
}
