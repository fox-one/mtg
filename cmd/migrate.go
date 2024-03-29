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
package cmd

import (
	"github.com/fox-one/pkg/store/db"
	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:     "migrate",
	Aliases: []string{"setdb"},
	Short:   "migrate database tables",
	Run: func(cmd *cobra.Command, args []string) {
		database := provideDatabase()
		defer database.Close()

		if err := db.Migrate(database); err != nil {
			cmd.PrintErrln("migrate tables", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
