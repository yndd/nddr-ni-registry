/*
Copyright 2021 NDD.

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

package registry

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/yndd/ndd-runtime/pkg/event"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/nddo-runtime/pkg/reconciler/managed"
	"github.com/yndd/nddo-runtime/pkg/resource"
	niregv1alpha1 "github.com/yndd/nddr-ni-registry/apis/registry/v1alpha1"
	"github.com/yndd/nddr-ni-registry/internal/handler"
	"github.com/yndd/nddr-ni-registry/internal/shared"
	"github.com/yndd/nddr-organization/pkg/registry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	gevent "sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	// timers
	reconcileTimeout = 1 * time.Minute
	shortWait        = 5 * time.Second
	veryShortWait    = 1 * time.Second
	// errors
	errUnexpectedResource = "unexpected infrastructure object"
	errGetK8sResource     = "cannot get infrastructure resource"
)

// Setup adds a controller that reconciles infra.
func Setup(mgr ctrl.Manager, o controller.Options, nddcopts *shared.NddControllerOptions) (string, chan gevent.GenericEvent, error) {
	name := "nddo/" + strings.ToLower(niregv1alpha1.RegistryGroupKind)
	rgfn := func() niregv1alpha1.Rg { return &niregv1alpha1.Registry{} }
	rglfn := func() niregv1alpha1.RgList { return &niregv1alpha1.RegistryList{} }
	//rrfn := func() niregv1alpha1.Rr { return &niregv1alpha1.Register{} }
	//rrlfn := func() niregv1alpha1.RrList { return &niregv1alpha1.RegisterList{} }

	events := make(chan gevent.GenericEvent)
	//speedy := make(map[string]int)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(niregv1alpha1.RegistryGroupVersionKind),
		managed.WithLogger(nddcopts.Logger.WithValues("controller", name)),
		managed.WithApplication(&application{
			client: resource.ClientApplicator{
				Client:     mgr.GetClient(),
				Applicator: resource.NewAPIPatchingApplicator(mgr.GetClient()),
			},
			log:             nddcopts.Logger.WithValues("applogic", name),
			newRegistry:     rgfn,
			newRegistryList: rglfn,
			registry:        nddcopts.Registry,
			handler:         nddcopts.Handler,
		}),
		//managed.WithSpeedy(speedy),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	)

	registerHandler := &EnqueueRequestForAllRegisters{
		client:          mgr.GetClient(),
		log:             nddcopts.Logger,
		ctx:             context.Background(),
		handler:         nddcopts.Handler,
		newRegistryList: rglfn,
	}

	return niregv1alpha1.RegistryGroupKind, events, ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&niregv1alpha1.Registry{}).
		Owns(&niregv1alpha1.Registry{}).
		WithEventFilter(resource.IgnoreUpdateWithoutGenerationChangePredicate()).
		Watches(&source.Kind{Type: &niregv1alpha1.Register{}}, registerHandler).
		Watches(&source.Channel{Source: events}, registerHandler).
		WithEventFilter(resource.IgnoreUpdateWithoutGenerationChangePredicate()).
		Complete(r)

}

type application struct {
	client resource.ClientApplicator
	log    logging.Logger

	newRegistry     func() niregv1alpha1.Rg
	newRegistryList func() niregv1alpha1.RgList

	registry registry.Registry
	handler  handler.Handler
}

func getCrName(cr niregv1alpha1.Rg) string {
	return strings.Join([]string{cr.GetNamespace(), cr.GetName()}, ".")
}

func (r *application) Initialize(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*niregv1alpha1.Registry)
	if !ok {
		return errors.New(errUnexpectedResource)
	}

	if err := cr.InitializeResource(); err != nil {
		r.log.Debug("Cannot initialize", "error", err)
		return err
	}

	return nil
}

func (r *application) Update(ctx context.Context, mg resource.Managed) (map[string]string, error) {
	cr, ok := mg.(*niregv1alpha1.Registry)
	if !ok {
		return nil, errors.New(errUnexpectedResource)
	}

	return r.handleAppLogic(ctx, cr)
}

func (r *application) FinalUpdate(ctx context.Context, mg resource.Managed) {
	//cr, _ := mg.(*niregv1alpha1.Registry)
	//crName := getCrName(cr)
	//r.infra[crName].PrintNodes(crName)
}

func (r *application) Timeout(ctx context.Context, mg resource.Managed) time.Duration {
	cr, _ := mg.(*niregv1alpha1.Registry)
	crName := getCrName(cr)
	speedy := r.handler.GetSpeedy(crName)
	if speedy <= 2 {
		r.handler.IncrementSpeedy(crName)
		r.log.Debug("Speedy incr", "number", r.handler.GetSpeedy(crName))
		switch speedy {
		case 0:
			return veryShortWait
		case 1, 2:
			return shortWait
		}

	}
	return reconcileTimeout
}

func (r *application) Delete(ctx context.Context, mg resource.Managed) (bool, error) {
	cr, ok := mg.(*niregv1alpha1.Registry)
	if !ok {
		return false, errors.New(errUnexpectedResource)
	}
	crName := getCrName(cr)

	allocated, _ := r.handler.GetAllocated(crName)
	if allocated > 0 {
		return false, nil
	}
	return true, nil
}

func (r *application) FinalDelete(ctx context.Context, mg resource.Managed) {
	cr, _ := mg.(*niregv1alpha1.Registry)
	crName := getCrName(cr)
	r.handler.Delete(crName)
}

func (r *application) handleAppLogic(ctx context.Context, cr niregv1alpha1.Rg) (map[string]string, error) {
	log := r.log.WithValues("function", "handleAppLogic", "crname", cr.GetName())
	log.Debug("handleAppLogic")

	//organizationName := cr.GetOrganizationName()

	// initialize speedy
	crName := getCrName(cr)
	r.handler.Init(crName, cr.GetSize())
	// update status based on a scan of the pool

	allocated, used := r.handler.GetAllocated(crName)
	log.Debug("handleAppLogic", "allocated", allocated, "used", used)
	cr.SetStatus(allocated, used)

	cr.SetOrganizationName(cr.GetOrganizationName())
	cr.SetRegistryName(cr.GetRegistryName())

	// trick to use speedy for fast updates
	return map[string]string{"dummy": "dummy"}, nil
}
