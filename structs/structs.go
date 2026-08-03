package structs

type Manifest struct {
	Kind string `yaml:"kind"`
	Spec Spec   `yaml:"spec"`
}

type Spec struct {
	Containers []Container `yaml:"containers"`
}

type Container struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Remote      string `yaml:"remote"`
	Revision    string `yaml:"revision,omitempty"`
	ComposeFile string `yaml:"composeFile,omitempty"`
}

type ComposeFile struct {
	Services map[string]ComposeService `yaml:"services"`
}

type ComposeService struct {
	Build         any    `yaml:"build"`
	Image         string `yaml:"image"`
	ContainerName string `yaml:"container_name"`
}
