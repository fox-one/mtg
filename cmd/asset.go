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
	"github.com/fox-one/mtg/core"
	"github.com/fox-one/mtg/internal/form"
	"github.com/spf13/cobra"
)

// assetCmd represents the asset command
var assetCmd = &cobra.Command{
	Use:   "asset",
	Short: "manager assets",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		database := provideDatabase()
		defer database.Close()

		assets := provideAssetStore(database, 0)
		if list, _ := cmd.Flags().GetBool("list"); list {
			allAssets, err := assets.ListAll(ctx)
			if err != nil {
				cmd.PrintErrln("list all assets", err)
				return
			}

			_ = form.Assets(allAssets).Fprint(cmd.OutOrStdout())
			return
		}

		arg, ok := getArg(args, 0)
		if !ok {
			cmd.Println("./uniswap asset top/asset id")
			return
		}

		client := provideMixinClient()
		assetz := provideAssetService(client)

		var loadedAssets []*core.Asset

		switch {
		case arg == "top":
			topAssets, err := assetz.ListAll(ctx)
			if err != nil {
				cmd.PrintErrln("list all assets", err)
				return
			}

			loadedAssets = topAssets
		default:
			asset, err := assetz.Find(ctx, arg)
			if err != nil {
				cmd.PrintErrln("find asset", err)
				return
			}

			loadedAssets = append(loadedAssets, asset)
		}

		for _, asset := range loadedAssets {
			if err := assets.Save(ctx, asset, "price"); err != nil {
				cmd.PrintErrln("save asset", err)
				return
			}

			cmd.Printf("asset %s saved\n", asset.Symbol)
		}
	},
}

func init() {
	rootCmd.AddCommand(assetCmd)
	assetCmd.Flags().BoolP("list", "l", false, "list assets")
}
