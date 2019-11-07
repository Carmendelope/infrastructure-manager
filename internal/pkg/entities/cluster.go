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

package entities

import (
	"k8s.io/api/core/v1"
)

type Cluster struct {
	KubernetesVersion    string
	Name                 string
	Description          string
	Hostname             string
	ControlPlaneHostname string
	Nodes                []Node
}

type Node struct {
	IP     string
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
