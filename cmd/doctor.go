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
	"github.com/asaskevich/govalidator"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/spf13/cobra"
)

var resourcePatterns = []string{
	"mixin://snapshots",
}

// doctorCmd represents the doctor command
var doctorCmd = &cobra.Command{
	Use: "doctor",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		client := provideMixinClient()

		for _, admin := range cfg.Group.Admins {
			if _, err := client.CreateContactConversation(ctx, admin); err != nil {
				cmd.PrintErrln("CreateContactConversation", err)
				return
			}
		}

		profile, err := client.UserMe(ctx)
		if err != nil {
			cmd.PrintErrln("UserMe", err)
			return
		}

		app := profile.App
		if app == nil {
			cmd.PrintErrln("keystore must be app")
			return
		}

		patterns := app.ResourcePatterns
		for _, p := range resourcePatterns {
			if !govalidator.IsIn(p, app.ResourcePatterns...) {
				patterns = append(patterns, p)
			}
		}

		if _, err := client.UpdateApp(ctx, client.ClientID, mixin.UpdateAppRequest{
			Category:         "TRADING",
			ResourcePatterns: patterns,
		}); err != nil {
			cmd.PrintErrln("UpdateApp", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
