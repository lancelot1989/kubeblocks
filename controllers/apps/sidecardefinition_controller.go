/*
Copyright (C) 2022-2024 ApeCloud Co., Ltd

This file is part of KubeBlocks project

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package apps

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"slices"
	"strings"

	"golang.org/x/exp/maps"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1 "github.com/apecloud/kubeblocks/apis/apps/v1"
	"github.com/apecloud/kubeblocks/pkg/constant"
	"github.com/apecloud/kubeblocks/pkg/controller/component"
	intctrlutil "github.com/apecloud/kubeblocks/pkg/controllerutil"
)

// SidecarDefinitionReconciler reconciles a SidecarDefinition object
type SidecarDefinitionReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=apps.kubeblocks.io,resources=sidecardefinitions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.kubeblocks.io,resources=sidecardefinitions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.kubeblocks.io,resources=sidecardefinitions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *SidecarDefinitionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqCtx := intctrlutil.RequestCtx{
		Ctx:      ctx,
		Req:      req,
		Log:      log.FromContext(ctx).WithValues("sidecarDefinition", req.NamespacedName),
		Recorder: r.Recorder,
	}

	sidecarDef := &appsv1.SidecarDefinition{}
	if err := r.Client.Get(reqCtx.Ctx, reqCtx.Req.NamespacedName, sidecarDef); err != nil {
		return intctrlutil.CheckedRequeueWithError(err, reqCtx.Log, "")
	}

	if res, err := intctrlutil.HandleCRDeletion(reqCtx, r, sidecarDef,
		sidecarDefinitionFinalizerName, r.deletionHandler(reqCtx, sidecarDef)); res != nil {
		return *res, err
	}

	if sidecarDef.Status.ObservedGeneration == sidecarDef.Generation &&
		sidecarDef.Status.Phase == appsv1.AvailablePhase {
		return intctrlutil.Reconciled()
	}

	if err := r.validate(r.Client, reqCtx, sidecarDef); err != nil {
		fmt.Printf("error: %v\n", err)
		if err1 := r.unavailable(reqCtx, sidecarDef, err); err1 != nil {
			return intctrlutil.CheckedRequeueWithError(err1, reqCtx.Log, "")
		}
		return intctrlutil.CheckedRequeueWithError(err, reqCtx.Log, "")
	}

	if err := r.immutableHash(r.Client, reqCtx, sidecarDef); err != nil {
		return intctrlutil.CheckedRequeueWithError(err, reqCtx.Log, "")
	}

	if err := r.available(reqCtx, sidecarDef); err != nil {
		return intctrlutil.CheckedRequeueWithError(err, reqCtx.Log, "")
	}

	intctrlutil.RecordCreatedEvent(r.Recorder, sidecarDef)

	return intctrlutil.Reconciled()
}

// SetupWithManager sets up the controller with the Manager.
func (r *SidecarDefinitionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return intctrlutil.NewNamespacedControllerManagedBy(mgr).
		For(&appsv1.SidecarDefinition{}).
		Watches(&appsv1.ComponentDefinition{}, handler.EnqueueRequestsFromMapFunc(r.matchedCompDefinition)).
		Complete(r)
}

func (r *SidecarDefinitionReconciler) matchedCompDefinition(ctx context.Context, obj client.Object) []reconcile.Request {
	compDef, ok := obj.(*appsv1.ComponentDefinition)
	if !ok {
		return nil
	}
	sidecarDefs := &appsv1.SidecarDefinitionList{}
	if err := r.Client.List(ctx, sidecarDefs); err != nil {
		return nil
	}
	requests := make([]reconcile.Request, 0)
	for _, sidecarDef := range sidecarDefs.Items {
		names := append([]string{sidecarDef.Spec.Owner}, sidecarDef.Spec.Selectors...)
		if slices.ContainsFunc(names, func(name string) bool {
			return component.DefNameMatched(compDef.Name, name)
		}) {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name: sidecarDef.Name,
				},
			})
		}
	}
	return requests
}

func (r *SidecarDefinitionReconciler) deletionHandler(rctx intctrlutil.RequestCtx, sidecarDef *appsv1.SidecarDefinition) func() (*ctrl.Result, error) {
	return func() (*ctrl.Result, error) {
		recordEvent := func() {
			r.Recorder.Event(sidecarDef, corev1.EventTypeWarning, "ExistsReferencedResources",
				"cannot be deleted because of existing referencing Cluster")
		}
		if res, err := intctrlutil.ValidateReferenceCR(rctx, r.Client, sidecarDef, constant.SidecarDefLabelKey,
			recordEvent, &appsv1.ClusterList{}); res != nil || err != nil {
			return res, err
		}
		return nil, nil
	}
}

func (r *SidecarDefinitionReconciler) available(rctx intctrlutil.RequestCtx, sidecarDef *appsv1.SidecarDefinition) error {
	return r.status(rctx, sidecarDef, appsv1.AvailablePhase, "")
}

func (r *SidecarDefinitionReconciler) unavailable(rctx intctrlutil.RequestCtx, sidecarDef *appsv1.SidecarDefinition, err error) error {
	message := ""
	if err != nil {
		message = err.Error()
	}
	return r.status(rctx, sidecarDef, appsv1.UnavailablePhase, message)
}

func (r *SidecarDefinitionReconciler) status(rctx intctrlutil.RequestCtx,
	sidecarDef *appsv1.SidecarDefinition, phase appsv1.Phase, message string) error {
	patch := client.MergeFrom(sidecarDef.DeepCopy())
	sidecarDef.Status.ObservedGeneration = sidecarDef.Generation
	sidecarDef.Status.Phase = phase
	sidecarDef.Status.Message = message
	return r.Client.Status().Patch(rctx.Ctx, sidecarDef, patch)
}

func (r *SidecarDefinitionReconciler) validate(cli client.Client, rctx intctrlutil.RequestCtx, sidecarDef *appsv1.SidecarDefinition) error {
	for _, validator := range []func(context.Context, client.Client, *appsv1.SidecarDefinition) error{
		r.validateOwner,
		r.validateSelectors,
	} {
		if err := validator(rctx.Ctx, cli, sidecarDef); err != nil {
			return err
		}
	}
	return r.immutableCheck(sidecarDef)
}

func (r *SidecarDefinitionReconciler) validateOwner(ctx context.Context, cli client.Client,
	sidecarDef *appsv1.SidecarDefinition) error {
	owner := sidecarDef.Spec.Owner
	if len(owner) == 0 {
		return fmt.Errorf("owner is required")
	}
	if err := component.ValidateDefNameRegexp(owner); err != nil {
		return err
	}

	compDefList := &appsv1.ComponentDefinitionList{}
	if err := cli.List(ctx, compDefList); err != nil {
		return err
	}
	for _, compDef := range compDefList.Items {
		if component.DefNameMatched(compDef.Name, owner) {
			return nil
		}
	}
	return fmt.Errorf("no matched owner found: %s", owner)
}

func (r *SidecarDefinitionReconciler) validateSelectors(ctx context.Context, cli client.Client,
	sidecarDef *appsv1.SidecarDefinition) error {
	selectors := sidecarDef.Spec.Selectors
	for _, selector := range selectors {
		if err := component.ValidateDefNameRegexp(selector); err != nil {
			return err
		}
	}
	return nil
}

func (r *SidecarDefinitionReconciler) immutableCheck(sidecarDef *appsv1.SidecarDefinition) error {
	if r.skipImmutableCheck(sidecarDef) {
		return nil
	}

	newHashValue, err := r.specHash(sidecarDef)
	if err != nil {
		return err
	}

	hashValue, ok := sidecarDef.Annotations[immutableHashAnnotationKey]
	if ok && hashValue != newHashValue {
		// TODO: fields been updated
		return fmt.Errorf("immutable fields can't be updated")
	}
	return nil
}

func (r *SidecarDefinitionReconciler) skipImmutableCheck(sidecarDef *appsv1.SidecarDefinition) bool {
	if sidecarDef.Annotations == nil {
		return false
	}
	skip, ok := sidecarDef.Annotations[constant.SkipImmutableCheckAnnotationKey]
	return ok && strings.ToLower(skip) == "true"
}

func (r *SidecarDefinitionReconciler) specHash(sidecarDef *appsv1.SidecarDefinition) (string, error) {
	data, err := json.Marshal(sidecarDef.Spec)
	if err != nil {
		return "", err
	}
	hash := fnv.New32a()
	hash.Write(data)
	return rand.SafeEncodeString(fmt.Sprintf("%d", hash.Sum32())), nil
}

func (r *SidecarDefinitionReconciler) immutableHash(cli client.Client, rctx intctrlutil.RequestCtx,
	sidecarDef *appsv1.SidecarDefinition) error {
	if r.skipImmutableCheck(sidecarDef) {
		return nil
	}

	if sidecarDef.Annotations != nil {
		_, ok := sidecarDef.Annotations[immutableHashAnnotationKey]
		if ok {
			return nil
		}
	}

	patch := client.MergeFrom(sidecarDef.DeepCopy())
	if sidecarDef.Annotations == nil {
		sidecarDef.Annotations = map[string]string{}
	}
	sidecarDef.Annotations[immutableHashAnnotationKey], _ = r.specHash(sidecarDef)
	return cli.Patch(rctx.Ctx, sidecarDef, patch)
}

func matchedSidecarDef4CompDefs(ctx context.Context, cli client.Reader, compDefs []string) (map[string][]*appsv1.SidecarDefinition, error) {
	sidecarList := &appsv1.SidecarDefinitionList{}
	if err := cli.List(ctx, sidecarList); err != nil {
		return nil, err
	}

	compDefsSet := sets.New[string](compDefs...)
	match := func(sidecarDef *appsv1.SidecarDefinition) ([]string, error) {
		owners := sets.New(strings.Split(sidecarDef.Status.Owners, ",")...)
		selectors := sets.New(strings.Split(sidecarDef.Status.Selectors, ",")...)

		owned := compDefsSet.Intersection(owners)
		if len(owned) == 0 {
			return nil, nil
		}
		hosted := compDefsSet.Intersection(selectors)
		if len(hosted) == 0 {
			return nil, nil
		}

		if hosted.Intersection(owned).Len() != 0 {
			return nil, fmt.Errorf("owner and selectors should not be overlapped: %s",
				strings.Join(sets.List(hosted.Intersection(owned)), ","))
		}
		return sets.List(hosted), nil
	}

	grouped := make(map[string]map[string]*appsv1.SidecarDefinition)
	for i, sidecarDef := range sidecarList.Items {
		defs, err := match(&sidecarDef)
		if err != nil {
			return nil, err
		}
		for _, def := range defs {
			if _, ok := grouped[def]; !ok {
				grouped[def] = map[string]*appsv1.SidecarDefinition{}
			}
			o, ok := grouped[def][sidecarDef.Spec.Name]
			if !ok || sidecarDef.Name > o.Name {
				grouped[def][sidecarDef.Spec.Name] = &sidecarList.Items[i]
			}
		}
	}

	available := func(sidecarDef *appsv1.SidecarDefinition) error {
		if sidecarDef.Generation != sidecarDef.Status.ObservedGeneration {
			return fmt.Errorf("the SidecarDefinition is not up to date: %s", sidecarDef.Name)
		}
		if sidecarDef.Status.Phase != appsv1.AvailablePhase {
			return fmt.Errorf("the SidecarDefinition is unavailable: %s", sidecarDef.Name)
		}
		return nil
	}

	result := make(map[string][]*appsv1.SidecarDefinition)
	for compDef, sidecarDefs := range grouped {
		result[compDef] = make([]*appsv1.SidecarDefinition, 0, len(sidecarDefs))
		names := maps.Keys(sidecarDefs)
		slices.Sort(names)
		for _, name := range names {
			if err := available(sidecarDefs[name]); err != nil {
				return nil, err
			}
			result[compDef] = append(result[compDef], sidecarDefs[name])
		}
	}
	return result, nil
}
