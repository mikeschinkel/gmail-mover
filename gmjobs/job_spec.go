package gmjobs

import (
	"reflect"
)

// JobSpec interface for different job types
// Note: Implementations live in gmover package where domain expertise resides
type JobSpec interface {
	JobType() string
	Name() string
	ToConfig() (Config, error)
}

// Registry for job spec types
var jobSpecs = make(map[string]reflect.Type)

// RegisterJobSpec registers a job spec type for unmarshaling
// Called from gmover package init() functions
func RegisterJobSpec(spec JobSpec) {
	jobSpecs[spec.JobType()] = reflect.TypeOf(spec).Elem()
}
