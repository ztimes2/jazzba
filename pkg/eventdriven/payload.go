package eventdriven

// NoteCreatedEventPayload represents a payload for the 'note_created' event.
type NoteCreatedEventPayload struct {
	NoteID int `json:"note_id"`
}

// NoteUpdatedEventPayload represents a payload for the 'note_updated' event.
type NoteUpdatedEventPayload struct {
	NoteID int `json:"note_id"`
}

// NoteDeletedEventPayload represents a payload for the 'note_deleted' event.
type NoteDeletedEventPayload struct {
	NoteID int `json:"note_id"`
}

// NoteTagCreatedEventPayload represents a payload for the 'note_tag_created' event.
type NoteTagCreatedEventPayload struct {
	NoteID int `json:"note_id"`
}

// NoteTagDeletedEventPayload represents a payload for the 'note_tag_deleted'
// event.
type NoteTagDeletedEventPayload struct {
	NoteID int `json:"note_id"`
}

// NotebookUpdatedEventPayload represents a payload for the 'notebook_updated'
// event.
type NotebookUpdatedEventPayload struct {
	NotebookID int `json:"notebook_id"`
}

// NotebookDeletedEventPayload represents a payload for the 'notebook_deleted'
// event.
type NotebookDeletedEventPayload struct {
	NotebookID int `json:"notebook_id"`
}
