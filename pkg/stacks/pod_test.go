/*
Copyright 2019 The Crossplane Authors.

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

package stacks

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/crossplane/crossplane-runtime/pkg/test"
)

func TestGetRunningPod(t *testing.T) {
	type want struct {
		pod *v1.Pod
		err error
	}

	tests := []struct {
		name    string
		envvars envvars
		kube    client.Client
		want    want
	}{
		{
			name: "EmptyPodNameEnvVar",
			envvars: envvars{
				podName:      "",
				podNamespace: "foo-ns",
			},
			kube: nil,
			want: want{
				pod: nil,
				err: errors.New("cannot detect the pod name. Please provide it using the downward API in the manifest file"),
			},
		},
		{
			name: "EmptyPodNamespaceEnvVar",
			envvars: envvars{
				podName:      "foo",
				podNamespace: "",
			},
			kube: nil,
			want: want{
				pod: nil,
				err: errors.New("cannot detect the pod namespace. Please provide it using the downward API in the manifest file"),
			},
		},
		{
			name: "SimpleGet",
			envvars: envvars{
				podName:      "foo",
				podNamespace: "foo-ns",
			},
			kube: fake.NewFakeClientWithScheme(scheme.Scheme, &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "foo-ns"}}),
			want: want{
				pod: &v1.Pod{TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"}, ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "foo-ns"}},
				err: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialEnvVars := saveEnvVars()
			defer restoreEnvVars(initialEnvVars)

			os.Setenv(PodNameEnvVar, tt.envvars.podName)
			os.Setenv(PodNamespaceEnvVar, tt.envvars.podNamespace)
			got, err := GetRunningPod(context.Background(), tt.kube)

			if diff := cmp.Diff(err, tt.want.err, test.EquateErrors()); diff != "" {
				t.Errorf("GetRunningPod() want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(got, tt.want.pod); diff != "" {
				t.Errorf("GetRunningPod() got != want:\n%v", diff)
			}
		})
	}
}

func TestGetSpecContainerImage(t *testing.T) {
	type want struct {
		image string
		err   error
	}

	tests := []struct {
		name          string
		pod           *v1.Pod
		containerName string
		initContainer bool
		want          want
	}{
		{
			name: "SingleContainer",
			pod: &v1.Pod{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{Name: "foo", Image: "foo-image"},
					},
				},
			},
			containerName: "foo",
			want: want{
				image: "foo-image",
				err:   nil,
			},
		},
		{
			name: "MultipleContainers",
			pod: &v1.Pod{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{Name: "bar", Image: "bar-image"},
						{Name: "foo", Image: "foo-image"},
					},
				},
			},
			containerName: "foo",
			want: want{
				image: "foo-image",
				err:   nil,
			},
		},
		{
			name: "InitContainer",
			pod: &v1.Pod{
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						{Name: "foo", Image: "foo-image"},
					},
					Containers: []v1.Container{
						{Name: "bar", Image: "bar-image"},
					},
				},
			},
			containerName: "foo",
			initContainer: true,
			want: want{
				image: "foo-image",
				err:   nil,
			},
		},
		{
			name: "NoMatches",
			pod: &v1.Pod{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{Name: "bar", Image: "bar-image"},
						{Name: "foo", Image: "foo-image"},
					},
				},
			},
			containerName: "baz",
			want: want{
				image: "",
				err:   errors.New("failed to find image for container baz"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetContainerImage(tt.pod, tt.containerName, tt.initContainer)

			if diff := cmp.Diff(err, tt.want.err, test.EquateErrors()); diff != "" {
				t.Errorf("GetContainerImage() want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(got, tt.want.image); diff != "" {
				t.Errorf("GetContainerImage() got != want:\n%v", diff)
			}
		})
	}
}

func TestGetContainerImagePullPolicy(t *testing.T) {
	type want struct {
		imagePullPolicy v1.PullPolicy
		err             error
	}

	tests := []struct {
		name          string
		pod           *v1.Pod
		containerName string
		initContainer bool
		want          want
	}{
		{
			name: "SingleContainer",
			pod: &v1.Pod{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{Name: "foo", Image: "foo-image", ImagePullPolicy: v1.PullAlways},
					},
				},
			},
			containerName: "foo",
			want: want{
				imagePullPolicy: v1.PullAlways,
				err:             nil,
			},
		},
		{
			name: "InitContainer",
			pod: &v1.Pod{
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						{Name: "foo", Image: "foo-image", ImagePullPolicy: v1.PullAlways},
					},
					Containers: []v1.Container{
						{Name: "bar", Image: "bar-image"},
					},
				},
			},
			containerName: "foo",
			initContainer: true,
			want: want{
				imagePullPolicy: v1.PullAlways,
				err:             nil,
			},
		},
		{
			name: "NoMatches",
			pod: &v1.Pod{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{Name: "bar", Image: "bar-image"},
						{Name: "foo", Image: "foo-image"},
					},
				},
			},
			containerName: "baz",
			want: want{
				imagePullPolicy: "",
				err:             errors.New("failed to find image for container baz"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetContainerImagePullPolicy(tt.pod, tt.containerName, tt.initContainer)

			if diff := cmp.Diff(err, tt.want.err, test.EquateErrors()); diff != "" {
				t.Errorf("GetContainerImagePullPolicy() want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(got, tt.want.imagePullPolicy); diff != "" {
				t.Errorf("GetContainerImagePullPolicy() got != want:\n%v", diff)
			}
		})
	}
}
