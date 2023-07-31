/*
Copyright 2023 The Tekton Authors

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

package internalparameters

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tektoncd/chains/pkg/chains/objects"
	"github.com/tektoncd/chains/pkg/internal/objectloader"
	"github.com/tektoncd/pipeline/pkg/apis/config"
)

func TestTektonInternalParameters(t *testing.T) {
	tr, err := objectloader.TaskRunFromFile("../../../testdata/v2alpha2/taskrun1.json")
	if err != nil {
		t.Fatal(err)
	}
	tro := objects.NewTaskRunObject(tr)
	got := TektonInternalParameters(tro)
	want := map[string]any{
		"labels":                         tro.GetLabels(),
		"annotations":                    tro.GetAnnotations(),
		"tekton-pipelines-feature-flags": config.FeatureFlags{EnableAPIFields: "beta", ResultExtractionMethod: "termination-message"},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("TaskRun(): -want +got: %s", diff)
	}
}

func TestSLSAInternalParameters(t *testing.T) {
	tr, err := objectloader.TaskRunFromFile("../../../testdata/v2alpha2/taskrun1.json")
	if err != nil {
		t.Fatal(err)
	}
	tro := objects.NewTaskRunObject(tr)
	got := SLSAInternalParameters(tro)
	want := map[string]any{
		"tekton-pipelines-feature-flags": config.FeatureFlags{EnableAPIFields: "beta", ResultExtractionMethod: "termination-message"},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("TaskRun(): -want +got: %s", diff)
	}
}
