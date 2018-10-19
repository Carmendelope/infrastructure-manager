/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package entities

import(
	"k8s.io/api/core/v1"
)

type Cluster struct {
	KubernetesVersion string
	Name string
	Description string
	Hostname string
	Nodes [] Node
}

type Node struct {
	IP string
	Labels map[string]string
}

func NewNode(n v1.Node) *Node {
	ip := "NotFound"
	if len(n.Status.Addresses) > 0 {
		ip = n.Status.Addresses[0].Address
	}
	return &Node{
		IP:     ip,
		Labels: n.Labels,
	}
}