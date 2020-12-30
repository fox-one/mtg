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
	"encoding/json"
	"time"

	"github.com/spf13/cobra"
)

// multisigCmd represents the multisig command
var multisigCmd = &cobra.Command{
	Use:   "multisig",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		client := provideMixinClient()
		sigs, err := client.ReadMultisigs(ctx, time.Time{}, 500)
		if err != nil {
			cmd.PrintErrln(err)
			return
		}

		var idx int
		for _, sig := range sigs {
			if sig.State == "signed" {
				sigs[idx] = sig
				idx++
			}
		}

		sigs = sigs[:idx]

		cmd.PrintErrln(len(sigs))
		_ = json.NewEncoder(cmd.OutOrStdout()).Encode(sigs)
	},
}

func init() {
	rootCmd.AddCommand(multisigCmd)
}
