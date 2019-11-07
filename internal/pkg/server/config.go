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
	// Message queue system address
	QueueAddress string
	// Debug mode
	Debug bool
}

func (conf *Config) Validate() derrors.Error {
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
	if conf.QueueAddress == "" {
		return derrors.NewInvalidArgumentError("queueAddress must be set")
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
	log.Info().Str("URL", conf.QueueAddress).Msg("Queue")
}
