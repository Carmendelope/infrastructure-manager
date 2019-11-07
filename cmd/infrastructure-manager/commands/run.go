/*
 * Copyright 2019 Nalej
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// This is an example of an executable command.

package commands

import (
	"github.com/nalej/infrastructure-manager/internal/pkg/server"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var config = server.Config{}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Launch the server API",
	Long:  `Launch the server API`,
	Run: func(cmd *cobra.Command, args []string) {
		SetupLogging()
		log.Info().Msg("Launching API!")
		config.Debug = debugLevel
		server := server.NewService(config)
		server.Run()
	},
}

func init() {
	runCmd.Flags().IntVar(&config.Port, "port", 8081, "Port to launch the Public API")
	runCmd.PersistentFlags().StringVar(&config.SystemModelAddress, "systemModelAddress", "localhost:8800",
		"System Model address (host:port)")
	runCmd.PersistentFlags().StringVar(&config.ProvisionerAddress, "provisionerAddress", "localhost:8930",
		"Infrastructure Manager address (host:port)")
	runCmd.PersistentFlags().StringVar(&config.InstallerAddress, "installerAddress", "localhost:8900",
		"Installer address (host:port)")
	runCmd.PersistentFlags().StringVar(&config.QueueAddress, "queueAddress", "localhost:6650",
		"Queue system address (host:port)")
	runCmd.PersistentFlags().StringVar(&config.TempDir, "tempDir", "", "Temporal directory for install related files")
	rootCmd.AddCommand(runCmd)
}
