package types

type CodeLocationsAsDocumentResponse struct {
	Locations []CodeLocation `json:"locations"`
}

type CodeLocation struct {
	Name       string                 `json:"location_name"`
	Image      string                 `json:"image"`
	CodeSource CodeLocationCodeSource `json:"code_source"`
}

type CodeLocationCodeSource struct {
	PythonFile string `json:"python_file"`
}
