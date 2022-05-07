/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
	"fmt"

	"github.com/spf13/cobra"

	"ocibundle/internal/version"
)

type versionOptions struct {
	short bool
}

var vo versionOptions

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:           "version [list|update]",
	Short:         "The version for chef",
	Long:          `Version contains version tag，git commit and go version.`,
	SilenceErrors: true,
	SilenceUsage:  true,
	Run: func(cmd *cobra.Command, args []string) {
		vo.run()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	flag := versionCmd.Flags()
	flag.BoolVar(&vo.short, "short", false, "only print the version number")
}

func (o *versionOptions) run() {
	v := version.Get()
	if o.short {
		if len(v.GitCommit) >= 7 {
			fmt.Printf("%s+g%s\n", v.Version, v.GitCommit[:7])
		} else {
			fmt.Println(version.GetVersion())
		}
		return
	}
	fmt.Printf("%#v\n", v)
}
