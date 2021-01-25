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

	apiConfig "github.com/curious-kitten/scratch-post/internal/config/api"
	storeConfig "github.com/curious-kitten/scratch-post/internal/config/store"
	"github.com/curious-kitten/scratch-post/internal/decoder"
	"github.com/curious-kitten/scratch-post/internal/health"
	"github.com/curious-kitten/scratch-post/internal/http/probes"
	"github.com/curious-kitten/scratch-post/internal/http/router"
	"github.com/curious-kitten/scratch-post/internal/info"
	"github.com/curious-kitten/scratch-post/internal/logger"
	"github.com/curious-kitten/scratch-post/internal/store"
	"github.com/curious-kitten/scratch-post/pkg/endpoints"
	"github.com/curious-kitten/scratch-post/pkg/executions"
	"github.com/curious-kitten/scratch-post/pkg/metadata"
	"github.com/curious-kitten/scratch-post/pkg/projects"
	"github.com/curious-kitten/scratch-post/pkg/scenarios"
	"github.com/curious-kitten/scratch-post/pkg/testplans"
)

var (
	dbconfigFile  *string
	apiconfigFile *string
)

func init() {
	dbconfigFile = flag.String("dbconfig", "/etc/db.json", "Path to DB config settings")
	apiconfigFile = flag.String("apiconfig", "/etc/api.json", "Path to API config settings")
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
	dbConfContents, err := os.Open(*dbconfigFile)
	if err != nil {
		panic(err)
	}
	storeCfg := &storeConfig.Config{}
	if err := decoder.Decode(storeCfg, dbConfContents); err != nil {
		panic(err)
	}
	// Reading API config file
	apiConfContents, err := os.Open(*apiconfigFile)
	if err != nil {
		panic(err)
	}
	apiCfg := &apiConfig.Config{}
	if err := decoder.Decode(apiCfg, apiConfContents); err != nil {
		panic(err)
	}

	log.Info("Starting app...")
	r := router.New(log)
	versionedRouter := r.PathPrefix(apiCfg.RootPrefix).Subrouter()

	conditions := health.NewConditions(app, instance)
	probes.RegisterHTTPProbes(versionedRouter.PathPrefix(apiCfg.Endpoints.Probes).Subrouter(), conditions)

	meta := metadata.NewMetaManager()

	client, err := store.Client(ctx, storeCfg.Address)
	if err != nil {
		panic(err)
	}
	//  Projects endpoint
	projectsCollection, err := store.Collection(storeCfg.DataBase, storeCfg.Collections.Projects, client)
	if err != nil {
		panic(err)
	}
	projectRouter := versionedRouter.PathPrefix(apiCfg.Endpoints.Projects).Subrouter()
	endpoints.Creator(ctx, projects.New(meta, projectsCollection), projectRouter)
	endpoints.Lister(ctx, projects.List(projectsCollection), projectRouter)
	endpoints.Getter(ctx, projects.Get(projectsCollection), projectRouter)
	endpoints.Deleter(ctx, projects.Delete(projectsCollection), projectRouter)
	endpoints.Updater(ctx, projects.Update(projectsCollection), projectRouter)

	// Scenario endpoints
	scenarioCollection, err := store.Collection(storeCfg.DataBase, storeCfg.Collections.Scenarios, client)
	if err != nil {
		panic(err)
	}
	scenarioRouter := versionedRouter.PathPrefix(apiCfg.Endpoints.Scenarios).Subrouter()
	endpoints.Creator(ctx, scenarios.New(meta, scenarioCollection, projects.Get(projectsCollection)), scenarioRouter)
	endpoints.Lister(ctx, scenarios.List(scenarioCollection), scenarioRouter)
	endpoints.Getter(ctx, scenarios.Get(scenarioCollection), scenarioRouter)
	endpoints.Deleter(ctx, scenarios.Delete(scenarioCollection), scenarioRouter)
	endpoints.Updater(ctx, scenarios.Update(scenarioCollection, projects.Get(projectsCollection)), scenarioRouter)

	// TestPlan endpoints
	testPlanCollection, err := store.Collection(storeCfg.DataBase, storeCfg.Collections.TestPlans, client)
	if err != nil {
		panic(err)
	}
	testPlanRouter := versionedRouter.PathPrefix(apiCfg.Endpoints.TestPlans).Subrouter()
	endpoints.Creator(ctx, testplans.New(meta, testPlanCollection, projects.Get(projectsCollection)), testPlanRouter)
	endpoints.Lister(ctx, testplans.List(testPlanCollection), testPlanRouter)
	endpoints.Getter(ctx, testplans.Get(testPlanCollection), testPlanRouter)
	endpoints.Deleter(ctx, testplans.Delete(testPlanCollection), testPlanRouter)
	endpoints.Updater(ctx, testplans.Update(testPlanCollection, projects.Get(projectsCollection)), testPlanRouter)

	// Executions endpoints
	executionCollection, err := store.Collection(storeCfg.DataBase, storeCfg.Collections.Executions, client)
	if err != nil {
		panic(err)
	}
	executionRouter := versionedRouter.PathPrefix(apiCfg.Endpoints.Executions).Subrouter()
	endpoints.Creator(ctx, executions.New(meta, executionCollection, projects.Get(projectsCollection), scenarios.Get(scenarioCollection), testplans.Get(testPlanCollection)), executionRouter)
	endpoints.Lister(ctx, executions.List(executionCollection), executionRouter)
	endpoints.Getter(ctx, executions.Get(executionCollection), executionRouter)
	endpoints.Updater(ctx, executions.Update(executionCollection, projects.Get(projectsCollection), scenarios.Get(scenarioCollection), testplans.Get(testPlanCollection)), executionRouter)

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
