package apiconfig

import (
	"encoding/json"
	"io/ioutil"

	"github.com/spf13/cobra"

	"github.com/curious-kitten/scratch-post/pkg/http/endpoints"
)

var rootPrefix string
var port string
var probes string
var projects string
var scenarios string
var testplans string
var executions string
var adminPrefix string
var users string
var file string

func init() {
	Command.Flags().StringVar(&rootPrefix, "rootPrefix", "/api/v1", "prefix for all api endpoints")
	Command.Flags().StringVar(&port, "port", "9090", "port for the server")
	Command.Flags().StringVar(&probes, "probes", "/probes", "probes endpoints")
	Command.Flags().StringVar(&projects, "projects", "/projects", "projects endpoint")
	Command.Flags().StringVar(&testplans, "testplans", "/testplans", "testplans endpoint")
	Command.Flags().StringVar(&scenarios, "scenarios", "/scenarios", "scenarios endpoint")
	Command.Flags().StringVar(&executions, "executions", "/executions", "executions endpoint")
	Command.Flags().StringVar(&adminPrefix, "adminPrefix", "/admin", "prefix for all admin endpoints")
	Command.Flags().StringVar(&users, "users", "/users", "users endpoint. Is part of the admin endpoints")

	Command.Flags().StringVar(&file, "file", "apiconfig.json", "file which will contain the configuration")
}

// Command is used to generate the config file for the Test DB
var Command = &cobra.Command{
	Use:   "api-config",
	Short: "api-config generates JSON file for configuring the REST API",
	Run: func(cmd *cobra.Command, args []string) {
		storeConfig := endpoints.Config{
			RootPrefix: rootPrefix,
			Port:       port,
			Endpoints: endpoints.Endpoints{
				Probes:     probes,
				Projects:   projects,
				Scenarios:  scenarios,
				TestPlans:  testplans,
				Executions: executions,
				Admin: endpoints.Admin{
					Prefix: adminPrefix,
					Users:  users,
				},
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
