package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"text/tabwriter"
)

func formatSize(bytes int64) string {
	switch {
	case bytes >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(bytes)/1024/1024)
	case bytes >= 1024:
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func PrintSummaryTable(results []buildResult) {
	slices.SortFunc(results, func(a, b buildResult) int {
		if a.platform.GOOS != b.platform.GOOS {
			if a.platform.GOOS < b.platform.GOOS {
				return -1
			}
			return 1
		}
		if a.platform.GOARCH < b.platform.GOARCH {
			return -1
		}
		if a.platform.GOARCH > b.platform.GOARCH {
			return 1
		}
		return 0
	})

	fmt.Println()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "TARGET\tSTATUS\tOUTPUT\tSIZE")
	fmt.Fprintln(w, "------\t------\t------\t----")

	for _, r := range results {
		target := fmt.Sprintf("%s/%s", r.platform.GOOS, r.platform.GOARCH)
		if r.success {
			fmt.Fprintf(w, "%s\t✓\t%s\t%s\n",
				target,
				filepath.Base(r.output),
				formatSize(r.fileSize),
			)
		} else {
			fmt.Fprintf(w, "%s\t✗\t-\t%s\n",
				target,
				r.err.Error(),
			)
		}
	}
	w.Flush()
	fmt.Println()
}
