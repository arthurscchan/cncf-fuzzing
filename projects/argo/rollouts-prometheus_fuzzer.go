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

package prometheus

import (
	fuzz "github.com/AdaLogics/go-fuzz-headers"
	log "github.com/sirupsen/logrus"

	"github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
)

func FuzzPrometheusProvider(data []byte) int {
	f := fuzz.NewConsumer(data)
	metric := v1alpha1.Metric{}
	err := f.GenerateStruct(&metric)
	if err != nil {
		return 0
	}
	e := log.Entry{}
	mock := &mockAPI{
		value: newScalar(10),
	}
	p, err := NewPrometheusProvider(mock, e, metric)
	if err != nil {
		return 0
	}
	_ = p.Run(newAnalysisRun(), metric)
	return 1
}
