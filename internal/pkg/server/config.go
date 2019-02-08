/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package server

import (
	"github.com/nalej/derrors"
	"github.com/nalej/infrastructure-manager/version"
	"github.com/rs/zerolog/log"
)

type Config struct {
	// Address where the API service will listen requests.
	Port int
	// Path of the temporal directory.
	TempDir string
	// SystemModelAddress with the host:port to connect to System Model
	SystemModelAddress string
	// InfrastructureManagerAddress with the host:port to connect to the Infrastructure Manager.
	ProvisionerAddress string
	// ApplicationsManagerAddress with the host:port to connect to the Applications manager.
	InstallerAddress string
}

func (conf * Config) Validate() derrors.Error {
	if conf.Port <= 0 {
		return derrors.NewInvalidArgumentError("port must be specified")
	}
	if conf.TempDir == "" {
		return derrors.NewInvalidArgumentError("tempDir must be set")
	}
	if conf.SystemModelAddress == "" {
		return derrors.NewInvalidArgumentError("systemModelAddress must be set")
	}
	if conf.InstallerAddress == "" {
		return derrors.NewInvalidArgumentError("installerAddress must be set")
	}
	return nil
}

func (conf *Config) Print() {
	log.Info().Str("app", version.AppVersion).Str("commit", version.Commit).Msg("Version")
	log.Info().Int("port", conf.Port).Msg("gRPC port")
	log.Info().Str("path", conf.TempDir).Msg("Temporal directory")
	log.Info().Str("URL", conf.SystemModelAddress).Msg("System Model")
	log.Info().Str("URL", conf.ProvisionerAddress).Msg("Provisioner")
	log.Info().Str("URL", conf.InstallerAddress).Msg("Installer")
}