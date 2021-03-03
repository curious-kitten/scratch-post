package adminconfig

import (
	"encoding/json"
	"io/ioutil"

	"github.com/spf13/cobra"

	"github.com/curious-kitten/scratch-post/internal/db"
)

var address string
var maxLifetime int
var maxIdle int
var maxOpen int
var file string

func init() {
	Command.Flags().StringVar(&address, "address", "", "testdb server address")
	Command.Flags().IntVar(&maxLifetime, "maxLifetime", 5, "maximum lifetime")
	Command.Flags().IntVar(&maxIdle, "maxIdle", 5, "maximum idle connections")
	Command.Flags().IntVar(&maxOpen, "maxOpen", 5, "maximum open connections")

	Command.Flags().StringVar(&file, "file", "admindb.json", "file which will contain the configuration")
	// address and databse are mandatory fields
	_ = cobra.MarkFlagRequired(Command.Flags(), "address")
}

var Command = &cobra.Command{
	Use:   "admin-db-config",
	Short: "admin-db-config generates JSON file for configuring the database to store administrative information",
	Long: `admin-db-config generates JSON file for configuring the Postgress database.
	This data base is used to store test information. 
	Information provided by this config file is:
	- address: the URL to connect to the instance`,
	Run: func(cmd *cobra.Command, args []string) {
		storeConfig := db.Config{
			Address: address,
			Connections: db.Connections{
				MaxLifetime: maxLifetime,
				MaxIdle:     maxIdle,
				MaxOpen:     maxOpen,
			},
		}
		cfg, err := json.MarshalIndent(storeConfig, "", "  ")
		if err != nil {
			panic(err)
		}
		if err := ioutil.WriteFile(file, cfg, 0644); err != nil {
			panic(err)
		}
	},
}
