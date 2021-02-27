/*
Copyright © 2020 MATACHE MIHAI <matache91mh@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

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

var (
	testDBConfigFile  *string
	adminDBConfigFile *string
	apiConfigFile     *string
	securityFile      *string
	isJWT             *bool
)

func init() {
	testDBConfigFile = flag.String("testdb", "testdb.json", "Path to DB config settings")
	adminDBConfigFile = flag.String("admindb", "admindb.json", "Path to admin DB config settings")
	apiConfigFile = flag.String("apiconfig", "apiconfig.json", "Path to API config settings")
	isJWT = flag.Bool("isJWT", false, "Sets the authentication type to JWT. Default is session ID")
	securityFile = flag.String("securityFile", "security.txt", "Path to file which contains the JWT security string")
	flag.Parse()
}

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := info.App{
		Name:    "scratch-post",
		Version: "0.0.1-beta",
	}

	instance := info.InstanceInfo()

	log, flush, err := logger.New(app, instance, true)
	if err != nil {
		panic(err)
	}
	defer flush()

	log.Info("Reading configurations...")
	// Reading DB config file
	testDBConfContents, err := os.Open(*testDBConfigFile)
	if err != nil {
		panic(err)
	}
	storeCfg := &store.Config{}
	if err := decoder.Decode(storeCfg, testDBConfContents); err != nil {
		panic(err)
	}

	adminDBConfContents, err := os.Open(*adminDBConfigFile)
	if err != nil {
		panic(err)
	}
	adminDBCfg := &db.Config{}
	if err := decoder.Decode(adminDBCfg, adminDBConfContents); err != nil {
		panic(err)
	}
	// Reading API config file
	apiConfContents, err := os.Open(*apiConfigFile)
	if err != nil {
		panic(err)
	}
	apiCfg := &endpoints.Config{}
	if err := decoder.Decode(apiCfg, apiConfContents); err != nil {
		panic(err)
	}

	log.Info("Starting app...")
	r := router.New(log)
	versionedRouter := r.PathPrefix(apiCfg.RootPrefix).Subrouter()

	conditions := health.NewConditions(app, instance)
	probes.RegisterHTTPProbes(versionedRouter.PathPrefix(apiCfg.Endpoints.Probes).Subrouter(), conditions)

	meta := metadata.NewMetaManager()

	sql, err := db.New(*adminDBCfg)
	if err != nil {
		panic(err)
	}
	err = sql.Ping()
	if err != nil {
		panic(err)
	}

	var authorizer auth.Authorizer
	if *isJWT {
		securityKey, err := os.Open(*securityFile)
		if err != nil {
			panic(err)
		}
		keyRetriever := &keys.Retriever{Item: securityKey}
		authorizer = auth.NewJWTHandler(keyRetriever)
	} else {
		authorizer = auth.NewSessionHandler(sql, log)
	}
	authorizer.Cleanup(24 * time.Hour)

	client, err := store.Client(ctx, storeCfg.Address)
	if err != nil {
		panic(err)
	}

	userDB := users.NewUserDB(sql)

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
	if err != nil {
		panic(err)
	}
	projectRouter := versionedRouter.PathPrefix(apiCfg.Endpoints.Projects).Subrouter()
	projectRouter.Use(middleware.Authorization(authorizer))
	methods.Post(ctx, projects.New(meta, projectsCollection), projectRouter, log)
	methods.List(ctx, projects.List(projectsCollection), projectRouter, log)
	methods.Get(ctx, projects.Get(projectsCollection), projectRouter, log)
	methods.Delete(ctx, projects.Delete(projectsCollection), projectRouter, log)
	methods.Put(ctx, projects.Update(meta, projectsCollection), projectRouter, log)

	// Scenario endpoints
	scenarioCollection, err := store.Collection(storeCfg.DataBase, storeCfg.Collections.Scenarios, client, []string{"projectId", "name"})
	if err != nil {
		panic(err)
	}
	scenarioRouter := versionedRouter.PathPrefix(apiCfg.Endpoints.Scenarios).Subrouter()
	scenarioRouter.Use(middleware.Authorization(authorizer))
	methods.Post(ctx, scenarios.New(meta, scenarioCollection, projects.Get(projectsCollection)), scenarioRouter, log)
	methods.List(ctx, scenarios.List(scenarioCollection), scenarioRouter, log)
	methods.Get(ctx, scenarios.Get(scenarioCollection), scenarioRouter, log)
	methods.Delete(ctx, scenarios.Delete(scenarioCollection), scenarioRouter, log)
	methods.Put(ctx, scenarios.Update(meta, scenarioCollection, projects.Get(projectsCollection)), scenarioRouter, log)

	// TestPlan endpoints
	testPlanCollection, err := store.Collection(storeCfg.DataBase, storeCfg.Collections.TestPlans, client, []string{"projectId", "name"})
	if err != nil {
		panic(err)
	}
	testPlanRouter := versionedRouter.PathPrefix(apiCfg.Endpoints.TestPlans).Subrouter()
	testPlanRouter.Use(middleware.Authorization(authorizer))
	methods.Post(ctx, testplans.New(meta, testPlanCollection, projects.Get(projectsCollection)), testPlanRouter, log)
	methods.List(ctx, testplans.List(testPlanCollection), testPlanRouter, log)
	methods.Get(ctx, testplans.Get(testPlanCollection), testPlanRouter, log)
	methods.Delete(ctx, testplans.Delete(testPlanCollection), testPlanRouter, log)
	methods.Put(ctx, testplans.Update(meta, testPlanCollection, projects.Get(projectsCollection)), testPlanRouter, log)

	// Executions endpoints
	executionCollection, err := store.Collection(storeCfg.DataBase, storeCfg.Collections.Executions, client, []string{})
	if err != nil {
		panic(err)
	}
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
	if err != nil {
		panic(err)
	}
}
