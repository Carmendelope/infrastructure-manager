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

// TODO Refactor this code an move it outside of the infrastructure-manager

package k8s

import (
	"github.com/nalej/derrors"
	"github.com/nalej/infrastructure-manager/internal/pkg/entities"
	"github.com/rs/zerolog/log"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"net/url"
)

type DiscoveryHelper struct {
	KubeConfigPath string
	Client         *kubernetes.Clientset
	ClusterName    string
	Server         string
}

func NewDiscoveryHelper(kubeConfigPath string) *DiscoveryHelper {
	return &DiscoveryHelper{
		KubeConfigPath: kubeConfigPath,
	}
}

// Validate checks that the KubeConfig file is valid and preloads basic data of the target cluster.
func (dh *DiscoveryHelper) Validate() derrors.Error {
	cclr := clientcmd.ClientConfigLoadingRules{ExplicitPath: dh.KubeConfigPath}
	cfg, err := cclr.Load()
	if err != nil {
		return derrors.AsError(err, "cannot load kubeconfig file")
	}

	if len(cfg.Clusters) != 1 {
		return derrors.NewInvalidArgumentError("kubeconfig must contain a single cluster")
	}

	for k, v := range cfg.Clusters {
		dh.ClusterName = k
		dh.Server = v.Server
		log.Debug().Str("name", dh.ClusterName).Str("server", dh.Server).Msg("New cluster added from config")
	}

	return nil
}

// Connect processes the configuration validating it first, and establishes the client channels with the cluster.
func (dh *DiscoveryHelper) Connect() derrors.Error {
	vErr := dh.Validate()
	if vErr != nil {
		return vErr
	}

	config, err := clientcmd.BuildConfigFromFlags("", dh.KubeConfigPath)
	if err != nil {
		log.Error().Err(err).Msg("error building configuration from kubeconfig")
		return derrors.AsError(err, "error building configuration from kubeconfig")
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error().Err(err).Msg("error using configuration to build k8s clientset")
		return derrors.AsError(err, "error using configuration to build k8s clientset")
	}

	dh.Client = clientset
	return nil
}

func (dh *DiscoveryHelper) Discover() (*entities.Cluster, derrors.Error) {
	sv, err := dh.Client.Discovery().ServerVersion()
	if err != nil {
		return nil, derrors.AsError(err, "cannot read version")
	}

	opts := metaV1.ListOptions{}
	nodeList, err := dh.Client.CoreV1().Nodes().List(opts)
	if err != nil {
		return nil, derrors.AsError(err, "cannot read resources")
	}

	nodes := make([]entities.Node, 0)
	for _, node := range nodeList.Items {
		toAdd := entities.NewNode(node)
		nodes = append(nodes, *toAdd)
	}

	u, err := url.Parse(dh.Server)
	if err != nil {
		return nil, derrors.AsError(err, "cannot parse server into URL")
	}

	/*
		host, _, err := net.SplitHostPort(u.Host)
		if err != nil{
			return nil, derrors.AsError(err, "cannot extract target cluster hostname")
		}
	*/

	return &entities.Cluster{
		KubernetesVersion:    sv.String(),
		Name:                 dh.ClusterName,
		Description:          "Autodiscovered cluster",
		ControlPlaneHostname: u.Hostname(),
		Nodes:                nodes,
	}, nil
}
