package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

func (a *App) listCmd() *cobra.Command {
	var filter string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available documentation sets",
		Long: `List all documentation sets available on DevDocs.io.

Use --filter to narrow results by name, slug, or type (case-insensitive).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			a.progressf("fetching documentation index...")
			docs, err := a.client.ListDocs(cmd.Context())
			if err != nil {
				return mapFetchErr(err)
			}
			if filter != "" {
				low := strings.ToLower(filter)
				filtered := docs[:0]
				for _, d := range docs {
					if strings.Contains(strings.ToLower(d.Name), low) ||
						strings.Contains(strings.ToLower(d.Slug), low) ||
						strings.Contains(strings.ToLower(d.Type), low) {
						filtered = append(filtered, d)
					}
				}
				docs = filtered
			}
			n := a.effectiveLimit(0)
			if n > 0 && n < len(docs) {
				docs = docs[:n]
			}
			return a.renderOrEmpty(docs, len(docs))
		},
	}

	cmd.Flags().StringVarP(&filter, "filter", "f", "", "case-insensitive substring filter on name/slug/type")
	return cmd
}
