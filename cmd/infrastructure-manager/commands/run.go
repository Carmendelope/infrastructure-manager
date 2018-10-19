/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
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
		"Applications Manager address (host:port)")
	rootCmd.AddCommand(runCmd)
}