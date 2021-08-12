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
	"github.com/spf13/cobra"

	"github.com/curious-kitten/scratch-post/internal/commands/generate"
	"github.com/curious-kitten/scratch-post/internal/commands/start"
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

func main() {
	if err := Root.Execute(); err != nil {
		panic(err)
	}
}
