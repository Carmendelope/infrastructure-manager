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

/*
RUN_INTEGRATION_TEST=true
IT_K8S_KUBECONFIG=/Users/daniel/.kube/config
*/

package k8s

import (
	"github.com/nalej/infrastructure-manager/internal/pkg/utils"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
	"os"
)

var _ = ginkgo.Describe("A launch command", func() {

	if !utils.RunIntegrationTests() {
		log.Warn().Msg("Integration tests are skipped")
		return
	}
	var (
		kubeConfigFile = os.Getenv("IT_K8S_KUBECONFIG")
	)

	if kubeConfigFile == "" {
		ginkgo.Fail("missing environment variables")
	}

	ginkgo.It("should discover the cluster", func() {
		dh := NewDiscoveryHelper(kubeConfigFile)
		err := dh.Connect()
		gomega.Expect(err).To(gomega.Succeed())
		cluster, err := dh.Discover()
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(cluster).ShouldNot(gomega.BeNil())
		gomega.Expect(cluster.KubernetesVersion).NotTo(gomega.BeEmpty())
		gomega.Expect(len(cluster.Nodes)).Should(gomega.Equal(1))
		n0 := cluster.Nodes[0]
		gomega.Expect(n0.IP).ShouldNot(gomega.BeEmpty())
	})

})
