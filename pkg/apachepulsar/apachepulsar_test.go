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
	"log"
	"os"
	"testing"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	sinksdk "github.com/numaproj/numaflow-go/pkg/sinker"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"

	"github.com/numaproj-contrib/apache-pulsar-sink-go/pkg/mocks"
)

var (
	pulsarClient pulsar.Client
	resource     *dockertest.Resource
	pool         *dockertest.Pool
)

const (
	host             = "pulsar://localhost:6650"
	topic            = "test-topic"
	subscriptionName = "test-subscription"
)

func initProducer(client pulsar.Client) (pulsar.Producer, error) {
	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: topic,
	})

	if err != nil {
		return nil, err
	}
	return producer, nil
}

func TestMain(m *testing.M) {
	var err error
	p, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not connect to docker ;is it running ? %s", err)
	}
	pool = p
	opts := dockertest.RunOptions{
		Repository:   "apachepulsar/pulsar",
		Tag:          "3.1.1",
		ExposedPorts: []string{"6650/tcp", "8080/tcp"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"6650/tcp": {
				{HostIP: "127.0.0.1", HostPort: "6650"},
			},
			"8080/tcp": {
				{HostIP: "127.0.0.1", HostPort: "8080"},
			},
		},
		Mounts: []string{"pulsardata:/pulsar/data", "pulsarconf:/pulsar/conf"},
		Cmd:    []string{"bin/pulsar", "standalone"},
	}
	resource, err = pool.RunWithOptions(&opts)
	if err != nil {
		log.Println(err)
		_ = pool.Purge(resource)
		log.Fatalf("could not start resource %s", err)
	}
	if err != nil {
		log.Fatalf("error -%s", err)
	}
	if err := pool.Retry(func() error {
		pulsarClient, err = pulsar.NewClient(pulsar.ClientOptions{
			URL:               host,
			OperationTimeout:  30 * time.Second,
			ConnectionTimeout: 30 * time.Second,
		})
		if err != nil {
			log.Fatalf("failed to create pulsar client: %v", err)
		}
		return nil
	}); err != nil {
		if resource != nil {
			_ = pool.Purge(resource)
		}
		log.Fatalf("could not connect to apache pulsar %s", err)
	}
	defer pulsarClient.Close()
	code := m.Run()
	if resource != nil {
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Couln't purge resource %s", err)
		}
	}
	os.Exit(code)
}

func TestPulsarSink_Sink(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	producer, err := initProducer(pulsarClient)
	assert.Nil(t, err)
	pulsarSink := NewPulsarSink(pulsarClient, producer)
	ch := make(chan sinksdk.Datum, 20)
	for i := 0; i < 20; i++ {
		ch <- mocks.Payload{
			Data: "pubsub test",
		}
	}
	close(ch)
	response := pulsarSink.Sink(ctx, ch)
	assert.Equal(t, 20, len(response.Items()))
}
