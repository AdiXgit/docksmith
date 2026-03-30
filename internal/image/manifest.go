package image

type Manifest struct {
	Name    string       `json:"name"`
	Tag     string       `json:"tag"`
	Digest  string       `json:"digest"`
	Created string       `json:"created"`
	Layers  []LayerEntry `json:"layers"`
	Config  Config       `json:"config"`
}

type LayerEntry struct {
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
	CreatedBy string `json:"createdBy"`
}

type Config struct {
	Env        []string `json:"env,omitempty"`
	Cmd        []string `json:"cmd,omitempty"`
	WorkingDir string   `json:"workingDir,omitempty"`
}
