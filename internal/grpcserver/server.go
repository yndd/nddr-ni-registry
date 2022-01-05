/*
Copyright 2021 NDDO.

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

package grpcserver

import (
	"context"
	"net"

	"github.com/pkg/errors"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/nddo-grpc/resource/resourcepb"
	"github.com/yndd/nddr-ni-registry/internal/handler"
	"google.golang.org/grpc"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

const (
	// errors
	errStartGRPCServer   = "cannot start GRPC server"
	errCreateTcpListener = "cannot create TCP listener"
	errGrpcServer        = "cannot serve GRPC server"
)

type server struct {
	resourcepb.UnimplementedResourceServer

	cfg Config

	// kubernetes
	client   client.Client
	eventChs map[string]chan event.GenericEvent

	//
	//poolmutex sync.Mutex
	//pool    map[string]hash.HashTable
	log     logging.Logger
	handler handler.Handler

	//newRegistry func() niv1alpha1.Rg

	// context
	ctx context.Context
}

func New(opts ...Option) (Server, error) {
	s := &server{}

	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

func (s *server) WithLogger(log logging.Logger) {
	s.log = log
}

func (s *server) WithConfig(cfg Config) {
	s.cfg = cfg
}

func (s *server) WithEventChannels(eventChs map[string]chan event.GenericEvent) {
	s.eventChs = eventChs
}

func (s *server) WithClient(c client.Client) {
	s.client = c
}

/*
func (s *server) WithNewResourceFn(f func() niv1alpha1.Rg) {
	s.newRegistry = f
}
*/

func (s *server) WithHandler(h handler.Handler) {
	s.handler = h
}

func (s *server) Run(ctx context.Context) error {
	log := s.log.WithValues("grpcServerAddress", s.cfg.Address)
	log.Debug("grpc server run...")
	s.ctx = ctx
	errChannel := make(chan error)
	go func() {
		if err := s.start(); err != nil {
			errChannel <- errors.Wrap(err, errStartGRPCServer)
		}
		errChannel <- nil
	}()
	return nil
}

// Start GRPC Server
func (s *server) start() error {
	log := s.log.WithValues("grpcServerAddress", s.cfg.Address)
	log.Debug("grpc server start...")

	// create a listener on a specific address:port
	l, err := net.Listen("tcp", s.cfg.Address)
	if err != nil {
		return errors.Wrap(err, errCreateTcpListener)
	}

	// TODO, proper handling of the certificates with CERT Manager
	/*
		opts, err := s.serverOpts()
		if err != nil {
			return err
		}
	*/
	// create a gRPC server object
	grpcServer := grpc.NewServer()

	// attach the gRPC service to the server
	resourcepb.RegisterResourceServer(grpcServer, s)

	// start the server
	log.Debug("grpc server serve...")
	if err := grpcServer.Serve(l); err != nil {
		s.log.Debug("Errors", "error", err)
		return errors.Wrap(err, errGrpcServer)
	}
	return nil
}
