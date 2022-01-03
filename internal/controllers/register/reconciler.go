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

package register

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/yndd/ndd-runtime/pkg/event"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/nddo-runtime/pkg/odr"
	"github.com/yndd/nddo-runtime/pkg/reconciler/managed"
	"github.com/yndd/nddo-runtime/pkg/resource"
	niregv1alpha1 "github.com/yndd/nddr-ni-registry/apis/registry/v1alpha1"
	"github.com/yndd/nddr-ni-registry/internal/handler"
	"github.com/yndd/nddr-ni-registry/internal/shared"
	"github.com/yndd/nddr-organization/pkg/registry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

const (
	// timers
	reconcileTimeout = 1 * time.Minute
	veryShortWait    = 1 * time.Second
	// errors
	errUnexpectedResource = "unexpected infrastructure object"
	errGetK8sResource     = "cannot get infrastructure resource"
)

// Setup adds a controller that reconciles infra.
func Setup(mgr ctrl.Manager, o controller.Options, nddcopts *shared.NddControllerOptions) error {
	name := "nddo/" + strings.ToLower(niregv1alpha1.RegisterGroupKind)
	rgfn := func() niregv1alpha1.Rg { return &niregv1alpha1.Registry{} }
	//rglfn := func() niregv1alpha1.RgList { return &niregv1alpha1.RegistryList{} }
	//rrfn := func() niregv1alpha1.Rr { return &niregv1alpha1.Register{} }
	//rrlfn := func() niregv1alpha1.RrList { return &niregv1alpha1.RegisterList{} }

	//speedy := make(map[string]int)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(niregv1alpha1.RegisterGroupVersionKind),
		managed.WithLogger(nddcopts.Logger.WithValues("controller", name)),
		managed.WithApplication(&application{
			client: resource.ClientApplicator{
				Client:     mgr.GetClient(),
				Applicator: resource.NewAPIPatchingApplicator(mgr.GetClient()),
			},
			log:         nddcopts.Logger.WithValues("applogic", name),
			newRegistry: rgfn,
			//newRegistryList: rglfn,
			//pool:    nddcopts.Pool,
			handler: nddcopts.Handler,
			//speedy:           speedy,
			registry: nddcopts.Registry,
		}),
		//managed.WithSpeedy(speedy),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&niregv1alpha1.Register{}).
		Owns(&niregv1alpha1.Register{}).
		WithEventFilter(resource.IgnoreUpdateWithoutGenerationChangePredicate()).
		WithEventFilter(resource.IgnoreUpdateWithoutGenerationChangePredicate()).
		Complete(r)

}

type application struct {
	client resource.ClientApplicator
	log    logging.Logger

	newRegistry func() niregv1alpha1.Rg

	//pool    map[string]hash.HashTable
	handler handler.Handler
	//speedy   map[string]int
	registry registry.Registry

	//poolmutex sync.Mutex
	//speedyMutex sync.Mutex
}

func getCrName(cr niregv1alpha1.Rr) string {
	odr, err := odr.GetOdrRegisterOrgFixedInfo(cr.GetName())
	if err != nil {
		return ""
	}

	return strings.Join([]string{cr.GetNamespace(), odr.FullRegistryName}, ".")
}

func (r *application) Initialize(ctx context.Context, mg resource.Managed) error {
	return nil
}

func (r *application) Update(ctx context.Context, mg resource.Managed) (map[string]string, error) {
	cr, ok := mg.(*niregv1alpha1.Register)
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
	/*
		cr, _ := mg.(*niregv1alpha1.Registry)
		crName := getCrName(cr)
		r.speedyMutex.Lock()
		speedy := r.speedy[crName]
		r.speedyMutex.Unlock()
		if speedy <= 5 {
			r.log.Debug("Speedy", "number", speedy)
			speedy++
			return veryShortWait
		}
	*/
	return reconcileTimeout
}

func (r *application) Delete(ctx context.Context, mg resource.Managed) (bool, error) {
	cr, ok := mg.(*niregv1alpha1.Register)
	if !ok {
		return true, errors.New(errUnexpectedResource)
	}
	log := r.log.WithValues("function", "handleAppLogic", "crname", cr.GetName())
	log.Debug("handleDelete")

	crName := getCrName(cr)
	odr, err := odr.GetOdrRegisterOrgFixedInfo(cr.GetName())
	if err != nil {
		return true, err
	}

	registerInfo := &handler.RegisterInfo{
		Namespace:    cr.GetNamespace(),
		RegistryName: odr.FullRegistryName,
		CrName:       crName,
		Selector:     cr.GetSelector(),
		SourceTag:    cr.GetSourceTag(),
	}

	log.Debug("resource dealloc", "registerInfo", registerInfo)

	if err := r.handler.DeRegister(ctx, registerInfo); err != nil {
		return true, err
	}

	return true, nil
}

func (r *application) FinalDelete(ctx context.Context, mg resource.Managed) {

}

func (r *application) handleAppLogic(ctx context.Context, cr niregv1alpha1.Rr) (map[string]string, error) {
	log := r.log.WithValues("function", "handleAppLogic", "crname", cr.GetName())
	log.Debug("handleAppLogic")

	crName := getCrName(cr)
	odr, err := odr.GetOdrRegisterOrgFixedInfo(cr.GetName())
	if err != nil {
		return nil, err
	}

	registerInfo := &handler.RegisterInfo{
		Namespace:    cr.GetNamespace(),
		RegistryName: odr.FullRegistryName,
		CrName:       crName,
		Selector:     cr.GetSelector(),
		SourceTag:    cr.GetSourceTag(),
	}

	log.Debug("resource alloc", "registerInfo", registerInfo)

	index, err := r.handler.Register(ctx, registerInfo)
	if err != nil {
		return nil, err
	}

	cr.SetNi(*index)

	cr.SetOrganizationName(cr.GetOrganizationName())
	cr.SetRegistryName(cr.GetRegistryName())

	return nil, nil
}
