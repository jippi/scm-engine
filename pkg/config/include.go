package config

type Include struct {
	Project string   `yaml:"project"`
	Ref     *string  `yaml:"ref"`
	Files   []string `yaml:"files"`
}
