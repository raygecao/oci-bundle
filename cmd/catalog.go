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

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
)

// catalogCmd lists all repositories of the oci registry
var catalogCmd = &cobra.Command{
	Use:           "catalog [registryHost]",
	Short:         "list all repositories of the registry",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rps, err := crane.Catalog(args[0], crane.Insecure)
		if err != nil {
			return err
		}
		table := uitable.New()
		table.AddRow("NO.", "NAME")
		for i, rp := range rps {
			table.AddRow(i, rp)
		}
		fmt.Println(table)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(catalogCmd)
}
