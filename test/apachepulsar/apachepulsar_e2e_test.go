//go:build test

/*
   Copyright 2022 The Numaproj Authors.

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

package apachepulsar

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/numaproj-contrib/numaflow-utils-go/testing/fixtures"
	pulsaradmin "github.com/streamnative/pulsar-admin-go"
	"github.com/streamnative/pulsar-admin-go/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	host                = "pulsar://localhost:6650"
	topic               = "testTopic"
	subscriptionName    = "testSubscription"
	tenant              = "public"
	namespace           = "test-namespace"
	pulsarAdminEndPoint = "http://localhost:8080"
)

type ApachePulsarSuite struct {
	fixtures.E2ESuite
}

func (suite *ApachePulsarSuite) SetupTest() {
	// Create Pulsar resources
	apachePulsarDeleteCmd := fmt.Sprintf("kubectl delete -k ../../config/apps/apachepulsar -n %s --ignore-not-found=true", fixtures.Namespace)
	suite.Given().When().Exec("sh", []string{"-c", apachePulsarDeleteCmd}, fixtures.OutputRegexp(""))
	apachePulsarCreateCmd := fmt.Sprintf("kubectl apply -k ../../config/apps/apachepulsar -n %s", fixtures.Namespace)
	suite.Given().When().Exec("sh", []string{"-c", apachePulsarCreateCmd}, fixtures.OutputRegexp("service/pulsar-broker created"))
	pulsarLabelSelector := fmt.Sprintf("app=%s", "pulsar-broker")
	suite.Given().When().WaitForPodReady("pulsar-broker-0", pulsarLabelSelector)
	suite.T().Log("apache pulsar resources are ready")
	suite.T().Log("port forwarding apache pulsar service")
	suite.StartPortForward("pulsar-broker-0", 6650)
	suite.StartPortForward("pulsar-broker-0", 8080)

}

func initClient() (pulsar.Client, error) {
	pulsarClient, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               host,
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return pulsarClient, nil
}

func initAdminClient() (pulsaradmin.Client, error) {
	cfg := &pulsaradmin.Config{
		WebServiceURL: pulsarAdminEndPoint,
	}
	pulsarAdmin, err := pulsaradmin.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return pulsarAdmin, err
}

func createTopic(pulsarAdmin pulsaradmin.Client, tenant, namespace, topic string, partitions int) error {
	topicPulsar, _ := utils.GetTopicName(fmt.Sprintf("%s/%s/%s", tenant, namespace, topic))
	err := pulsarAdmin.Topics().Create(*topicPulsar, partitions)
	if err != nil {
		return err
	}
	return nil
}

func createNameSpace(pulsarAdmin pulsaradmin.Client) error {
	err := pulsarAdmin.Namespaces().CreateNamespace(fmt.Sprintf("%s/%s", tenant, namespace))
	if err != nil {
		return err
	}
	return nil
}

func ifPulsarContainsMessage(client pulsar.Client, ctx context.Context, data string) (bool, error) {
	subscribe, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:            fmt.Sprintf("%s/%s/%s", tenant, namespace, topic),
		SubscriptionName: subscriptionName,
		Type:             pulsar.Shared,
	})
	if err != nil {
		return false, err
	}
	msg, err := subscribe.Receive(ctx)
	if err != nil {
		return false, err
	}
	if strings.Contains(string(msg.Payload()), data) {
		return true, nil
	}
	return false, nil
}

func (a *ApachePulsarSuite) TestApachePulsarSink() {
	data := "Createdts"
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	w := a.Given().Pipeline("@testdata/apachepulsar_sink.yaml").
		When().
		CreatePipelineAndWait()
	w.Expect().VertexPodsRunning()
	client, err := initClient()
	a.NoError(err)
	adminClient, err := initAdminClient()
	assert.Nil(a.T(), err)
	err = createNameSpace(adminClient)
	assert.Nil(a.T(), err)
	err = createTopic(adminClient, tenant, namespace, topic, 2)
	assert.Nil(a.T(), err)
	message, err := ifPulsarContainsMessage(client, ctx, data)
	a.NoError(err)
	a.True(message)
	w.DeletePipelineAndWait(3 * time.Minute)
}

func TestApachePulsarSourceSuite(t *testing.T) {
	suite.Run(t, new(ApachePulsarSuite))
}
