// /*
// Copyright 2023 The Tekton Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */

package taskrun

import (
	"context"
	"encoding/json"
	"fmt"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
	slsa "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v1"
	v1 "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v1"
	"github.com/tektoncd/chains/pkg/chains/formats/slsa/extract"
	builddefinitions "github.com/tektoncd/chains/pkg/chains/formats/slsa/v2alpha2/internal/build_definitions"
	"github.com/tektoncd/chains/pkg/chains/formats/slsa/v2alpha2/internal/pipelinerun"
	"github.com/tektoncd/chains/pkg/chains/objects"
)

const taskRunResults = "taskRunResults/%s"

type BuildDefintion interface {
	ExternalParameters() map[string]any
	InternalParameters() map[string]any
	ResolvedDependencies(context.Context) ([]v1.ResourceDescriptor, error)
}

// GenerateAttestation generates a provenance statement with SLSA v1.0 predicate for a task run.
func GenerateAttestation(ctx context.Context, builderID string, tro *objects.TaskRunObject, buildType string) (interface{}, error) {
	bp, err := byproducts(tro)
	if err != nil {
		return nil, err
	}

	bd, err := getBuildDefinition(buildType, tro)
	if err != nil {
		return nil, err
	}
	rd, err := bd.ResolvedDependencies(ctx)
	if err != nil {
		return nil, err
	}

	att := intoto.ProvenanceStatementSLSA1{
		StatementHeader: intoto.StatementHeader{
			Type:          intoto.StatementInTotoV01,
			PredicateType: slsa.PredicateSLSAProvenance,
			Subject:       extract.SubjectDigests(ctx, tro),
		},
		Predicate: slsa.ProvenancePredicate{
			BuildDefinition: slsa.ProvenanceBuildDefinition{
				BuildType:            buildType,
				ExternalParameters:   bd.ExternalParameters(),
				InternalParameters:   bd.InternalParameters(),
				ResolvedDependencies: rd,
			},
			RunDetails: slsa.ProvenanceRunDetails{
				Builder: slsa.Builder{
					ID: builderID,
				},
				BuildMetadata: metadata(tro),
				Byproducts:    bp,
			},
		},
	}
	return att, nil
}

func metadata(tro *objects.TaskRunObject) slsa.BuildMetadata {
	m := slsa.BuildMetadata{
		InvocationID: string(tro.ObjectMeta.UID),
	}
	if tro.Status.StartTime != nil {
		utc := tro.Status.StartTime.Time.UTC()
		m.StartedOn = &utc
	}
	if tro.Status.CompletionTime != nil {
		utc := tro.Status.CompletionTime.Time.UTC()
		m.FinishedOn = &utc
	}
	return m
}

// internalParameters adds the tekton feature flags that were enabled
// for the taskrun.
func internalParameters(tro *objects.TaskRunObject) map[string]any {
	internalParams := make(map[string]any)
	if tro.Status.Provenance != nil && tro.Status.Provenance.FeatureFlags != nil {
		internalParams["tekton-pipelines-feature-flags"] = *tro.Status.Provenance.FeatureFlags
	}
	return internalParams
}

// externalParameters adds the task run spec
func externalParameters(tro *objects.TaskRunObject) map[string]any {
	externalParams := make(map[string]any)
	// add origin of the top level task config
	// isRemoteTask checks if the task was fetched using a remote resolver
	isRemoteTask := false
	if tro.Spec.TaskRef != nil {
		if tro.Spec.TaskRef.Resolver != "" && tro.Spec.TaskRef.Resolver != "Cluster" {
			isRemoteTask = true
		}
	}
	if t := tro.Status.Provenance; t != nil && t.RefSource != nil && isRemoteTask {
		ref := ""
		for alg, hex := range t.RefSource.Digest {
			ref = fmt.Sprintf("%s:%s", alg, hex)
			break
		}
		buildConfigSource := map[string]string{
			"ref":        ref,
			"repository": t.RefSource.URI,
			"path":       t.RefSource.EntryPoint,
		}
		externalParams["buildConfigSource"] = buildConfigSource
	}
	externalParams["runSpec"] = tro.Spec
	return externalParams
}

// byproducts contains the taskRunResults
func byproducts(tro *objects.TaskRunObject) ([]slsa.ResourceDescriptor, error) {
	byProd := []slsa.ResourceDescriptor{}
	for _, key := range tro.Status.TaskRunResults {
		content, err := json.Marshal(key.Value)
		if err != nil {
			return nil, err
		}
		bp := slsa.ResourceDescriptor{
			Name:      fmt.Sprintf(taskRunResults, key.Name),
			Content:   content,
			MediaType: pipelinerun.JsonMediaType,
		}
		byProd = append(byProd, bp)
	}
	return byProd, nil
}

func getBuildDefinition(buildType string, tro *objects.TaskRunObject) (BuildDefintion, error) {
	switch buildType {
	case "https://tekton.dev/chains/v2/slsa":
		return builddefinitions.SLSATaskBuildType{
			BuildType: buildType,
			Tro:       tro,
		}, nil
	case "tekton-build-type":
		return builddefinitions.TektonTaskBuildType{
			BuildType: buildType,
			Tro:       tro,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported buildType %v", buildType)
	}
}
