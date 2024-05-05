package types

import "github.com/datarootsio/terraform-provider-dagster/internal/utils"

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

func (cl CodeLocation) Validate() error {
	if cl.Name == "" {
		return &ErrInvalid{What: "CodeLocation", Message: "Name is required"}
	}

	if cl.Git == (CodeLocationGit{}) && cl.Image == "" {
		return &ErrInvalid{What: "CodeLocation", Message: "Git or Image is required"}
	}

	if cl.Git != (CodeLocationGit{}) && cl.Image != "" {
		return &ErrInvalid{What: "CodeLocation", Message: "You can only specify one of Git or Image"}
	}

	if cl.Git != (CodeLocationGit{}) && (cl.Git.CommitHash == "" || cl.Git.URL == "") {
		return &ErrInvalid{What: "CodeLocation", Message: "Must specify fields Git.CommitHash and Git.URL"}
	}

	if cl.CodeSource.ModuleName == "" && cl.CodeSource.PackageName == "" && cl.CodeSource.PythonFile == "" {
		return &ErrInvalid{What: "CodeLocation", Message: "CodeSource.ModuleName or CodeSource.PackageName or CodeSource.PythonFile is required"}
	}

	if !utils.AreMutuallyExclusive(cl.CodeSource.ModuleName, cl.CodeSource.PackageName, cl.CodeSource.PythonFile) {
		return &ErrInvalid{What: "CodeLocation", Message: "CodeSource.ModuleName/CodeSource.PackageName/CodeSource.PythonFile should be mutually exclusive"}
	}

	return nil
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
