package httphandling

type createNotebookRequestBody struct {
	Name string `json:"name"`
}

type updateNotebookRequestBody struct {
	Name string `json:"name"`
}

type createNoteRequestBody struct {
	Name       string `json:"name"`
	Content    string `json:"content"`
	NotebookID int    `json:"notebook_id"`
}

type updateNoteRequestBody struct {
	Name       string `json:"name"`
	Content    string `json:"content"`
	NotebookID int    `json:"notebook_id"`
}

type createNoteTagRequestBody struct {
	TagName string `json:"tag_name"`
}
