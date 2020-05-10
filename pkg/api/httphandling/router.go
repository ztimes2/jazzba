package httphandling

import (
	"net/http"

	"github.com/ztimes2/jazzba/pkg/api/service"

	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
)

const (
	pathParamNotebookID = "notebook_id"
	pathParamNoteID     = "note_id"
	pathParamTagName    = "tag_name"
)

const (
	regexInteger = `^[1-9]\d*$`
)

func newIntPathParamPlaceholder(name string) string {
	return "{" + name + ":" + regexInteger + "}"
}

func newStringPathParamPlaceholder(name string) string {
	return "{" + name + "}"
}

// RouterConfig holds parameters required for initializing an HTTP router.
type RouterConfig struct {
	NotebookService service.Notebooker
	NoteService     service.Noter
	NoteTagService  service.NoteTagger
	Logger          logrus.FieldLogger
}

// NewRouter initializes a new HTTP router.
func NewRouter(cfg RouterConfig) http.Handler {
	notebookHandler := newNotebookHandler(cfg.NotebookService, cfg.Logger)
	noteHandler := newNoteHandler(cfg.NoteService, cfg.Logger)
	noteTagHandler := newNoteTagHandler(cfg.NoteTagService, cfg.Logger)

	notebooksSubRouter := chi.NewRouter().Route("/notebooks", func(r chi.Router) {
		r.Post("/", notebookHandler.createNotebook)
		r.Get("/", notebookHandler.fetchNotebooks)

		r.Route("/"+newIntPathParamPlaceholder(pathParamNotebookID), func(r chi.Router) {
			r.Get("/", notebookHandler.fetchNotebook)
			r.Put("/", notebookHandler.updateNotebook)
			r.Delete("/", notebookHandler.deleteNotebook)
			r.Get("/notes", noteHandler.fetchNotesByNotebook)
		})
	})

	notesSubRouter := chi.NewRouter().Route("/notes", func(r chi.Router) {
		r.Post("/", noteHandler.createNote)
		r.Get("/", noteHandler.fetchNotes)

		r.Route("/"+newIntPathParamPlaceholder(pathParamNoteID), func(r chi.Router) {
			r.Get("/", noteHandler.fetchNote)
			r.Put("/", noteHandler.updateNote)
			r.Delete("/", noteHandler.deleteNote)
		})

		r.Route("/"+newIntPathParamPlaceholder(pathParamNoteID)+"/tags", func(r chi.Router) {
			r.Post("/", noteTagHandler.createNoteTag)
			r.Get("/", noteTagHandler.fetchNoteTagsByNote)
			r.Delete("/"+newStringPathParamPlaceholder(pathParamTagName),
				noteTagHandler.deleteNoteTag)
		})

	})

	bulkSubRouter := chi.NewRouter().Route("/bulk", func(r chi.Router) {
		r.Get("/notebook_notes", noteHandler.fetchNotesByNotebooks)
		r.Get("/note_tags", noteTagHandler.fetchNoteTagsByNotes)
	})

	routerV1 := chi.NewRouter().Route("/v1", func(r chi.Router) {
		r.Mount("/", notebooksSubRouter)
		r.Mount("/", notesSubRouter)
		r.Mount("/", bulkSubRouter)
		r.Get("/search/notes", noteHandler.fetchNotesBySearchQuery)
	})

	return routerV1
}
