// types.go

package compose

// ComposeFile represents the structure of a Docker Compose file.
type ComposeFile struct {
	Version  string            `json:"version"`  // The version of the Compose file format
	Services map[string]ServiceDef `json:"services"` // The services defined in the Compose file
}

// ServiceDef defines a service in Docker Compose.
type ServiceDef struct {
	Image      string     `json:"image"`      // The Docker image for the service
	Build      BuildDef   `json:"build"`      // Build definition if using a Dockerfile
	Ports      []string   `json:"ports"`      // Ports to expose
}

// BuildDef defines the build context for a service.
type BuildDef struct {
	Context string `json:"context"` // Build context directory
	Dockerfile string `json:"dockerfile"` // Dockerfile path
}