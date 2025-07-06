/*
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

package metrics

import (
	"context"
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// NodeMetrics represents the metrics for a single node
type NodeMetrics struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Timestamp         metav1.Time     `json:"timestamp"`
	Window            metav1.Duration `json:"window"`
	Usage             v1.ResourceList `json:"usage"`
}

// NodeMetricsList is a list of NodeMetrics
type NodeMetricsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeMetrics `json:"items"`
}

// Client is a client for the Kubernetes metrics API
type Client struct {
	restClient *rest.RESTClient
}

// NewClient creates a new metrics client
func NewClient(config *rest.Config) (*Client, error) {
	gv := schema.GroupVersion{Group: "metrics.k8s.io", Version: "v1beta1"}

	configShallowCopy := *config
	configShallowCopy.GroupVersion = &gv
	configShallowCopy.APIPath = "/apis"
	configShallowCopy.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	configShallowCopy.UserAgent = rest.DefaultKubernetesUserAgent()

	restClient, err := rest.RESTClientFor(&configShallowCopy)
	if err != nil {
		return nil, fmt.Errorf("failed to create REST client: %w", err)
	}

	return &Client{
		restClient: restClient,
	}, nil
}

// GetNodeMetrics fetches metrics for all nodes
func (c *Client) GetNodeMetrics(ctx context.Context) (*NodeMetricsList, error) {
	result := &NodeMetricsList{}
	data, err := c.restClient.Get().
		Resource("nodes").
		Do(ctx).
		Raw()
	if err != nil {
		return nil, fmt.Errorf("failed to get node metrics: %w", err)
	}

	if err := json.Unmarshal(data, result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal node metrics: %w", err)
	}

	return result, nil
}

// GetNodeMetric fetches metrics for a specific node by name
func (c *Client) GetNodeMetric(ctx context.Context, nodeName string) (*NodeMetrics, error) {
	result := &NodeMetrics{}
	data, err := c.restClient.Get().
		Resource("nodes").
		Name(nodeName).
		Do(ctx).
		Raw()
	if err != nil {
		return nil, fmt.Errorf("failed to get node metrics for %s: %w", nodeName, err)
	}

	if err := json.Unmarshal(data, result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal node metrics for %s: %w", nodeName, err)
	}

	return result, nil
}
