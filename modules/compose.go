package modules

// BuildComposeArgs constructs standard docker compose command arguments.
// Docker Compose automatically resolves compose and standard override files in the working directory.
func BuildComposeArgs(projectName string, subCmd ...string) []string {
	args := []string{"compose"}
	if projectName != "" {
		args = append(args, "--project-name", projectName)
	}
	args = append(args, subCmd...)
	return args
}
