# Apache Pulsar Sink for Numaflow

The Apache Pulsar Sink is a custom user-defined sink for [Numaflow](https://numaflow.numaproj.io/) that allows the integration of Apache Pulsar as a sink within your Numaflow pipelines. This setup enables the efficient transfer of data from Numaflow pipelines into Apache Pulsar topics.

## Quick Start
This guide will assist you in setting up an Apache Pulsar sink in a Numaflow pipeline, including configuring your Apache Pulsar environment.

### Prerequisites
* [Install Numaflow on your Kubernetes cluster](https://numaflow.numaproj.io/quick-start/)
* An Apache Pulsar cluster, either self-hosted or using a cloud service.
* Familiarity with Kubernetes and Apache Pulsar.

### Step-by-step Guide

#### 1. Set Up Apache Pulsar

Before deploying the Numaflow pipeline, ensure your Apache Pulsar cluster is up and running. Below is an example configuration for a Pulsar cluster:

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: pulsar-broker
spec:
  replicas: 1
  selector:
    matchLabels:
      app: pulsar-broker
  template:
    metadata:
      labels:
        app: pulsar-broker
    spec:
      containers:
        - name: pulsar
          image: apachepulsar/pulsar:latest
          ports:
            - containerPort: 6650
            - containerPort: 8080
          env:
            - name: PULSAR_MEM
              value: "-Xms1g -Xmx1g -XX:MaxDirectMemorySize=1g"
          command: ["bin/pulsar"]
          args: ["standalone"]
          livenessProbe:
            failureThreshold: 5
            httpGet:
              path: /admin/v2/brokers/health
              port: 8080
            initialDelaySeconds: 30
            periodSeconds: 10
            timeoutSeconds: 10
          readinessProbe:
            failureThreshold: 5
            httpGet:
                path: /admin/v2/brokers/health
                port: 8080
            initialDelaySeconds: 15
            periodSeconds: 10
            timeoutSeconds: 10
  serviceName: pulsar-broker-service
---
apiVersion: v1
kind: Service
metadata:
  name: pulsar-broker
spec:
  selector:
    app: pulsar-broker
  ports:
    - name: pulsar
      protocol: TCP
      port: 6650
      targetPort: 6650
    - name: http
      protocol: TCP
      port: 8080
      targetPort: 8080
```

Deploy this configuration to your Kubernetes cluster to establish the Pulsar broker.

#### 2. Deploy a Numaflow Pipeline with Apache Pulsar Sink

- Save the following Kubernetes manifest to a file (e.g., `pulsar-sink-pipeline.yaml`)
- Customize the configuration to match your Apache Pulsar setup.

```yaml
apiVersion: numaflow.numaproj.io/v1alpha1
kind: Pipeline
metadata:
  name: apache-pulsar-sink
spec:
  vertices:
    - name: in
      source:
        generator:
          rpu: 10
          duration: 1s
          msgSize: 100
    - name: out
      sink:
        udsink:
          container:
            image: "quay.io/numaio/numaflow-go/apache-pulsar-sink-go:latest"
            env:
              - name: PULSAR_TOPIC
                value: "testTopic"
              - name: PULSAR_SUBSCRIPTION_NAME
                value: "testSubscription"
              - name: PULSAR_HOST
                value: "pulsar-broker.numaflow-system.svc.cluster.local:6650"
              - name: PULSAR_ADMIN_ENDPOINT
                value: "http://pulsar-broker.numaflow-system.svc.cluster.local:8080"
  edges:
    - from: in
      to: out
```
Then apply it to your cluster:
```bash
kubectl apply -f pulsar-sink-pipeline.yaml
```

#### 3. Verify the Apache Pulsar Sink

Ensure that messages are being published to Apache Pulsar topics as expected.

#### 4. Clean up

To delete the Numaflow pipeline:
```bash
kubectl delete -f pulsar-sink-pipeline.yaml
```

Congratulations! You have successfully set up an Apache Pulsar sink in a Numaflow pipeline.

## Additional Resources

For more detailed information on Numaflow and its usage, visit the [Numaflow Documentation](https://numaflow.numaproj.io/). For Apache Pulsar specific configuration and setup, refer to the [Apache Pulsar Documentation](https://pulsar.apache.org/docs/en/standalone/).