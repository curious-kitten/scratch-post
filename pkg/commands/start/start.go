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
	"github.com/curious-kitten/scratch-post/internal/http/endpoints"
	"github.com/curious-kitten/scratch-post/internal/http/methods"
	"github.com/curious-kitten/scratch-post/internal/http/router"
	"github.com/curious-kitten/scratch-post/internal/info"
	"github.com/curious-kitten/scratch-post/internal/keys"
	"github.com/curious-kitten/scratch-post/internal/logger"
	"github.com/curious-kitten/scratch-post/internal/store"
	"github.com/curious-kitten/scratch-post/pkg/administration/users"
	"github.com/curious-kitten/scratch-post/pkg/administration/users/auth"
	"github.com/curious-kitten/scratch-post/pkg/executions"
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
	RunE: func(cmd *cobra.Command, args []string) error {

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		app := info.AppInfo()
		instance := info.InstanceInfo()

		log, flush, err := logger.New(app, instance, true)
		if err != nil {
			err = fmt.Errorf("%s : %w", "could not start logger", err)
			log.Errorw("fatal error during startup", "error", err)
			return err
		}
		defer func() {
			_ = flush()
		}()

		log.Info("Reading configurations...")
		// Reading DB config file
		testDBConfContents, err := os.Open(testDBConfigFile)
		if err != nil {
			err = fmt.Errorf("%s : %w", "could not read test DB config", err)
			log.Errorw("fatal error during startup", "error", err)
			return err
		}

		storeCfg := &store.Config{}
		err = decoder.Decode(storeCfg, testDBConfContents)
		if err != nil {
			err = fmt.Errorf("%s : %w", "could not decode test DB config", err)
			log.Errorw("fatal error during startup", "error", err)
			return err
		}

		adminDBConfContents, err := os.Open(adminDBConfigFile)
		if err != nil {
			err = fmt.Errorf("%s : %w", "could not read admin DB config", err)
			log.Errorw("fatal error during startup", "error", err)
			return err
		}
		adminDBCfg := &db.Config{}
		err = decoder.Decode(adminDBCfg, adminDBConfContents)
		if err != nil {
			err = fmt.Errorf("%s : %w", "could not decode admin DB config", err)
			log.Errorw("fatal error during startup", "error", err)
			return err
		}
		// Reading API config file
		apiConfContents, err := os.Open(apiConfigFile)
		if err != nil {
			err = fmt.Errorf("%s : %w", "could not read api config", err)
			log.Errorw("fatal error during startup", "error", err)
			return err
		}
		apiCfg := &endpoints.Config{}
		err = decoder.Decode(apiCfg, apiConfContents)
		if err != nil {
			err = fmt.Errorf("%s : %w", "could not decode endpoints", err)
			log.Errorw("fatal error during startup", "error", err)
			return err
		}

		log.Info("Starting app...")
		r := router.New(log)
		versionedRouter := r.PathPrefix(apiCfg.RootPrefix).Subrouter()

		conditions := health.NewConditions(app, instance)
		health.RegisterHTTPProbes(versionedRouter.PathPrefix(apiCfg.Endpoints.Probes).Subrouter(), conditions)

		meta := metadata.NewMetaManager()

		sql, err := db.New(*adminDBCfg)
		if err != nil {
			err = fmt.Errorf("%s : %w", "DB connection error", err)
			log.Errorw("fatal error during startup", "error", err)
			return err
		}
		defer func() {
			log.Info("closing DB connection")
			err = sql.Close()
			if err != nil {
				log.Error("error closing DB connection", "error", err)
			}
		}()

		conditions.RegisterReadynessCondition(func() health.Condition {
			ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			err := sql.PingContext(ctx)
			if err != nil {
				return health.Condition{
					Ready:   false,
					Message: err.Error(),
					Name:    "sql ping",
				}
			}
			return health.Condition{
				Ready:   true,
				Message: "Ping success",
				Name:    "sql ping",
			}

		})

		var authorizer auth.Authorizer
		if isJWT {
			securityKey, err := os.Open(securityFile)
			if err != nil {
				err = fmt.Errorf("%s : %w", "could not open security file", err)
				log.Errorw("fatal error during startup", "error", err)
				return err
			}
			keyRetriever := &keys.Retriever{Item: securityKey}
			authorizer = auth.NewJWTHandler(keyRetriever)
		} else {
			authorizer = auth.NewSessionHandler(sql, log)
		}
		authorizer.Cleanup(24 * time.Hour)

		client, err := store.Client(ctx, storeCfg.Address)
		if err != nil {
			err = fmt.Errorf("%s : %w", "DB connection error", err)
			log.Errorw("fatal error during startup", "error", err)
			return err
		}

		conditions.RegisterReadynessCondition(func() health.Condition {
			ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			err = client.Ping(ctx, nil)
			if err != nil {
				return health.Condition{
					Ready:   false,
					Message: err.Error(),
					Name:    "mongo ping",
				}
			}
			return health.Condition{
				Ready:   true,
				Message: "Ping success",
				Name:    "mongo ping",
			}
		})

		userDB, err := users.NewUserDB(sql)
		if err != nil {
			err = fmt.Errorf("%s : %w", "DB connection error", err)
			log.Errorw("fatal error during startup", "error", err)
			return err
		}

		authEndpoints := auth.NewEndpoints(ctx, users.IsPasswordCorrect(userDB), authorizer)
		authEndpoints.Register(versionedRouter)

		// Admin endpoints
		administrationRouter := versionedRouter.PathPrefix(apiCfg.Endpoints.Admin.Prefix).Subrouter()

		// User endpoints
		usersRouter := administrationRouter.PathPrefix(apiCfg.Endpoints.Admin.Users).Subrouter()
		usersRouter.Use(auth.Authorization(authorizer))
		methods.Post(ctx, users.Create(userDB), auth.GetUserIDFromRequest, usersRouter, log)
		methods.Get(ctx, users.Get(userDB), usersRouter, log)

		//  Projects endpoint
		projectsCollection, err := store.Collection(storeCfg.DataBase, storeCfg.Collections.Projects, client, []string{"name"})
		if err != nil {
			err = fmt.Errorf("%s : %w", "could not start collection", err)
			log.Errorw("fatal error during startup", "error", err)
			return err
		}
		projectRouter := versionedRouter.PathPrefix(apiCfg.Endpoints.Projects).Subrouter()
		projectRouter.Use(auth.Authorization(authorizer))
		methods.Post(ctx, projects.New(meta, projectsCollection), auth.GetUserIDFromRequest, projectRouter, log)
		methods.List(ctx, projects.List(projectsCollection), projectRouter, log)
		methods.Get(ctx, projects.Get(projectsCollection), projectRouter, log)
		methods.Delete(ctx, projects.Delete(projectsCollection), projectRouter, log)
		methods.Put(ctx, projects.Update(meta, projectsCollection), auth.GetUserIDFromRequest, projectRouter, log)

		// Scenario endpoints
		scenarioCollection, err := store.Collection(storeCfg.DataBase, storeCfg.Collections.Scenarios, client, []string{"projectId", "name"})
		if err != nil {
			err = fmt.Errorf("%s : %w", "could not start collection", err)
			log.Errorw("fatal error during startup", "error", err)
			return err
		}
		scenarioRouter := versionedRouter.PathPrefix(apiCfg.Endpoints.Scenarios).Subrouter()
		scenarioRouter.Use(auth.Authorization(authorizer))
		methods.Post(ctx, scenarios.New(meta, scenarioCollection, projects.Get(projectsCollection)), auth.GetUserIDFromRequest, scenarioRouter, log)
		methods.List(ctx, scenarios.List(scenarioCollection), scenarioRouter, log)
		methods.Get(ctx, scenarios.Get(scenarioCollection), scenarioRouter, log)
		methods.Delete(ctx, scenarios.Delete(scenarioCollection), scenarioRouter, log)
		methods.Put(ctx, scenarios.Update(meta, scenarioCollection, projects.Get(projectsCollection)), auth.GetUserIDFromRequest, scenarioRouter, log)

		// TestPlan endpoints
		testPlanCollection, err := store.Collection(storeCfg.DataBase, storeCfg.Collections.TestPlans, client, []string{"projectId", "name"})
		if err != nil {
			err = fmt.Errorf("%s : %w", "could not start collection", err)
			log.Errorw("fatal error during startup", "error", err)
			return err
		}
		testPlanRouter := versionedRouter.PathPrefix(apiCfg.Endpoints.TestPlans).Subrouter()
		testPlanRouter.Use(auth.Authorization(authorizer))
		methods.Post(ctx, testplans.New(meta, testPlanCollection, projects.Get(projectsCollection)), auth.GetUserIDFromRequest, testPlanRouter, log)
		methods.List(ctx, testplans.List(testPlanCollection), testPlanRouter, log)
		methods.Get(ctx, testplans.Get(testPlanCollection), testPlanRouter, log)
		methods.Delete(ctx, testplans.Delete(testPlanCollection), testPlanRouter, log)
		methods.Put(ctx, testplans.Update(meta, testPlanCollection, projects.Get(projectsCollection)), auth.GetUserIDFromRequest, testPlanRouter, log)

		// Executions endpoints
		executionCollection, err := store.Collection(storeCfg.DataBase, storeCfg.Collections.Executions, client, []string{})
		if err != nil {
			err = fmt.Errorf("%s : %w", "could not start collection", err)
			log.Errorw("fatal error during startup", "error", err)
			return err
		}
		executionRouter := versionedRouter.PathPrefix(apiCfg.Endpoints.Executions).Subrouter()
		executionRouter.Use(auth.Authorization(authorizer))
		methods.Post(
			ctx,
			executions.New(meta, executionCollection, projects.Get(projectsCollection), scenarios.Get(scenarioCollection), testplans.Get(testPlanCollection)),
			auth.GetUserIDFromRequest,
			executionRouter,
			log,
		)
		methods.List(ctx, executions.List(executionCollection), executionRouter, log)
		methods.Get(ctx, executions.Get(executionCollection), executionRouter, log)
		methods.Put(
			ctx,
			executions.Update(meta, executionCollection, projects.Get(projectsCollection), scenarios.Get(scenarioCollection), testplans.Get(testPlanCollection)),
			auth.GetUserIDFromRequest,
			executionRouter,
			log)

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
		defer func() {
			ctx, cancel = context.WithTimeout(ctx, time.Second*10)
			defer cancel()
			err = srv.Shutdown(ctx)
			if err != nil {
				err = fmt.Errorf("%s : %w", "shutdown issue", err)
				log.Errorw("fatal error during startup", "error", err)
			}
		}()

		<-c
		log.Info("Shutting down...")
		return nil
	},
}
