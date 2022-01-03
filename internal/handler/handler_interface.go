package handler

import (
	"context"

	"github.com/yndd/ndd-runtime/pkg/logging"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Option can be used to manipulate Options.
type Option func(Handler)

// WithLogger specifies how the Reconciler should log messages.
func WithLogger(log logging.Logger) Option {
	return func(s Handler) {
		s.WithLogger(log)
	}
}

/*
func WithPool(pool map[string]hash.HashTable) Option {
	return func(s Handler) {
		s.WithPool(pool)
	}
}
*/

func WithClient(c client.Client) Option {
	return func(s Handler) {
		s.WithClient(c)
	}
}

/*
func WithNewResourceFn(f func() niregv1alpha1.Rg) Option {
	return func(r Handler) {
		r.WithNewResourceFn(f)
	}
}
*/

type Handler interface {
	WithLogger(log logging.Logger)
	//WithPool(pool map[string]hash.HashTable)
	WithClient(a client.Client)
	//WithNewResourceFn(f func() niregv1alpha1.Rg)
	Init(string, uint32)
	Delete(string)
	GetAllocated(string) (uint32, []*string)
	ResetSpeedy(string)
	GetSpeedy(crName string) int
	IncrementSpeedy(crName string)
	Register(context.Context, *RegisterInfo) (*uint32, error)
	DeRegister(context.Context, *RegisterInfo) error
}
