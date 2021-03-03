package generate

import (
	"github.com/spf13/cobra"

	"github.com/curious-kitten/scratch-post/pkg/commands/generate/adminconfig"
	"github.com/curious-kitten/scratch-post/pkg/commands/generate/apiconfig"
	"github.com/curious-kitten/scratch-post/pkg/commands/generate/storeconfig"
)

func init() {
	Command.AddCommand(
		storeconfig.Command,
		adminconfig.Command,
		apiconfig.Command,
	)
}

// Command is used to colocate all the generate commands
var Command = &cobra.Command{
	Use:   "generate",
	Short: "generate is used to generate the configurations needed to run scratch-post",
}
