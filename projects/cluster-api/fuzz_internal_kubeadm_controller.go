// Copyright 2022 ADA Logics Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package controllers

import (
	"context"
	"encoding/json"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/collections"
	"sigs.k8s.io/cluster-api/controlplane/kubeadm/internal"
	//ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
	"strings"

	fuzz "github.com/AdaLogics/go-fuzz-headers"
)

var (
	fakeGenericMachineTemplateCRD = &apiextensionsv1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiextensionsv1.SchemeGroupVersion.String(),
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "genericmachinetemplate.generic.io",
			Labels: map[string]string{
				"cluster.x-k8s.io/v1beta1": "v1",
			},
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "generic.io",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind: "GenericMachineTemplate",
			},
		},
	}
)

func FuzzKubeadmControlPlaneReconciler(f *testing.F) {
	f.Fuzz(func (t *testing.T, data []byte){
		fdp := fuzz.NewConsumer(data)
		cluster, kcp, tmpl, err := createClusterWithControlPlaneFuzz(fdp)
		if err != nil {
			return
		}
		if tmpl == nil {
			return
		}
		objs := []client.Object{fakeGenericMachineTemplateCRD, cluster.DeepCopy(), kcp.DeepCopy(), tmpl.DeepCopy()}

		m := &clusterv1.Machine{}
		err = fdp.GenerateStruct(m)
		if err != nil {
			return
		}
		cfg := &bootstrapv1.KubeadmConfig{}
		err = fdp.GenerateStruct(cfg)
		if err != nil {
			return
		}
		objs = append(objs, m, cfg)
		fmc := &fakeManagementCluster{
			Machines: collections.Machines{},
			Workload: &fakeWorkloadCluster{},
		}
		fmc.Machines.Insert(m)
		fakeClient := newFakeClient(objs...)
		fmc.Reader = fakeClient

		r := &KubeadmControlPlaneReconciler{
			Client:                    fakeClient,
			managementCluster:         fmc,
			managementClusterUncached: fmc,
		}
		controlPlane := &internal.ControlPlane{}
		err = fdp.GenerateStruct(controlPlane)
		if err != nil {
			return
		}

		_, err = r.reconcile(context.Background(), controlPlane)
	})
}

func createClusterWithControlPlaneFuzz(f *fuzz.ConsumeFuzzer) (*clusterv1.Cluster, *controlplanev1.KubeadmControlPlane, *unstructured.Unstructured, error) {
	cluster := &clusterv1.Cluster{}
	err := f.GenerateStruct(cluster)
	if err != nil {
		return nil, nil, nil, err
	}

	kcp := &controlplanev1.KubeadmControlPlane{}
	err = f.GenerateStruct(kcp)
	if err != nil {
		return nil, nil, nil, err
	}

	unstructuredStr, err := f.GetString()
	if err != nil {
		return nil, nil, nil, err
	}
	unstr, err := UnstructuredForFuzzing(unstructuredStr)
	if err != nil {
		return nil, nil, nil, err
	}

	return cluster, kcp, unstr, nil
}

func UnstructuredForFuzzing(text string) (*unstructured.Unstructured, error) {
	un := &unstructured.Unstructured{}
	var err error
	if strings.HasPrefix(text, "{") {
		err = json.Unmarshal([]byte(text), &un)
	} else {
		err = yaml.Unmarshal([]byte(text), &un)
	}
	if err != nil {
		return nil, err
	}
	return un, nil
}
