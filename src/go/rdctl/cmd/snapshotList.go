package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/rancher-sandbox/rancher-desktop/src/go/rdctl/pkg/snapshot"
	"github.com/spf13/cobra"
)

// SortableSnapshots are []snapshot.Snapshot sortable by date created.
type SortableSnapshots []snapshot.Snapshot

func (snapshots SortableSnapshots) Len() int {
	return len(snapshots)
}

func (snapshots SortableSnapshots) Less(i, j int) bool {
	return snapshots[i].Created.Sub(snapshots[j].Created) < 0
}

func (snapshots SortableSnapshots) Swap(i, j int) {
	temp := snapshots[i]
	snapshots[i] = snapshots[j]
	snapshots[j] = temp
}

var snapshotListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List snapshots",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		return exitWithJsonOrErrorCondition(listSnapshot())
	},
}

func init() {
	snapshotCmd.AddCommand(snapshotListCmd)
	snapshotListCmd.Flags().BoolVar(&outputJsonFormat, "json", false, "output json format")
}

func listSnapshot() error {
	manager, err := snapshot.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create snapshot manager: %w", err)
	}
	snapshots, err := manager.List(false)
	if err != nil {
		return fmt.Errorf("failed to list snapshots: %w", err)
	}
	sort.Sort(SortableSnapshots(snapshots))
	if outputJsonFormat {
		return jsonOutput(snapshots)
	}
	return tabularOutput(snapshots)
}

func jsonOutput(snapshots []snapshot.Snapshot) error {
	for _, aSnapshot := range snapshots {
		aSnapshot.ID = ""
		jsonBuffer, err := json.Marshal(aSnapshot)
		if err != nil {
			return err
		}
		fmt.Println(string(jsonBuffer))
	}
	return nil
}

func tabularOutput(snapshots []snapshot.Snapshot) error {
	if len(snapshots) == 0 {
		fmt.Fprintln(os.Stderr, "No snapshots present.")
		return nil
	}
	writer := tabwriter.NewWriter(os.Stdout, 0, 4, 4, ' ', 0)
	fmt.Fprintf(writer, "NAME\tCREATED\tDESCRIPTION\n")
	for _, aSnapshot := range snapshots {
		prettyCreated := aSnapshot.Created.Format(time.RFC1123)
		desc := aSnapshot.Description
		idx := strings.Index(desc, "\n")
		if idx >= 0 {
			// If the description starts with a newline, it will appear empty in this view.
			// Use the json view to get the full description
			desc = desc[0:idx]
		}
		if len(desc) > 63 {
			desc = desc[0:60] + "..."
		} else if idx >= 0 {
			// The string was truncated because of a newline, so add an ellipsis to show that
			// Do this even if the newline was the last character - we've still truncated *something*.
			desc += "..."
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\n", aSnapshot.Name, prettyCreated, desc)
	}
	writer.Flush()
	return nil
}
