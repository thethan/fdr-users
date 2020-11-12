module github.com/thethan/fdr-users

go 1.15

require (
	cloud.google.com/go/firestore v1.2.0
	firebase.google.com/go v3.12.1+incompatible
	github.com/aws/aws-sdk-go v1.29.15
	github.com/caarlos0/env v3.5.0+incompatible
	github.com/go-kit/kit v0.10.0
	github.com/gogo/protobuf v1.3.1
	github.com/google/uuid v1.1.1
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/kubemq-io/kubemq-go v1.4.0
	github.com/kubemq-io/protobuf v1.1.0
	github.com/oklog/run v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	github.com/thethan/fdr_proto v0.0.0-20200417043340-bff7591e2122
	go.elastic.co/apm v1.8.0
	go.elastic.co/apm/module/apmgrpc v1.8.0
	go.elastic.co/apm/module/apmlogrus v1.8.0
	go.mongodb.org/mongo-driver v1.4.1
	go.opencensus.io v0.22.4
	go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux v0.12.0
	go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo v0.12.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.12.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.12.0
	go.opentelemetry.io/otel v0.12.0
	go.opentelemetry.io/otel/exporters/metric/prometheus v0.12.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.12.0
	go.opentelemetry.io/otel/sdk v0.12.0
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43
	google.golang.org/api v0.32.0
	google.golang.org/grpc v1.32.0
	google.golang.org/grpc/examples v0.0.0-20201002194053-b2c5f4a808fd // indirect

)
