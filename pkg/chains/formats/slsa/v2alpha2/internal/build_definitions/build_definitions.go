package builddefinitions

import (
	"context"
	"fmt"

	slsa "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v1"
	resolveddependencies "github.com/tektoncd/chains/pkg/chains/formats/slsa/v2alpha2/internal/resolved_dependencies"
	"github.com/tektoncd/chains/pkg/chains/objects"
)

type TektonBuildDefinition struct {
	BuildType string
}

type GenericBuildDefinition struct {
	BuildType string
}

func (t TektonBuildDefinition) GenerateBuildDefinition(ctx context.Context, pro *objects.PipelineRunObject) (slsa.ProvenanceBuildDefinition, error) {
	rd, err := resolveddependencies.PipelineRun(ctx, pro)
	if err != nil {
		return slsa.ProvenanceBuildDefinition{}, err
	}

	return slsa.ProvenanceBuildDefinition{
		BuildType:            t.BuildType,
		ExternalParameters:   externalParameters(pro),
		InternalParameters:   internalParameters(pro),
		ResolvedDependencies: rd,
	}, nil
}

func (g GenericBuildDefinition) GenerateBuildDefinition(ctx context.Context, pro *objects.PipelineRunObject) (slsa.ProvenanceBuildDefinition, error) {
	rd, err := resolveddependencies.PipelineRun(ctx, pro)
	if err != nil {
		return slsa.ProvenanceBuildDefinition{}, err
	}

	return slsa.ProvenanceBuildDefinition{
		BuildType:            g.BuildType,
		ExternalParameters:   externalParameters(pro),
		InternalParameters:   internalParameters(pro),
		ResolvedDependencies: rd,
	}, nil
}

// internalParameters adds the tekton feature flags that were enabled
// for the pipelinerun.
func internalParameters(pro *objects.PipelineRunObject) map[string]any {
	internalParams := make(map[string]any)
	if pro.Status.Provenance != nil && pro.Status.Provenance.FeatureFlags != nil {
		internalParams["tekton-pipelines-feature-flags"] = *pro.Status.Provenance.FeatureFlags
	}
	return internalParams
}

// externalParameters adds the pipeline run spec
func externalParameters(pro *objects.PipelineRunObject) map[string]any {
	externalParams := make(map[string]any)

	// add the origin of top level pipeline config
	// isRemotePipeline checks if the pipeline was fetched using a remote resolver
	isRemotePipeline := false
	if pro.Spec.PipelineRef != nil {
		if pro.Spec.PipelineRef.Resolver != "" && pro.Spec.PipelineRef.Resolver != "Cluster" {
			isRemotePipeline = true
		}
	}

	if p := pro.Status.Provenance; p != nil && p.RefSource != nil && isRemotePipeline {
		ref := ""
		for alg, hex := range p.RefSource.Digest {
			ref = fmt.Sprintf("%s:%s", alg, hex)
			break
		}
		buildConfigSource := map[string]string{
			"ref":        ref,
			"repository": p.RefSource.URI,
			"path":       p.RefSource.EntryPoint,
		}
		externalParams["buildConfigSource"] = buildConfigSource
	}
	externalParams["runSpec"] = pro.Spec
	return externalParams
}
