package handler

import (
	"context"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"github.com/yndd/ndd-runtime/pkg/logging"
	niregv1alpha1 "github.com/yndd/nddr-ni-registry/apis/ni/v1alpha1"
	"github.com/yndd/nddr-ni-registry/internal/hash"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func New(opts ...Option) (Handler, error) {
	rgfn := func() niregv1alpha1.Rg { return &niregv1alpha1.Registry{} }
	s := &handler{
		pool:        make(map[string]hash.HashTable),
		speedy:      make(map[string]int),
		newRegistry: rgfn,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

func (r *handler) WithLogger(log logging.Logger) {
	r.log = log
}

//func (r *handler) WithPool(pool map[string]hash.HashTable) {
//	r.pool = pool
//}

func (r *handler) WithClient(c client.Client) {
	r.client = c
}

//func (r *handler) WithNewResourceFn(f func() niregv1alpha1.Rg) {
//	r.newRegistry = f
//}

type RegisterInfo struct {
	Namespace    string
	RegisterName string
	RegistryName string
	CrName       string
	Selector     map[string]string
	SourceTag    map[string]string
}

type handler struct {
	log logging.Logger
	// kubernetes
	client client.Client

	newRegistry func() niregv1alpha1.Rg
	poolMutex   sync.Mutex
	pool        map[string]hash.HashTable
	speedyMutex sync.Mutex
	speedy      map[string]int
}

func (r *handler) Init(crName string, size uint32) {
	r.poolMutex.Lock()
	defer r.poolMutex.Unlock()
	if _, ok := r.pool[crName]; !ok {
		r.pool[crName] = hash.New(size)
	}

	r.speedyMutex.Lock()
	defer r.speedyMutex.Unlock()
	if _, ok := r.speedy[crName]; !ok {
		r.speedy[crName] = 0
	}
}

func (r *handler) Delete(crName string) {
	r.poolMutex.Lock()
	defer r.poolMutex.Unlock()
	delete(r.pool, crName)

	r.speedyMutex.Lock()
	defer r.speedyMutex.Unlock()
	delete(r.speedy, crName)
}

func (r *handler) GetAllocated(crName string) (uint32, []*string) {
	r.poolMutex.Lock()
	defer r.poolMutex.Unlock()
	if pool, ok := r.pool[crName]; ok {
		return pool.GetAllocated()
	}
	return 0, make([]*string, 0)
}

func (r *handler) ResetSpeedy(crName string) {
	r.speedyMutex.Lock()
	defer r.speedyMutex.Unlock()
	if _, ok := r.speedy[crName]; ok {
		r.speedy[crName] = 0
	}
}

func (r *handler) GetSpeedy(crName string) int {
	r.speedyMutex.Lock()
	defer r.speedyMutex.Unlock()
	if _, ok := r.speedy[crName]; ok {
		return r.speedy[crName]
	}
	return 9999
}

func (r *handler) IncrementSpeedy(crName string) {
	r.speedyMutex.Lock()
	defer r.speedyMutex.Unlock()
	if _, ok := r.speedy[crName]; ok {
		r.speedy[crName]++
	}
}

func (r *handler) Register(ctx context.Context, info *RegisterInfo) (*uint32, error) {
	pool, niName, err := r.validateRegister(ctx, info)
	if err != nil {
		return nil, err
	}
	registerName := info.RegisterName
	sourceTag := info.SourceTag

	r.log.Debug("pool insert", "niName", niName)
	index := pool.Insert(*niName, registerName, sourceTag)
	r.log.Debug("pool inserted", "niName", niName, "index", index)

	return &index, nil
}

func (r *handler) DeRegister(ctx context.Context, info *RegisterInfo) error {

	pool, niName, err := r.validateRegister(ctx, info)
	if err != nil {
		return err
	}
	registerName := info.RegisterName
	sourceTag := info.SourceTag

	r.log.Debug("pool delete", "niName", niName)
	pool.Delete(*niName, registerName, sourceTag)
	r.log.Debug("pool deleted", "niName", niName)

	return nil
}

func (r *handler) validateRegister(ctx context.Context, info *RegisterInfo) (hash.HashTable, *string, error) {
	namespace := info.Namespace
	registryName := info.RegistryName
	crName := info.CrName
	selector := info.Selector

	// find registry in k8s api
	registry := r.newRegistry()
	if err := r.client.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      registryName}, registry); err != nil {
		// can happen when the ipam is not found
		r.log.Debug("registry not found")
		return nil, nil, errors.Wrap(err, "registry not found")
	}

	// check is registry is ready
	if registry.GetCondition(niregv1alpha1.ConditionKindReady).Status != corev1.ConditionTrue {
		r.log.Debug("Registry not ready")
		return nil, nil, errors.New("Registry not ready")
	}

	// check if the supplied info is available
	if _, ok := selector["name"]; !ok {
		return nil, nil, errors.New("selector does not contain a name")
	}
	niName := selector["name"]

	// check if the pool/register is ready to handle new registrations
	r.poolMutex.Lock()
	defer r.poolMutex.Unlock()
	if _, ok := r.pool[crName]; !ok {
		r.log.Debug("pool/tree not ready", "crName", crName)
		return nil, nil, fmt.Errorf("pool/tree not ready, crName: %s", crName)
	}
	pool := r.pool[crName]

	return pool, &niName, nil
}
