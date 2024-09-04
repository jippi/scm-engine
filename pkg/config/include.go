package config

type Include struct {
	// The project to include files from
	//
	// See: https://jippi.github.io/scm-engine/configuration/#include.project
	Project string `json:"project" yaml:"project"`

	// The list of files to include from the project. The paths must be relative to the repository root, e.x. label/some-config-file.yml; NOT /label/some-config-file.yml
	//
	// See: https://jippi.github.io/scm-engine/configuration/#include.files
	Files []string `json:"files" yaml:"files"`

	// (Optional) Git reference to read the configuration from; it can be a tag, branch, or commit SHA.
	//
	// If omitted, HEAD is used; meaning your default branch.
	//
	// See: https://jippi.github.io/scm-engine/configuration/#include.ref
	Ref *string `json:"ref,omitempty" yaml:"ref"`
}
