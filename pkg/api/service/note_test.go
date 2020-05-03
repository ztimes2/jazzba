package service

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"testing"

	"github.com/ztimes2/jazzba/pkg/api/p8n"
	"github.com/ztimes2/jazzba/pkg/eventdriven"
	"github.com/ztimes2/jazzba/pkg/mock"
	"github.com/ztimes2/jazzba/pkg/search"
	"github.com/ztimes2/jazzba/pkg/storage"

	validator "github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var _ Noter = (*NoteService)(nil)

type noteValidationTestData struct {
	Name struct {
		Valid              string `json:"valid"`
		ExceedingCharLimit string `json:"exceeding_character_limit"`
	} `json:"name"`
	Content struct {
		Valid              string `json:"valid"`
		ExceedingCharLimit string `json:"exceeding_character_limit"`
	} `json:"content"`
}

const (
	filePathNoteValidationTestData = "testdata/note_validation.json"
)

func fetchNoteValidationTestData(t *testing.T) noteValidationTestData {
	rawJSON, err := ioutil.ReadFile(filePathNoteValidationTestData)
	if err != nil {
		t.Fatal(err)
	}

	var testData noteValidationTestData
	if err := json.Unmarshal(rawJSON, &testData); err != nil {
		t.Fatal(err)
	}

	return testData
}

