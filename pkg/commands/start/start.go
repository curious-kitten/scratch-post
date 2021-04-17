package start

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"

	"github.com/curious-kitten/scratch-post/internal/db"
	"github.com/curious-kitten/scratch-post/internal/decoder"
	"github.com/curious-kitten/scratch-post/internal/health"
	"github.com/curious-kitten/scratch-post/internal/info"
	"github.com/curious-kitten/scratch-post/internal/keys"
	"github.com/curious-kitten/scratch-post/internal/logger"
	"github.com/curious-kitten/scratch-post/internal/store"
	"github.com/curious-kitten/scratch-post/pkg/administration/users"
	"github.com/curious-kitten/scratch-post/pkg/executions"
	"github.com/curious-kitten/scratch-post/pkg/http/auth"
	"github.com/curious-kitten/scratch-post/pkg/http/endpoints"
	"github.com/curious-kitten/scratch-post/pkg/http/methods"
	"github.com/curious-kitten/scratch-post/pkg/http/middleware"
	"github.com/curious-kitten/scratch-post/pkg/http/probes"
	"github.com/curious-kitten/scratch-post/pkg/http/router"
	"github.com/curious-kitten/scratch-post/pkg/metadata"
	"github.com/curious-kitten/scratch-post/pkg/projects"
	"github.com/curious-kitten/scratch-post/pkg/scenarios"
	"github.com/curious-kitten/scratch-post/pkg/testplans"
)

var testDBConfigFile string
var adminDBConfigFile string
var apiConfigFile string
var securityFile string
var isJWT bool

func init() {
	Command.Flags().StringVar(&testDBConfigFile, "testdb", "testdb.json", "Path to DB config settings")
	Command.Flags().StringVar(&adminDBConfigFile, "admindb", "admindb.json", "Path to admin DB config settings")
	Command.Flags().StringVar(&apiConfigFile, "apiconfig", "apiconfig.json", "Path to API config settings")
	Command.Flags().StringVar(&securityFile, "scenarios", "scenarios", "collection name to be used for scenarios")
	Command.Flags().BoolVar(&isJWT, "isJWT", false, "Sets the authentication type to JWT. Default is session ID")
	Command.Flags().StringVar(&securityFile, "securityFile", "security.txt", "Path to file which contains the JWT security string")
}

