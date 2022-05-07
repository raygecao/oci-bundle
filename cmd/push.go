/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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

	"github.com/spf13/cobra"

	"ocibundle/internal/action"
)

var pushOpts action.PushOpts

// pushCmd pushes an artifact to OCI registry
var pushCmd = &cobra.Command{
	Use:           "push",
	Short:         "push an artifact to oci registry",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return action.Push(context.TODO(), args[0], &pushOpts)
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
	flag := pushCmd.Flags()
	flag.StringSliceVar(&pushOpts.Paths, "files", nil, "file paths for uploading")
	flag.StringVarP(&pushOpts.Type, "type", "t", "unknown", "the type for the artifacts")
	flag.StringVarP(&pushOpts.Author, "author", "a", "", "the author for the artifacts")
	flag.StringSliceVar(&pushOpts.Annotations, "annotations", nil, "the annotations for the artifacts, such as `a=b,c=d,e=f'")
}
