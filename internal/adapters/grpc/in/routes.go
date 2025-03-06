package in

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/go-playground/validator.v9"
	command "k8s-golang-addons-boilerplate/internal/services/example/command"
	"k8s-golang-addons-boilerplate/internal/services/example/query"
	"k8s-golang-addons-boilerplate/pkg/log"
	"k8s-golang-addons-boilerplate/pkg/net/http"
	"k8s-golang-addons-boilerplate/pkg/opentelemetry"
	mproto "k8s-golang-addons-boilerplate/pkg/proto/example"
)

// NewRouterGRPC registers routes to the grpc.
func NewRouterGRPC(lg log.Logger, tl *opentelemetry.Telemetry, exq *query.ExampleQuery, exc *command.ExampleCommand) *grpc.Server {
	tlMid := http.NewTelemetryMiddleware(tl)

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			tlMid.WithTelemetryInterceptor(tl),
			http.WithGrpcLogging(http.WithCustomLogger(lg)),
			tlMid.EndTracingSpansInterceptor(),
		),
	)

	reflection.Register(server)

	exampleProto := &ExampleProto{
		ExampleQuery:   exq,
		ExampleCommand: exc,
		Validator:      validator.New(),
	}

	mproto.RegisterExampleServer(server, exampleProto)

	return server
}
