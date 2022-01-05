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

	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/nddr-ni-registry/internal/handler"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type Config struct {
	// Address
	Address string
	// Generic
	MaxSubscriptions int64
	MaxUnaryRPC      int64
	// TLS
	InSecure   bool
	SkipVerify bool
	CaFile     string
	CertFile   string
	KeyFile    string
	// observability
	EnableMetrics bool
	Debug         bool
}

// Option can be used to manipulate Options.
type Option func(Server)

// WithLogger specifies how the Reconciler should log messages.
func WithLogger(log logging.Logger) Option {
	return func(s Server) {
		s.WithLogger(log)
	}
}

func WithConfig(cfg Config) Option {
	return func(s Server) {
		s.WithConfig(cfg)
	}
}

func WithEventChannels(eventChs map[string]chan event.GenericEvent) Option {
	return func(s Server) {
		s.WithEventChannels(eventChs)
	}
}

func WithClient(c client.Client) Option {
	return func(s Server) {
		s.WithClient(c)
	}
}

/*
func WithNewResourceFn(f func() niv1alpha1.Rg) Option {
	return func(r Server) {
		r.WithNewResourceFn(f)
	}
}
*/

func WithHandler(h handler.Handler) Option {
	return func(r Server) {
		r.WithHandler(h)
	}
}

type Server interface {
	WithLogger(log logging.Logger)
	WithConfig(cfg Config)
	WithEventChannels(map[string]chan event.GenericEvent)
	WithClient(a client.Client)
	//WithNewResourceFn(f func() niv1alpha1.Rg)
	WithHandler(handler.Handler)
	Run(ctx context.Context) error
}
