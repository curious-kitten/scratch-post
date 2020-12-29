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

	"github.com/gorilla/mux"

	"github.com/curious-kitten/scratch-post/internal/health"
	"github.com/curious-kitten/scratch-post/internal/http/middleware"
	"github.com/curious-kitten/scratch-post/internal/http/probes"
	"github.com/curious-kitten/scratch-post/internal/info"
	"github.com/curious-kitten/scratch-post/internal/logger"
)

var (
	port *string
)

func init() {
	port = flag.String("port", "9090", "Port of server")
	flag.Parse()
}

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
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

	log.Info("Starting app...")
	r := mux.NewRouter()
	r.Use(middleware.Logging(log))
	versionedRouter := r.PathPrefix("/api/v1").Subrouter()

	conditions := health.NewConditions(app, instance)

	probes.RegisterHTTPProbes(versionedRouter.PathPrefix("/probes").Subrouter(), conditions)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", *port),
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-c

	log.Info("Shutting down...")
	ctx, cancel = context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	err = srv.Shutdown(ctx)
	if err != nil {
		log.Error(err)
	}
}
