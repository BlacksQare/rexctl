package structs

type Manifest struct {
	Kind string `yaml:"kind"`
	Spec Spec   `yaml:"spec"`
}

type Spec struct {
	Containers []Container `yaml:"containers"`
}

type Container struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Remote   string `yaml:"remote"`
	Revision string `yaml:"revision,omitempty"`
}


type ComposeFile struct {
	Services map[string]ComposeService `yaml:"services"`
}

type ComposeService struct {
	Build         any    `yaml:"build"`
	Image         string `yaml:"image"`
	ContainerName string `yaml:"container_name"`
}

type ContainerInfo struct {
	Name              string `json:"name"`
	Type              string `json:"type"`
	Remote            string `json:"remote"`
	RequestedRevision string `json:"requested_revision,omitempty"`
	CurrentRevision   string `json:"current_revision,omitempty"`
	State             string `json:"state"`
	ImageTag          string `json:"image_tag,omitempty"`
}

type WorkspaceInfo struct {
	StackName  string          `json:"stack_name"`
	IsRunning  bool            `json:"is_running"`
	Containers []ContainerInfo `json:"containers"`
}

