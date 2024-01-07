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

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	sinksdk "github.com/numaproj/numaflow-go/pkg/sinker"

	"github.com/numaproj-contrib/apache-pulsar-sink-go/pkg/apachepulsar"
)

func main() {
	topic := os.Getenv("PULSAR_TOPIC")
	host := os.Getenv("PULSAR_HOST")
	tenant := os.Getenv("PULSAR_TENANT")
	nameSpace := os.Getenv("PULSAR_NAMESPACE")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               fmt.Sprintf("pulsar://%s", host),
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
	})

	if err != nil {
		log.Fatalf("could not instantiate Pulsar client: %v", err)
	}

	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: fmt.Sprintf("%s/%s/%s", tenant, nameSpace, topic),
	})

	if err != nil {
		log.Fatalf("could not create  Pulsar producer: %v", err)
	}

	pulsarSink := apachepulsar.NewPulsarSink(client, producer)
	if err := sinksdk.NewServer(pulsarSink).Start(ctx); err != nil {
		log.Fatalf("failed to start sinker server: %s", err)
	}
}
