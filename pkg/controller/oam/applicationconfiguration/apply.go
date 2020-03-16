/*
Copyright 2020 The Crossplane Authors.

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

package applicationconfiguration

import (
	"context"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/crossplane/apis/oam/v1alpha2"
)

// Reconcile error strings.
const (
	errFmtApplyWorkload  = "cannot apply workload %q"
	errFmtSetWorkloadRef = "cannot set trait %q reference to %q"
	errFmtApplyTrait     = "cannot apply trait %q"
)

// A WorkloadApplicator creates or updates workloads and their traits.
type WorkloadApplicator interface {
	// Apply a workload and its traits.
	Apply(ctx context.Context, w []Workload) error
}

// A WorkloadApplyFn creates or updates workloads and their traits.
type WorkloadApplyFn func(ctx context.Context, w []Workload) error

// Apply a workload and its traits.
func (fn WorkloadApplyFn) Apply(ctx context.Context, w []Workload) error {
	return fn(ctx, w)
}

type workloads struct {
	client client.Client
}

func (a *workloads) Apply(ctx context.Context, w []Workload) error {
	for _, wl := range w {
		if err := resource.Apply(ctx, a.client, wl.Workload, resource.ControllersMustMatch()); err != nil {
			return errors.Wrapf(err, errFmtApplyWorkload, wl.Workload.GetName())
		}

		for i := range wl.Traits {
			t := &wl.Traits[i]

			ref := v1alpha2.WorkloadReference{
				APIVersion: wl.Workload.GetAPIVersion(),
				Kind:       wl.Workload.GetKind(),
				Name:       wl.Workload.GetName(),
				UID:        wl.Workload.GetUID(),
			}
			if err := fieldpath.Pave(t.UnstructuredContent()).SetValue("spec.workloadRef", ref); err != nil {
				return errors.Wrapf(err, errFmtSetWorkloadRef, t.GetName(), wl.Workload.GetName())
			}

			if err := resource.Apply(ctx, a.client, t, resource.ControllersMustMatch()); err != nil {
				return errors.Wrapf(err, errFmtApplyTrait, t.GetName())
			}
		}
	}

	return nil
}
