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
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"

	"github.com/google/go-containerregistry/pkg/crane"
)

// manifestCmd get manifest from oci registry
var manifestCmd = &cobra.Command{
	Use:           "manifest [ref]",
	Short:         "Get the manifest of an artifact with ref",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := crane.Manifest(args[0], crane.Insecure)
		if err != nil {
			return err
		}
		fmt.Println(string(pretty.Color(pretty.Pretty(info), nil)))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(manifestCmd)
}
