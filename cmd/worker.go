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
	"fmt"
	"net/http"
	"time"

	"github.com/fox-one/mtg/handler/hc"
	"github.com/fox-one/mtg/notifier"
	"github.com/fox-one/mtg/worker"
	"github.com/fox-one/mtg/worker/cashier"
	"github.com/fox-one/mtg/worker/pricesync"
	"github.com/fox-one/mtg/worker/spentsync"
	"github.com/fox-one/mtg/worker/syncer"
	"github.com/fox-one/mtg/worker/txsender"
	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/store/db"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

// workerCmd represents the worker command
var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "run mtg worker",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		cfg.DB.ReadHost = ""
		database := provideDatabase()
		defer database.Close()
		// migrate db tables
		if err := db.Migrate(database); err != nil {
			panic(err)
		}

		client := provideMixinClient()
		property := providePropertyStore(database)
		assets := provideAssetStore(database, time.Hour)
		assetz := provideAssetService(client)
		wallets := provideWalletStore(database)
		walletz := provideWalletService(client)
		messages := provideMessageStore(database)
		system := provideSystem()

		notify := notifier.Mute()
		if ok, _ := cmd.Flags().GetBool("notify"); ok {
			notify = notifier.New(system, assets, messages)
		}

		workers := []worker.Worker{
			txsender.New(wallets),
			spentsync.New(wallets, notify),
			cashier.New(wallets, walletz, system),
			syncer.New(assets, assetz, wallets, walletz, property, system),
			pricesync.New(assets, assetz),
		}

		// worker api
		{
			mux := chi.NewMux()
			mux.Use(middleware.Recoverer)
			mux.Use(middleware.StripSlashes)
			mux.Use(cors.AllowAll().Handler)
			mux.Use(logger.WithRequestID)
			mux.Use(middleware.Logger)

			// hc
			{
				mux.Mount("/hc", hc.Handle(rootCmd.Version))
			}

			// launch server
			port, _ := cmd.Flags().GetInt("port")
			addr := fmt.Sprintf(":%d", port)

			go http.ListenAndServe(addr, mux)
		}

		cmd.Printf("mtg worker with version %q launched!\n", rootCmd.Version)

		g, ctx := errgroup.WithContext(ctx)
		for idx := range workers {
			w := workers[idx]
			g.Go(func() error {
				return w.Run(ctx)
			})
		}

		if err := g.Wait(); err != nil {
			cmd.PrintErrln("run worker", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(workerCmd)
	workerCmd.Flags().Int("port", 9245, "worker api port")
	workerCmd.Flags().Bool("notify", false, "notify")
}
