package builddefinitions

import (
	"fmt"

	"github.com/tektoncd/chains/pkg/chains/objects"
)

type TektonTaskBuildType struct {
	BuildType string
	Tro       *objects.TaskRunObject
}

type SLSATaskBuildType struct {
	BuildType string
	Tro       *objects.TaskRunObject
}

// internalParameters adds the tekton feature flags that were enabled
// for the taskrun.
func (s SLSATaskBuildType) InternalParameters() map[string]any {
	internalParams := make(map[string]any)
	if s.Tro.Status.Provenance != nil && s.Tro.Status.Provenance.FeatureFlags != nil {
		internalParams["tekton-pipelines-feature-flags"] = *s.Tro.Status.Provenance.FeatureFlags
	}
	return internalParams
}

// externalParameters adds the task run spec
func (s SLSATaskBuildType) ExternalParameters() map[string]any {
	externalParams := make(map[string]any)
	// add origin of the top level task config
	// isRemoteTask checks if the task was fetched using a remote resolver
	isRemoteTask := false
	if s.Tro.Spec.TaskRef != nil {
		if s.Tro.Spec.TaskRef.Resolver != "" && s.Tro.Spec.TaskRef.Resolver != "Cluster" {
			isRemoteTask = true
		}
	}
	if t := s.Tro.Status.Provenance; t != nil && t.RefSource != nil && isRemoteTask {
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
	externalParams["runSpec"] = s.Tro.Spec
	return externalParams
}

func (t TektonTaskBuildType) InternalParameters() map[string]any {
	return make(map[string]any)
}

func (t TektonTaskBuildType) ExternalParameters() map[string]any {
	externalParams := make(map[string]any)
	return externalParams
}
