/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
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

	if ! utils.RunIntegrationTests() {
		log.Warn().Msg("Integration tests are skipped")
		return
	}
	var (
		kubeConfigFile= os.Getenv("IT_K8S_KUBECONFIG")
	)

	if kubeConfigFile == "" {
		ginkgo.Fail("missing environment variables")
	}

	ginkgo.It("should discover the cluster", func(){
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
