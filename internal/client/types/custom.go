package types

type CodeLocationsAsDocumentResponse struct {
	Locations []CodeLocation `json:"locations"`
}

type CodeLocation struct {
	Name             string                 `json:"location_name"`
	Image            string                 `json:"image,omitempty"`
	CodeSource       CodeLocationCodeSource `json:"code_source"`
	WorkingDirectory string                 `json:"working_directory,omitempty"`
	ExecutablePath   string                 `json:"executable_path,omitempty"`
	Attribute        string                 `json:"attribute,omitempty"`
	Git              CodeLocationGit        `json:"git,omitempty"`
	AgentQueue       string                 `json:"agent_queue,omitempty"`
}

type CodeLocationCodeSource struct {
	ModuleName  string `json:"module_name,omitempty"`
	PackageName string `json:"package_name,omitempty"`
	PythonFile  string `json:"python_file,omitempty"`
}

type CodeLocationGit struct {
	CommitHash string `json:"commit_hash"`
	URL        string `json:"url"`
}
