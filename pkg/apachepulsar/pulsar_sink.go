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

	"github.com/AthenZ/athenz/libs/go/athenz-common/log"
	"github.com/apache/pulsar-client-go/pulsar"
	sinksdk "github.com/numaproj/numaflow-go/pkg/sinker"

	"github.com/numaproj-contrib/apache-pulsar-sink-go/pkg/payloads"
)

type PulsarSink struct {
	client   pulsar.Client
	producer pulsar.Producer
}

func NewPulsarSink(client pulsar.Client, producer pulsar.Producer) *PulsarSink {
	return &PulsarSink{
		client:   client,
		producer: producer,
	}
}

func (ps *PulsarSink) Sink(ctx context.Context, datumStreamCh <-chan sinksdk.Datum) sinksdk.Responses {
	responses := sinksdk.ResponsesBuilder()
	var results []payloads.SendResult
	for datum := range datumStreamCh {
		send, err := ps.producer.Send(ctx, &pulsar.ProducerMessage{
			Payload: datum.Value(),
		})
		result := payloads.SendResult{
			MessageID: send,
			Err:       err,
		}
		results = append(results, result)
		if err != nil {
			log.Printf("error sending message to pulsar sink %s", err)
		}
	}
	for _, res := range results {
		if res.Err != nil {
			responses = append(responses, sinksdk.ResponseFailure(res.MessageID.String(), res.Err.Error()))
		}
		responses = append(responses, sinksdk.ResponseOK(res.MessageID.String()))
	}
	return responses
}
