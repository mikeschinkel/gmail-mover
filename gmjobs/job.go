package gmjobs

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

// Job represents a Gmail operation job with parsed spec
type Job struct {
	Version string
	JobType string
	Name    string
	Spec    JobSpec
}

// jobFile is the internal JSON structure for loading
type jobFile struct {
	Version string          `json:"version"`
	JobType string          `json:"job_type"`
	Name    string          `json:"name"`
	Spec    json.RawMessage `json:"spec"`
}

// LoadJobFile loads and parses a job file
func LoadJobFile(filename string) (job *Job, err error) {
	var data []byte
	var jf jobFile
	var specType reflect.Type
	var exists bool
	var spec JobSpec

	data, err = os.ReadFile(filename)
	if err != nil {
		goto end
	}

	err = json.Unmarshal(data, &jf)
	if err != nil {
		goto end
	}

	if jf.Version == "" {
		err = fmt.Errorf("job version is required")
		goto end
	}

	if jf.JobType == "" {
		err = fmt.Errorf("job_type is required")
		goto end
	}

	specType, exists = jobSpecs[jf.JobType]
	if !exists {
		err = fmt.Errorf("unknown job type: %s", jf.JobType)
		goto end
	}

	spec = reflect.New(specType).Interface().(JobSpec)
	err = json.Unmarshal(jf.Spec, spec)
	if err != nil {
		goto end
	}

	// Validation happens during ToConfig() call

	job = &Job{
		Version: jf.Version,
		JobType: jf.JobType,
		Name:    jf.Name,
		Spec:    spec,
	}

end:
	return job, err
}

// SaveJobFile saves a job spec to a file
func SaveJobFile(filename JobFile, spec JobSpec) (err error) {
	var specData []byte
	var jobData []byte
	var jf jobFile

	// Validation happens during ToConfig() call

	specData, err = json.MarshalIndent(spec, "", "  ")
	if err != nil {
		goto end
	}

	jf = jobFile{
		Version: "1.0",
		JobType: spec.JobType(),
		Name:    spec.Name(),
		Spec:    json.RawMessage(specData),
	}

	jobData, err = json.MarshalIndent(jf, "", "  ")
	if err != nil {
		goto end
	}

	err = os.WriteFile(string(filename), jobData, 0644)

end:
	return err
}

// ToConfig converts the job to gmover config
func (j *Job) ToConfig() (Config, error) {
	return j.Spec.ToConfig()
}