var Command = &cobra.Command{
	Use:   "start",
	Short: "Starts the server for managing test cases",
	Run: func(cmd *cobra.Command, args []string) {

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		app := info.AppInfo()
		instance := info.InstanceInfo()

		log, flush, err := logger.New(app, instance, true)
		exitOnError(log, err)
		defer func() {
			_ = flush()
		}()

		log.Info("Reading configurations...")
		// Reading DB config file
		testDBConfContents, err := os.Open(testDBConfigFile)
		exitOnError(log, err)

		storeCfg := &store.Config{}
		err = decoder.Decode(storeCfg, testDBConfContents)
		exitOnError(log, err)

		adminDBConfContents, err := os.Open(adminDBConfigFile)
		exitOnError(log, err)
		adminDBCfg := &db.Config{}
		err = decoder.Decode(adminDBCfg, adminDBConfContents)
		exitOnError(log, err)
		// Reading API config file
		apiConfContents, err := os.Open(apiConfigFile)
		exitOnError(log, err)
		apiCfg := &endpoints.Config{}
		err = decoder.Decode(apiCfg, apiConfContents)
		exitOnError(log, err)

		log.Info("Starting app...")
		r := router.New(log)
		versionedRouter := r.PathPrefix(apiCfg.RootPrefix).Subrouter()

		conditions := health.NewConditions(app, instance)
		probes.RegisterHTTPProbes(versionedRouter.PathPrefix(apiCfg.Endpoints.Probes).Subrouter(), conditions)

		meta := metadata.NewMetaManager()

		sql, err := db.New(*adminDBCfg)
		exitOnError(log, err)

		err = sql.Ping()
		exitOnError(log, err)

		var authorizer auth.Authorizer
		if isJWT {
			securityKey, err := os.Open(securityFile)
			exitOnError(log, err)
			keyRetriever := &keys.Retriever{Item: securityKey}
			authorizer = auth.NewJWTHandler(keyRetriever)
		} else {
			authorizer = auth.NewSessionHandler(sql, log)
		}
		authorizer.Cleanup(24 * time.Hour)

		client, err := store.Client(ctx, storeCfg.Address)
		exitOnError(log, err)

		userDB, err := users.NewUserDB(sql)
		exitOnError(log, err)

		authEndpoints := auth.NewEndpoints(ctx, users.IsPasswordCorrect(userDB), authorizer)
		authEndpoints.Register(versionedRouter)

		// Admin endpoints
		administrationRouter := versionedRouter.PathPrefix(apiCfg.Endpoints.Admin.Prefix).Subrouter()

		// User endpoints
		usersRouter := administrationRouter.PathPrefix(apiCfg.Endpoints.Admin.Users).Subrouter()
		usersRouter.Use(middleware.Authorization(authorizer))
		methods.Post(ctx, users.Create(userDB), usersRouter, log)
		methods.Get(ctx, users.Get(userDB), usersRouter, log)

		//  Projects endpoint
		projectsCollection, err := store.Collection(storeCfg.DataBase, storeCfg.Collections.Projects, client, []string{"name"})
		exitOnError(log, err)
		projectRouter := versionedRouter.PathPrefix(apiCfg.Endpoints.Projects).Subrouter()
		projectRouter.Use(middleware.Authorization(authorizer))
		methods.Post(ctx, projects.New(meta, projectsCollection), projectRouter, log)
		methods.List(ctx, projects.List(projectsCollection), projectRouter, log)
		methods.Get(ctx, projects.Get(projectsCollection), projectRouter, log)
		methods.Delete(ctx, projects.Delete(projectsCollection), projectRouter, log)
		methods.Put(ctx, projects.Update(meta, projectsCollection), projectRouter, log)

		// Scenario endpoints
		scenarioCollection, err := store.Collection(storeCfg.DataBase, storeCfg.Collections.Scenarios, client, []string{"projectId", "name"})
		exitOnError(log, err)
		scenarioRouter := versionedRouter.PathPrefix(apiCfg.Endpoints.Scenarios).Subrouter()
		scenarioRouter.Use(middleware.Authorization(authorizer))
		methods.Post(ctx, scenarios.New(meta, scenarioCollection, projects.Get(projectsCollection)), scenarioRouter, log)
		methods.List(ctx, scenarios.List(scenarioCollection), scenarioRouter, log)
		methods.Get(ctx, scenarios.Get(scenarioCollection), scenarioRouter, log)
		methods.Delete(ctx, scenarios.Delete(scenarioCollection), scenarioRouter, log)
		methods.Put(ctx, scenarios.Update(meta, scenarioCollection, projects.Get(projectsCollection)), scenarioRouter, log)

		// TestPlan endpoints
		testPlanCollection, err := store.Collection(storeCfg.DataBase, storeCfg.Collections.TestPlans, client, []string{"projectId", "name"})
		exitOnError(log, err)
		testPlanRouter := versionedRouter.PathPrefix(apiCfg.Endpoints.TestPlans).Subrouter()
		testPlanRouter.Use(middleware.Authorization(authorizer))
		methods.Post(ctx, testplans.New(meta, testPlanCollection, projects.Get(projectsCollection)), testPlanRouter, log)
		methods.List(ctx, testplans.List(testPlanCollection), testPlanRouter, log)
		methods.Get(ctx, testplans.Get(testPlanCollection), testPlanRouter, log)
		methods.Delete(ctx, testplans.Delete(testPlanCollection), testPlanRouter, log)
		methods.Put(ctx, testplans.Update(meta, testPlanCollection, projects.Get(projectsCollection)), testPlanRouter, log)

		// Executions endpoints
		executionCollection, err := store.Collection(storeCfg.DataBase, storeCfg.Collections.Executions, client, []string{})
		exitOnError(log, err)
		executionRouter := versionedRouter.PathPrefix(apiCfg.Endpoints.Executions).Subrouter()
		executionRouter.Use(middleware.Authorization(authorizer))
		methods.Post(ctx, executions.New(meta, executionCollection, projects.Get(projectsCollection), scenarios.Get(scenarioCollection), testplans.Get(testPlanCollection)), executionRouter, log)
		methods.List(ctx, executions.List(executionCollection), executionRouter, log)
		methods.Get(ctx, executions.Get(executionCollection), executionRouter, log)
		methods.Put(ctx, executions.Update(meta, executionCollection, projects.Get(projectsCollection), scenarios.Get(scenarioCollection), testplans.Get(testPlanCollection)), executionRouter, log)

		// Start HTTP Server
		srv := &http.Server{
			Addr:    fmt.Sprintf(":%s", apiCfg.Port),
			Handler: r,
		}

		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}()
		log.Infof("Server started on port %s", apiCfg.Port)

		<-c

		log.Info("Shutting down...")
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		err = srv.Shutdown(ctx)
		exitOnError(log, err)
	},
}

func exitOnError(log logger.Logger, err error) {
	if err != nil {
		log.Errorw("fatal error during startup", "error", err)
		panic(err)
	}
}
