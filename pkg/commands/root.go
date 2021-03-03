package commands

import (
	"github.com/spf13/cobra"

	"github.com/curious-kitten/scratch-post/pkg/commands/generate"
	"github.com/curious-kitten/scratch-post/pkg/commands/start"
)

func init() {
	Root.AddCommand(
		generate.Command,
		start.Command,
	)
}

var Root = &cobra.Command{
	Use:   "scratch-post",
	Short: "scratch-post is a test management platform",
}