func TestNoteService_CreateNote(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	validator := validator.New()

	testData := fetchNoteValidationTestData(t)

	tests := []struct {
		description                string
		noteService                *NoteService
		params                     CreateNoteParameters
		expectedNote               *storage.Note
		expectedErrorAssertionFunc assert.ErrorAssertionFunc
	}{
		{
			description: "returns error because name parameter is empty",
			noteService: NewNoteService(
				mock.NewNoteStore(mockController),
				mock.NewNoteSearcher(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			params: CreateNoteParameters{
				Name:       "",
				Content:    testData.Content.Valid,
				NotebookID: 1,
			},
			expectedNote:               nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns error because name parameter exceeds character length limit",
			noteService: NewNoteService(
				mock.NewNoteStore(mockController),
				mock.NewNoteSearcher(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			params: CreateNoteParameters{
				Name:       testData.Name.ExceedingCharLimit,
				Content:    testData.Content.Valid,
				NotebookID: 1,
			},
			expectedNote:               nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns error because content parameter is empty",
			noteService: NewNoteService(
				mock.NewNoteStore(mockController),
				mock.NewNoteSearcher(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			params: CreateNoteParameters{
				Name:       testData.Name.Valid,
				Content:    "",
				NotebookID: 1,
			},
			expectedNote:               nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns error because content parameter exceeds character length limit",
			noteService: NewNoteService(
				mock.NewNoteStore(mockController),
				mock.NewNoteSearcher(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			params: CreateNoteParameters{
				Name:       testData.Name.Valid,
				Content:    testData.Content.ExceedingCharLimit,
				NotebookID: 1,
			},
			expectedNote:               nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns error because note store transaction could not be started",
			noteService: NewNoteService(
				func() storage.NoteStore {
					mockNoteStore := mock.NewNoteStore(mockController)
					mockNoteStore.
						EXPECT().
						BeginTx().
						Return(nil, errors.New("something went wrong"))
					return mockNoteStore
				}(),
				mock.NewNoteSearcher(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			params: CreateNoteParameters{
				Name:       testData.Name.Valid,
				Content:    testData.Content.Valid,
				NotebookID: 1,
			},
			expectedNote:               nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns error because note store's failure and rollbacks transaction",
			noteService: NewNoteService(
				func() storage.NoteStore {
					mockTx := mock.NewTx(mockController)
					mockTx.EXPECT().Rollback().Times(1)

					mockNoteStore := mock.NewNoteStore(mockController)
					mockNoteStore.
						EXPECT().BeginTx().Return(mockTx, nil)
					mockNoteStore.
						EXPECT().
						CreateOne(storage.CreateNoteParameters{
							Transaction: mockTx,
							Name:        testData.Name.Valid,
							Content:     testData.Content.Valid,
							NotebookID:  1,
						}).
						Return(nil, errors.New("something went wrong"))
					return mockNoteStore
				}(),
				mock.NewNoteSearcher(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			params: CreateNoteParameters{
				Name:       testData.Name.Valid,
				Content:    testData.Content.Valid,
				NotebookID: 1,
			},
			expectedNote:               nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns created note and commits transaction without error, but fails to produce event",
			noteService: NewNoteService(
				func() storage.NoteStore {
					mockTx := mock.NewTx(mockController)
					mockTx.EXPECT().Commit().Times(1)

					mockNoteStore := mock.NewNoteStore(mockController)
					mockNoteStore.
						EXPECT().BeginTx().Return(mockTx, nil)
					mockNoteStore.
						EXPECT().
						CreateOne(storage.CreateNoteParameters{
							Transaction: mockTx,
							Name:        testData.Name.Valid,
							Content:     testData.Content.Valid,
							NotebookID:  1,
						}).
						Return(
							&storage.Note{
								ID:         1,
								Name:       testData.Name.Valid,
								Content:    testData.Content.Valid,
								NotebookID: 1,
							},
							nil,
						)
					return mockNoteStore
				}(),
				mock.NewNoteSearcher(mockController),
				validator,
				func() eventdriven.Producer {
					mockEventProducer := mock.NewEventProducer(mockController)
					mockEventProducer.
						EXPECT().
						Produce(
							eventdriven.EventTypeNoteCreated,
							eventdriven.NoteCreatedEventPayload{
								NoteID: 1,
							},
						).
						Return(errors.New("something went wrong"))
					return mockEventProducer
				}(),
				func() logrus.FieldLogger {
					mockLogger := mock.NewLogger(mockController)
					mockLogger.
						EXPECT().
						Error(gomock.Any()).
						Times(1)
					return mockLogger
				}(),
			),
			params: CreateNoteParameters{
				Name:       testData.Name.Valid,
				Content:    testData.Content.Valid,
				NotebookID: 1,
			},
			expectedNote: &storage.Note{
				ID:         1,
				Name:       testData.Name.Valid,
				Content:    testData.Content.Valid,
				NotebookID: 1,
			},
			expectedErrorAssertionFunc: assert.NoError,
		},
		{
			description: "returns created note and commits transaction without error, and produces event without error",
			noteService: NewNoteService(
				func() storage.NoteStore {
					mockTx := mock.NewTx(mockController)
					mockTx.EXPECT().Commit().Times(1)

					mockNoteStore := mock.NewNoteStore(mockController)
					mockNoteStore.
						EXPECT().BeginTx().Return(mockTx, nil)
					mockNoteStore.
						EXPECT().
						CreateOne(storage.CreateNoteParameters{
							Transaction: mockTx,
							Name:        testData.Name.Valid,
							Content:     testData.Content.Valid,
							NotebookID:  1,
						}).
						Return(
							&storage.Note{
								ID:         1,
								Name:       testData.Name.Valid,
								Content:    testData.Content.Valid,
								NotebookID: 1,
							},
							nil,
						)
					return mockNoteStore
				}(),
				mock.NewNoteSearcher(mockController),
				validator,
				func() eventdriven.Producer {
					mockEventProducer := mock.NewEventProducer(mockController)
					mockEventProducer.
						EXPECT().
						Produce(
							eventdriven.EventTypeNoteCreated,
							eventdriven.NoteCreatedEventPayload{
								NoteID: 1,
							},
						).
						Return(nil)
					return mockEventProducer
				}(),
				mock.NewLogger(mockController),
			),
			params: CreateNoteParameters{
				Name:       testData.Name.Valid,
				Content:    testData.Content.Valid,
				NotebookID: 1,
			},
			expectedNote: &storage.Note{
				ID:         1,
				Name:       testData.Name.Valid,
				Content:    testData.Content.Valid,
				NotebookID: 1,
			},
			expectedErrorAssertionFunc: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actualNote, err := test.noteService.CreateNote(test.params)
			assert.Equal(t, test.expectedNote, actualNote)
			test.expectedErrorAssertionFunc(t, err)
		})
	}
}

func TestNoteService_FetchNote(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	validator := validator.New()

	tests := []struct {
		description                string
		noteService                *NoteService
		noteID                     int
		expectedNote               *storage.Note
		expectedErrorAssertionFunc assert.ErrorAssertionFunc
	}{
		{
			description: "returns error because of note store's failure",
			noteService: NewNoteService(
				func() storage.NoteStore {
					mockNoteStore := mock.NewNoteStore(mockController)
					mockNoteStore.
						EXPECT().
						FetchOne(1).
						Return(nil, errors.New("something went wrong"))
					return mockNoteStore
				}(),
				mock.NewNoteSearcher(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			noteID:                     1,
			expectedNote:               nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns note without error",
			noteService: NewNoteService(
				func() storage.NoteStore {
					mockNoteStore := mock.NewNoteStore(mockController)
					mockNoteStore.
						EXPECT().
						FetchOne(1).
						Return(
							&storage.Note{
								ID:      1,
								Name:    "Test name",
								Content: "Test content",
							},
							nil,
						)
					return mockNoteStore
				}(),
				mock.NewNoteSearcher(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			noteID: 1,
			expectedNote: &storage.Note{
				ID:      1,
				Name:    "Test name",
				Content: "Test content",
			},
			expectedErrorAssertionFunc: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actualNote, err := test.noteService.FetchNote(test.noteID)
			assert.Equal(t, test.expectedNote, actualNote)
			test.expectedErrorAssertionFunc(t, err)
		})
	}
}

func TestNoteService_FetchNotes(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	validator := validator.New()

	tests := []struct {
		description                string
		noteService                *NoteService
		page                       p8n.Page
		expectedPaginatedNotes     *PaginatedNotes
		expectedErrorAssertionFunc assert.ErrorAssertionFunc
	}{
		{
			description: "returns error because of note store's failure",
			noteService: NewNoteService(
				func() storage.NoteStore {
					mockNoteStore := mock.NewNoteStore(mockController)
					mockNoteStore.
						EXPECT().
						FetchAllPaginated(gomock.Eq(10), gomock.Eq(0)).
						Return(nil, errors.New("something went wrong"))
					return mockNoteStore
				}(),
				mock.NewNoteSearcher(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			page:                       p8n.NewPage(10, 0),
			expectedPaginatedNotes:     nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns notes without error",
			noteService: NewNoteService(
				func() storage.NoteStore {
					mockNoteStore := mock.NewNoteStore(mockController)
					mockNoteStore.
						EXPECT().
						FetchAllPaginated(gomock.Eq(10), gomock.Eq(0)).
						Return(
							[]storage.Note{
								storage.Note{
									ID:      1,
									Name:    "Test name",
									Content: "Test content",
								},
								storage.Note{
									ID:      2,
									Name:    "Test name",
									Content: "Test content",
								},
								storage.Note{
									ID:      3,
									Name:    "Test name",
									Content: "Test content",
								},
							},
							nil,
						)
					return mockNoteStore
				}(),
				mock.NewNoteSearcher(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			page: p8n.NewPage(10, 0),
			expectedPaginatedNotes: &PaginatedNotes{
				Notes: []storage.Note{
					storage.Note{
						ID:      1,
						Name:    "Test name",
						Content: "Test content",
					},
					storage.Note{
						ID:      2,
						Name:    "Test name",
						Content: "Test content",
					},
					storage.Note{
						ID:      3,
						Name:    "Test name",
						Content: "Test content",
					},
				},
				Pagination: p8n.NewPagination(3, p8n.NewPage(10, 0)),
			},
			expectedErrorAssertionFunc: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actualPaginatedNotes, err := test.noteService.FetchNotes(test.page)
			assert.Equal(t, test.expectedPaginatedNotes, actualPaginatedNotes)
			test.expectedErrorAssertionFunc(t, err)
		})
	}
}

func TestNoteService_FetchNotesByNotebook(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	validator := validator.New()

	tests := []struct {
		description                string
		noteService                *NoteService
		notebookID                 int
		page                       p8n.Page
		expectedPaginatedNotes     *PaginatedNotes
		expectedErrorAssertionFunc assert.ErrorAssertionFunc
	}{
		{
			description: "returns error because of note store's failure",
			noteService: NewNoteService(
				func() storage.NoteStore {
					mockNoteStore := mock.NewNoteStore(mockController)
					mockNoteStore.
						EXPECT().
						FetchManyByNotebookPaginated(
							gomock.Eq(1), gomock.Eq(10), gomock.Eq(0)).
						Return(nil, errors.New("something went wrong"))
					return mockNoteStore
				}(),
				mock.NewNoteSearcher(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			notebookID:                 1,
			page:                       p8n.NewPage(10, 0),
			expectedPaginatedNotes:     nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns notes without error",
			noteService: NewNoteService(
				func() storage.NoteStore {
					mockNoteStore := mock.NewNoteStore(mockController)
					mockNoteStore.
						EXPECT().
						FetchManyByNotebookPaginated(
							gomock.Eq(1), gomock.Eq(10), gomock.Eq(0)).
						Return(
							[]storage.Note{
								storage.Note{
									ID:         1,
									Name:       "Test name",
									Content:    "Test content",
									NotebookID: 1,
								},
								storage.Note{
									ID:         2,
									Name:       "Test name",
									Content:    "Test content",
									NotebookID: 1,
								},
								storage.Note{
									ID:         3,
									Name:       "Test name",
									Content:    "Test content",
									NotebookID: 1,
								},
							},
							nil,
						)
					return mockNoteStore
				}(),
				mock.NewNoteSearcher(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			notebookID: 1,
			page:       p8n.NewPage(10, 0),
			expectedPaginatedNotes: &PaginatedNotes{
				Notes: []storage.Note{
					storage.Note{
						ID:         1,
						Name:       "Test name",
						Content:    "Test content",
						NotebookID: 1,
					},
					storage.Note{
						ID:         2,
						Name:       "Test name",
						Content:    "Test content",
						NotebookID: 1,
					},
					storage.Note{
						ID:         3,
						Name:       "Test name",
						Content:    "Test content",
						NotebookID: 1,
					},
				},
				Pagination: p8n.NewPagination(3, p8n.NewPage(10, 0)),
			},
			expectedErrorAssertionFunc: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actualPaginatedNotes, err := test.noteService.FetchNotesByNotebook(
				test.notebookID, test.page)
			assert.Equal(t, test.expectedPaginatedNotes, actualPaginatedNotes)
			test.expectedErrorAssertionFunc(t, err)
		})
	}
}

func TestNoteService_FetchNotesByNotebooks(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	validator := validator.New()

	tests := []struct {
		description                string
		noteService                *NoteService
		notebookIDs                []int
		expectedNotebookNotesMap   storage.NotebookNotesMap
		expectedErrorAssertionFunc assert.ErrorAssertionFunc
	}{
		{
			description: "returns error because of note store's failure",
			noteService: NewNoteService(
				func() storage.NoteStore {
					mockNoteStore := mock.NewNoteStore(mockController)
					mockNoteStore.
						EXPECT().
						FetchManyByNotebooks([]int{1, 2, 3}).
						Return(nil, errors.New("something went wrong"))
					return mockNoteStore
				}(),
				mock.NewNoteSearcher(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			notebookIDs:                []int{1, 2, 3},
			expectedNotebookNotesMap:   nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns notebook notes map without error",
			noteService: NewNoteService(
				func() storage.NoteStore {
					mockNoteStore := mock.NewNoteStore(mockController)
					mockNoteStore.
						EXPECT().
						FetchManyByNotebooks([]int{1, 2, 3}).
						Return(
							storage.NotebookNotesMap{
								1: []storage.Note{
									storage.Note{
										ID:         1,
										Name:       "Test name",
										Content:    "Test content",
										NotebookID: 1,
									},
									storage.Note{
										ID:         2,
										Name:       "Test name",
										Content:    "Test content",
										NotebookID: 1,
									},
									storage.Note{
										ID:         3,
										Name:       "Test name",
										Content:    "Test content",
										NotebookID: 1,
									},
								},
								2: []storage.Note{
									storage.Note{
										ID:         1,
										Name:       "Test name",
										Content:    "Test content",
										NotebookID: 2,
									},
									storage.Note{
										ID:         2,
										Name:       "Test name",
										Content:    "Test content",
										NotebookID: 2,
									},
									storage.Note{
										ID:         3,
										Name:       "Test name",
										Content:    "Test content",
										NotebookID: 2,
									},
								},
								3: []storage.Note{
									storage.Note{
										ID:         1,
										Name:       "Test name",
										Content:    "Test content",
										NotebookID: 3,
									},
									storage.Note{
										ID:         2,
										Name:       "Test name",
										Content:    "Test content",
										NotebookID: 3,
									},
									storage.Note{
										ID:         3,
										Name:       "Test name",
										Content:    "Test content",
										NotebookID: 3,
									},
								},
							},
							nil,
						)
					return mockNoteStore
				}(),
				mock.NewNoteSearcher(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			notebookIDs: []int{1, 2, 3},
			expectedNotebookNotesMap: storage.NotebookNotesMap{
				1: []storage.Note{
					storage.Note{
						ID:         1,
						Name:       "Test name",
						Content:    "Test content",
						NotebookID: 1,
					},
					storage.Note{
						ID:         2,
						Name:       "Test name",
						Content:    "Test content",
						NotebookID: 1,
					},
					storage.Note{
						ID:         3,
						Name:       "Test name",
						Content:    "Test content",
						NotebookID: 1,
					},
				},
				2: []storage.Note{
					storage.Note{
						ID:         1,
						Name:       "Test name",
						Content:    "Test content",
						NotebookID: 2,
					},
					storage.Note{
						ID:         2,
						Name:       "Test name",
						Content:    "Test content",
						NotebookID: 2,
					},
					storage.Note{
						ID:         3,
						Name:       "Test name",
						Content:    "Test content",
						NotebookID: 2,
					},
				},
				3: []storage.Note{
					storage.Note{
						ID:         1,
						Name:       "Test name",
						Content:    "Test content",
						NotebookID: 3,
					},
					storage.Note{
						ID:         2,
						Name:       "Test name",
						Content:    "Test content",
						NotebookID: 3,
					},
					storage.Note{
						ID:         3,
						Name:       "Test name",
						Content:    "Test content",
						NotebookID: 3,
					},
				},
			},
			expectedErrorAssertionFunc: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actualNotebookNotesMap, err := test.noteService.
				FetchNotesByNotebooks(test.notebookIDs)
			assert.Equal(t, test.expectedNotebookNotesMap, actualNotebookNotesMap)
			test.expectedErrorAssertionFunc(t, err)
		})
	}
}

func TestNoteService_FetchNotesBySearchQuery(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	validator := validator.New()

	tests := []struct {
		description                string
		noteService                *NoteService
		query                      string
		page                       p8n.Page
		expectedPaginatedNotes     *PaginatedNotes
		expectedErrorAssertionFunc assert.ErrorAssertionFunc
	}{
		{
			description: "returns error because of note searcher's failure",
			noteService: NewNoteService(
				mock.NewNoteStore(mockController),
				func() search.NoteSearcher {
					mockNoteSearcher := mock.NewNoteSearcher(mockController)
					mockNoteSearcher.
						EXPECT().
						SearchByQuery("test query", gomock.Eq(10), gomock.Eq(0)).
						Return(nil, errors.New("something went wrong"))
					return mockNoteSearcher
				}(),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			query:                      "test query",
			page:                       p8n.NewPage(10, 0),
			expectedPaginatedNotes:     nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns error because of note store's failure",
			noteService: NewNoteService(
				func() storage.NoteStore {
					mockNoteStore := mock.NewNoteStore(mockController)
					mockNoteStore.
						EXPECT().
						FetchMany([]int{1, 2, 3}).
						Return(nil, errors.New("something went wrong"))
					return mockNoteStore
				}(),
				func() search.NoteSearcher {
					mockNoteSearcher := mock.NewNoteSearcher(mockController)
					mockNoteSearcher.
						EXPECT().
						SearchByQuery("test query", gomock.Eq(10), gomock.Eq(0)).
						Return(
							[]search.Note{
								search.Note{
									ID: 1,
								},
								search.Note{
									ID: 2,
								},
								search.Note{
									ID: 3,
								},
							}, nil)
					return mockNoteSearcher
				}(),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			query:                      "test query",
			page:                       p8n.NewPage(10, 0),
			expectedPaginatedNotes:     nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns 0 notes without error",
			noteService: NewNoteService(
				mock.NewNoteStore(mockController),
				func() search.NoteSearcher {
					mockNoteSearcher := mock.NewNoteSearcher(mockController)
					mockNoteSearcher.
						EXPECT().
						SearchByQuery("test query", gomock.Eq(10), gomock.Eq(0)).
						Return([]search.Note{}, nil)
					return mockNoteSearcher
				}(),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			query: "test query",
			page:  p8n.NewPage(10, 0),
			expectedPaginatedNotes: &PaginatedNotes{
				Pagination: p8n.NewPagination(0, p8n.NewPage(10, 0)),
			},
			expectedErrorAssertionFunc: assert.NoError,
		},
		{
			description: "returns notes without error",
			noteService: NewNoteService(
				func() storage.NoteStore {
					mockNoteStore := mock.NewNoteStore(mockController)
					mockNoteStore.
						EXPECT().
						FetchMany([]int{1, 2, 3}).
						Return(
							[]storage.Note{
								storage.Note{
									ID:         1,
									Name:       "Test name",
									Content:    "Test content",
									NotebookID: 1,
								},
								storage.Note{
									ID:         2,
									Name:       "Test name",
									Content:    "Test content",
									NotebookID: 1,
								},
								storage.Note{
									ID:         3,
									Name:       "Test name",
									Content:    "Test content",
									NotebookID: 1,
								},
							},
							nil,
						)
					return mockNoteStore
				}(),
				func() search.NoteSearcher {
					mockNoteSearcher := mock.NewNoteSearcher(mockController)
					mockNoteSearcher.
						EXPECT().
						SearchByQuery("test query", gomock.Eq(10), gomock.Eq(0)).
						Return(
							[]search.Note{
								search.Note{
									ID: 1,
								},
								search.Note{
									ID: 2,
								},
								search.Note{
									ID: 3,
								},
							}, nil)
					return mockNoteSearcher
				}(),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			query: "test query",
			page:  p8n.NewPage(10, 0),
			expectedPaginatedNotes: &PaginatedNotes{
				Notes: []storage.Note{
					storage.Note{
						ID:         1,
						Name:       "Test name",
						Content:    "Test content",
						NotebookID: 1,
					},
					storage.Note{
						ID:         2,
						Name:       "Test name",
						Content:    "Test content",
						NotebookID: 1,
					},
					storage.Note{
						ID:         3,
						Name:       "Test name",
						Content:    "Test content",
						NotebookID: 1,
					},
				},
				Pagination: p8n.NewPagination(3, p8n.NewPage(10, 0)),
			},
			expectedErrorAssertionFunc: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actualPaginatedNotes, err := test.noteService.FetchNotesBySearchQuery(
				test.query, test.page)
			assert.Equal(t, test.expectedPaginatedNotes, actualPaginatedNotes)
			test.expectedErrorAssertionFunc(t, err)
		})
	}
}

func TestNoteService_UpdateNote(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	validator := validator.New()

	testData := fetchNoteValidationTestData(t)

	tests := []struct {
		description                string
		noteService                *NoteService
		params                     UpdateNoteParameters
		expectedNote               *storage.Note
		expectedErrorAssertionFunc assert.ErrorAssertionFunc
	}{
		{
			description: "returns error because name parameter is empty",
			noteService: NewNoteService(
				mock.NewNoteStore(mockController),
				mock.NewNoteSearcher(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			params: UpdateNoteParameters{
				NoteID:     1,
				Name:       "",
				Content:    testData.Content.Valid,
				NotebookID: 1,
			},
			expectedNote:               nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns error because name parameter exceeds character length limit",
			noteService: NewNoteService(
				mock.NewNoteStore(mockController),
				mock.NewNoteSearcher(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			params: UpdateNoteParameters{
				NoteID:     1,
				Name:       testData.Name.ExceedingCharLimit,
				Content:    testData.Content.Valid,
				NotebookID: 1,
			},
			expectedNote:               nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns error because content parameter is empty",
			noteService: NewNoteService(
				mock.NewNoteStore(mockController),
				mock.NewNoteSearcher(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			params: UpdateNoteParameters{
				NoteID:     1,
				Name:       testData.Name.Valid,
				Content:    "",
				NotebookID: 1,
			},
			expectedNote:               nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns error because content parameter exceeds character length limit",
			noteService: NewNoteService(
				mock.NewNoteStore(mockController),
				mock.NewNoteSearcher(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			params: UpdateNoteParameters{
				NoteID:     1,
				Name:       testData.Name.Valid,
				Content:    testData.Content.ExceedingCharLimit,
				NotebookID: 1,
			},
			expectedNote:               nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns error because note store transaction could not be started",
			noteService: NewNoteService(
				func() storage.NoteStore {
					mockNoteStore := mock.NewNoteStore(mockController)
					mockNoteStore.
						EXPECT().
						BeginTx().
						Return(nil, errors.New("something went wrong"))
					return mockNoteStore
				}(),
				mock.NewNoteSearcher(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			params: UpdateNoteParameters{
				NoteID:     1,
				Name:       testData.Name.Valid,
				Content:    testData.Content.Valid,
				NotebookID: 1,
			},
			expectedNote:               nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns error because note store's failure and rollbacks transaction",
			noteService: NewNoteService(
				func() storage.NoteStore {
					mockTx := mock.NewTx(mockController)
					mockTx.EXPECT().Rollback().Times(1)

					mockNoteStore := mock.NewNoteStore(mockController)
					mockNoteStore.
						EXPECT().BeginTx().Return(mockTx, nil)
					mockNoteStore.
						EXPECT().
						UpdateOne(storage.UpdateNoteParameters{
							Transaction: mockTx,
							NoteID:      1,
							Name:        testData.Name.Valid,
							Content:     testData.Content.Valid,
							NotebookID:  1,
						}).
						Return(nil, errors.New("something went wrong"))
					return mockNoteStore
				}(),
				mock.NewNoteSearcher(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			params: UpdateNoteParameters{
				NoteID:     1,
				Name:       testData.Name.Valid,
				Content:    testData.Content.Valid,
				NotebookID: 1,
			},
			expectedNote:               nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns updated note and commits transaction without error, but fails to produce event",
			noteService: NewNoteService(
				func() storage.NoteStore {
					mockTx := mock.NewTx(mockController)
					mockTx.EXPECT().Commit().Times(1)

					mockNoteStore := mock.NewNoteStore(mockController)
					mockNoteStore.
						EXPECT().BeginTx().Return(mockTx, nil)
					mockNoteStore.
						EXPECT().
						UpdateOne(storage.UpdateNoteParameters{
							Transaction: mockTx,
							NoteID:      1,
							Name:        testData.Name.Valid,
							Content:     testData.Content.Valid,
							NotebookID:  1,
						}).
						Return(
							&storage.Note{
								ID:         1,
								Name:       testData.Name.Valid,
								Content:    testData.Content.Valid,
								NotebookID: 1,
							},
							nil,
						)
					return mockNoteStore
				}(),
				mock.NewNoteSearcher(mockController),
				validator,
				func() eventdriven.Producer {
					mockEventProducer := mock.NewEventProducer(mockController)
					mockEventProducer.
						EXPECT().
						Produce(
							eventdriven.EventTypeNoteUpdated,
							eventdriven.NoteUpdatedEventPayload{
								NoteID: 1,
							},
						).
						Return(errors.New("something went wrong"))
					return mockEventProducer
				}(),
				func() logrus.FieldLogger {
					mockLogger := mock.NewLogger(mockController)
					mockLogger.
						EXPECT().
						Error(gomock.Any()).
						Times(1)
					return mockLogger
				}(),
			),
			params: UpdateNoteParameters{
				NoteID:     1,
				Name:       testData.Name.Valid,
				Content:    testData.Content.Valid,
				NotebookID: 1,
			},
			expectedNote: &storage.Note{
				ID:         1,
				Name:       testData.Name.Valid,
				Content:    testData.Content.Valid,
				NotebookID: 1,
			},
			expectedErrorAssertionFunc: assert.NoError,
		},
		{
			description: "returns updated note and commits transaction without error, and produces event without error",
			noteService: NewNoteService(
				func() storage.NoteStore {
					mockTx := mock.NewTx(mockController)
					mockTx.EXPECT().Commit().Times(1)

					mockNoteStore := mock.NewNoteStore(mockController)
					mockNoteStore.
						EXPECT().BeginTx().Return(mockTx, nil)
					mockNoteStore.
						EXPECT().
						UpdateOne(storage.UpdateNoteParameters{
							Transaction: mockTx,
							NoteID:      1,
							Name:        testData.Name.Valid,
							Content:     testData.Content.Valid,
							NotebookID:  1,
						}).
						Return(
							&storage.Note{
								ID:         1,
								Name:       testData.Name.Valid,
								Content:    testData.Content.Valid,
								NotebookID: 1,
							},
							nil,
						)
					return mockNoteStore
				}(),
				mock.NewNoteSearcher(mockController),
				validator,
				func() eventdriven.Producer {
					mockEventProducer := mock.NewEventProducer(mockController)
					mockEventProducer.
						EXPECT().
						Produce(
							eventdriven.EventTypeNoteUpdated,
							eventdriven.NoteUpdatedEventPayload{
								NoteID: 1,
							},
						).
						Return(nil)
					return mockEventProducer
				}(),
				mock.NewLogger(mockController),
			),
			params: UpdateNoteParameters{
				NoteID:     1,
				Name:       testData.Name.Valid,
				Content:    testData.Content.Valid,
				NotebookID: 1,
			},
			expectedNote: &storage.Note{
				ID:         1,
				Name:       testData.Name.Valid,
				Content:    testData.Content.Valid,
				NotebookID: 1,
			},
			expectedErrorAssertionFunc: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actualNote, err := test.noteService.UpdateNote(test.params)
			assert.Equal(t, test.expectedNote, actualNote)
			test.expectedErrorAssertionFunc(t, err)
		})
	}
}

func TestNoteService_DeleteNote(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	validator := validator.New()

	tests := []struct {
		description                string
		noteService                *NoteService
		noteID                     int
		expectedErrorAssertionFunc assert.ErrorAssertionFunc
	}{
		{
			description: "returns error because note store transaction could not be started",
			noteService: NewNoteService(
				func() storage.NoteStore {
					mockNoteStore := mock.NewNoteStore(mockController)
					mockNoteStore.
						EXPECT().
						BeginTx().
						Return(nil, errors.New("something went wrong"))
					return mockNoteStore
				}(),
				mock.NewNoteSearcher(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			noteID:                     1,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns error because note store's failure and rollbacks transaction",
			noteService: NewNoteService(
				func() storage.NoteStore {
					mockTx := mock.NewTx(mockController)
					mockTx.EXPECT().Rollback().Times(1)

					mockNoteStore := mock.NewNoteStore(mockController)
					mockNoteStore.
						EXPECT().BeginTx().Return(mockTx, nil)
					mockNoteStore.
						EXPECT().
						DeleteOne(mockTx, 1).
						Return(errors.New("something went wrong"))
					return mockNoteStore
				}(),
				mock.NewNoteSearcher(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			noteID:                     1,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "deletes note and commits transaction without error, but fails to produce event",
			noteService: NewNoteService(
				func() storage.NoteStore {
					mockTx := mock.NewTx(mockController)
					mockTx.EXPECT().Commit().Times(1)

					mockNoteStore := mock.NewNoteStore(mockController)
					mockNoteStore.
						EXPECT().BeginTx().Return(mockTx, nil)
					mockNoteStore.
						EXPECT().
						DeleteOne(mockTx, 1).
						Return(nil)
					return mockNoteStore
				}(),
				mock.NewNoteSearcher(mockController),
				validator,
				func() eventdriven.Producer {
					mockEventProducer := mock.NewEventProducer(mockController)
					mockEventProducer.
						EXPECT().
						Produce(
							eventdriven.EventTypeNoteDeleted,
							eventdriven.NoteDeletedEventPayload{
								NoteID: 1,
							},
						).
						Return(errors.New("something went wrong"))
					return mockEventProducer
				}(),
				func() logrus.FieldLogger {
					mockLogger := mock.NewLogger(mockController)
					mockLogger.
						EXPECT().
						Error(gomock.Any()).
						Times(1)
					return mockLogger
				}(),
			),
			noteID:                     1,
			expectedErrorAssertionFunc: assert.NoError,
		},
		{
			description: "deletes note and commits transaction without error, and produces event without error",
			noteService: NewNoteService(
				func() storage.NoteStore {
					mockTx := mock.NewTx(mockController)
					mockTx.EXPECT().Commit().Times(1)

					mockNoteStore := mock.NewNoteStore(mockController)
					mockNoteStore.
						EXPECT().BeginTx().Return(mockTx, nil)
					mockNoteStore.
						EXPECT().
						DeleteOne(mockTx, 1).
						Return(nil)
					return mockNoteStore
				}(),
				mock.NewNoteSearcher(mockController),
				validator,
				func() eventdriven.Producer {
					mockEventProducer := mock.NewEventProducer(mockController)
					mockEventProducer.
						EXPECT().
						Produce(
							eventdriven.EventTypeNoteDeleted,
							eventdriven.NoteDeletedEventPayload{
								NoteID: 1,
							},
						).
						Return(nil)
					return mockEventProducer
				}(),
				mock.NewLogger(mockController),
			),
			noteID:                     1,
			expectedErrorAssertionFunc: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			err := test.noteService.DeleteNote(test.noteID)
			test.expectedErrorAssertionFunc(t, err)
		})
	}
}
