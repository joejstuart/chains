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

package pipelinerun

import (
	"context"
	"encoding/json"
	"fmt"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
	slsa "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v1"
	"github.com/tektoncd/chains/pkg/chains/formats/slsa/extract"
	"github.com/tektoncd/chains/pkg/chains/objects"
)

const (
	pipelineRunResults = "pipelineRunResults/%s"
	// JsonMediaType is the media type of json encoded content used in resource descriptors
	JsonMediaType = "application/json"
)

type BuildDefintion interface {
	GenerateBuildDefinition(context.Context, *objects.PipelineRunObject) (slsa.ProvenanceBuildDefinition, error)
}

// GenerateAttestation generates a provenance statement with SLSA v1.0 predicate for a pipeline run.
func GenerateAttestation(ctx context.Context, builderID string, pro *objects.PipelineRunObject, bd BuildDefintion) (interface{}, error) {
	bp, err := byproducts(pro)
	if err != nil {
		return nil, err
	}
	generatedBd, err := bd.GenerateBuildDefinition(ctx, pro)
	if err != nil {
		return nil, err
	}
	att := intoto.ProvenanceStatementSLSA1{
		StatementHeader: intoto.StatementHeader{
			Type:          intoto.StatementInTotoV01,
			PredicateType: slsa.PredicateSLSAProvenance,
			Subject:       extract.SubjectDigests(ctx, pro),
		},
		Predicate: slsa.ProvenancePredicate{
			BuildDefinition: generatedBd,
			RunDetails: slsa.ProvenanceRunDetails{
				Builder: slsa.Builder{
					ID: builderID,
				},
				BuildMetadata: metadata(pro),
				Byproducts:    bp,
			},
		},
	}
	return att, nil
}

func metadata(pro *objects.PipelineRunObject) slsa.BuildMetadata {
	m := slsa.BuildMetadata{
		InvocationID: string(pro.ObjectMeta.UID),
	}
	if pro.Status.StartTime != nil {
		utc := pro.Status.StartTime.Time.UTC()
		m.StartedOn = &utc
	}
	if pro.Status.CompletionTime != nil {
		utc := pro.Status.CompletionTime.Time.UTC()
		m.FinishedOn = &utc
	}
	return m
}

// byproducts contains the pipelineRunResults
func byproducts(pro *objects.PipelineRunObject) ([]slsa.ResourceDescriptor, error) {
	byProd := []slsa.ResourceDescriptor{}
	for _, key := range pro.Status.PipelineResults {
		content, err := json.Marshal(key.Value)
		if err != nil {
			return nil, err
		}
		bp := slsa.ResourceDescriptor{
			Name:      fmt.Sprintf(pipelineRunResults, key.Name),
			Content:   content,
			MediaType: JsonMediaType,
		}
		byProd = append(byProd, bp)
	}
	return byProd, nil
}
