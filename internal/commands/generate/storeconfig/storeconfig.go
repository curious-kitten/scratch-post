package storeconfig

import (
	"encoding/json"
	"io/ioutil"

	"github.com/spf13/cobra"

	"github.com/curious-kitten/scratch-post/internal/store"
)

var address string
var database string
var projects string
var scenarios string
var testplans string
var executions string
var file string

func init() {
	Command.Flags().StringVar(&address, "address", "", "testdb server address")
	Command.Flags().StringVar(&database, "database", "", "mongo database name")
	Command.Flags().StringVar(&projects, "projects", "projects", "collection name to be used for projects")
	Command.Flags().StringVar(&scenarios, "scenarios", "scenarios", "collection name to be used for scenarios")
	Command.Flags().StringVar(&testplans, "testplans", "testplans", "collection name to be used for testplans")
	Command.Flags().StringVar(&executions, "executions", "executions", "collection name to be used for executions")
	Command.Flags().StringVar(&file, "file", "testdb.json", "file which will contain the configuration")
	// address and databse are mandatory fields
	_ = cobra.MarkFlagRequired(Command.Flags(), "address")
	_ = cobra.MarkFlagRequired(Command.Flags(), "database")
}

var Command = &cobra.Command{
	Use:   "test-db-config",
	Short: "test-db-config generates JSON file for configuring the database to store test information",
	Long: `test-db-config generates JSON file for configuring the Mongo database.
	This data base is used to store test information. 
	Information provided by this config file is:
	- address:     the URL to connect to the instance 
	- database:    the specific database to be used in the instance
	- collections: a map which you can use to specify what collection each scratch-post item type can use`,
	RunE: func(cmd *cobra.Command, args []string) error {
		storeConfig := store.Config{
			Address:  address,
			DataBase: database,
			Collections: store.Collections{
				Projects:   projects,
				Scenarios:  scenarios,
				TestPlans:  testplans,
				Executions: executions,
			},
		}
		cfg, err := json.MarshalIndent(storeConfig, "", "  ")
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(file, cfg, 0600); err != nil {
			return err
		}
		return nil
	},
}
