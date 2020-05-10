package service

import (
	"errors"
	"testing"
	"time"

	"github.com/ztimes2/jazzba/pkg/api/p8n"
	"github.com/ztimes2/jazzba/pkg/eventdriven"
	"github.com/ztimes2/jazzba/pkg/mock"
	"github.com/ztimes2/jazzba/pkg/storage"

	validator "github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var _ NoteTagger = (*NoteTagService)(nil)

func TestNoteTagService_CreateNoteTag(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	validator := validator.New()

	tests := []struct {
		description                string
		noteTagService             *NoteTagService
		params                     CreateNoteTagParameters
		expectedNoteTag            *storage.NoteTag
		expectedErrorAssertionFunc assert.ErrorAssertionFunc
	}{
		{
			description: "returns error because tag name is empty",
			noteTagService: NewNoteTagService(
				mock.NewNoteTagStore(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			params: CreateNoteTagParameters{
				NoteID:  1,
				TagName: "",
			},
			expectedNoteTag:            nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns error because tag name exceeds character length limit",
			noteTagService: NewNoteTagService(
				mock.NewNoteTagStore(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			params: CreateNoteTagParameters{
				NoteID:  1,
				TagName: "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean ma",
			},
			expectedNoteTag:            nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns error because note tag store transaction could not be started",
			noteTagService: NewNoteTagService(
				func() storage.NoteTagStore {
					mockNoteTagStore := mock.NewNoteTagStore(mockController)
					mockNoteTagStore.
						EXPECT().
						BeginTx().
						Return(nil, errors.New("something went wrong"))
					return mockNoteTagStore
				}(),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			params: CreateNoteTagParameters{
				NoteID:  1,
				TagName: "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
			},
			expectedNoteTag:            nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns error because of note tag store's failure and rollbacks transaction",
			noteTagService: NewNoteTagService(
				func() storage.NoteTagStore {
					mockTx := mock.NewTx(mockController)
					mockTx.EXPECT().Rollback().Times(1)

					mockNoteTagStore := mock.NewNoteTagStore(mockController)
					mockNoteTagStore.
						EXPECT().
						BeginTx().
						Return(mockTx, nil)
					mockNoteTagStore.
						EXPECT().
						CreateOne(mockTx, storage.CreateNoteTagParameters{
							NoteID:  1,
							TagName: "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
						}).
						Return(nil, errors.New("something went wrong"))
					return mockNoteTagStore
				}(),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			params: CreateNoteTagParameters{
				NoteID:  1,
				TagName: "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
			},
			expectedNoteTag:            nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns created note tag and commits transaction without error, but fails to produce event",
			noteTagService: NewNoteTagService(
				func() storage.NoteTagStore {
					mockTx := mock.NewTx(mockController)
					mockTx.EXPECT().Commit().Times(1)

					mockNoteTagStore := mock.NewNoteTagStore(mockController)
					mockNoteTagStore.
						EXPECT().
						BeginTx().
						Return(mockTx, nil)
					mockNoteTagStore.
						EXPECT().
						CreateOne(mockTx, storage.CreateNoteTagParameters{
							NoteID:  1,
							TagName: "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
						}).
						Return(
							&storage.NoteTag{
								NoteID:    1,
								TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
								CreatedAt: time.Time{},
								UpdatedAt: time.Time{},
							},
							nil,
						)
					return mockNoteTagStore
				}(),
				validator,
				func() eventdriven.Producer {
					mockEventProducer := mock.NewEventProducer(mockController)
					mockEventProducer.
						EXPECT().
						Produce(
							eventdriven.EventTypeNoteTagCreated,
							eventdriven.NoteTagCreatedEventPayload{
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
			params: CreateNoteTagParameters{
				NoteID:  1,
				TagName: "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
			},
			expectedNoteTag: &storage.NoteTag{
				NoteID:    1,
				TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
				CreatedAt: time.Time{},
				UpdatedAt: time.Time{},
			},
			expectedErrorAssertionFunc: assert.NoError,
		},
		{
			description: "returns created note tag and commits transaction without error, and produces event without error",
			noteTagService: NewNoteTagService(
				func() storage.NoteTagStore {
					mockTx := mock.NewTx(mockController)
					mockTx.EXPECT().Commit().Times(1)

					mockNoteTagStore := mock.NewNoteTagStore(mockController)
					mockNoteTagStore.
						EXPECT().
						BeginTx().
						Return(mockTx, nil)
					mockNoteTagStore.
						EXPECT().
						CreateOne(mockTx, storage.CreateNoteTagParameters{
							NoteID:  1,
							TagName: "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
						}).
						Return(
							&storage.NoteTag{
								NoteID:    1,
								TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
								CreatedAt: time.Time{},
								UpdatedAt: time.Time{},
							},
							nil,
						)
					return mockNoteTagStore
				}(),
				validator,
				func() eventdriven.Producer {
					mockEventProducer := mock.NewEventProducer(mockController)
					mockEventProducer.
						EXPECT().
						Produce(
							eventdriven.EventTypeNoteTagCreated,
							eventdriven.NoteTagCreatedEventPayload{
								NoteID: 1,
							},
						).
						Return(nil)
					return mockEventProducer
				}(),
				mock.NewLogger(mockController),
			),
			params: CreateNoteTagParameters{
				NoteID:  1,
				TagName: "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
			},
			expectedNoteTag: &storage.NoteTag{
				NoteID:    1,
				TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
				CreatedAt: time.Time{},
				UpdatedAt: time.Time{},
			},
			expectedErrorAssertionFunc: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actualNoteTag, err := test.noteTagService.CreateNoteTag(test.params)
			assert.Equal(t, test.expectedNoteTag, actualNoteTag)
			test.expectedErrorAssertionFunc(t, err)
		})
	}
}

func TestNoteTagService_FetchNoteTagsByNote(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	validator := validator.New()

	tests := []struct {
		description                string
		noteTagService             *NoteTagService
		noteID                     int
		page                       p8n.Page
		expectedPaginatedNoteTags  *PaginatedNoteTags
		expectedErrorAssertionFunc assert.ErrorAssertionFunc
	}{
		{
			description: "returns error because of note tag store's failure",
			noteTagService: NewNoteTagService(
				func() storage.NoteTagStore {
					mockNoteTagStore := mock.NewNoteTagStore(mockController)
					mockNoteTagStore.
						EXPECT().
						FetchManyByNotePaginated(
							gomock.Eq(1),
							gomock.Eq(10),
							gomock.Eq(0),
						).
						Return(nil, errors.New("something went wrong"))
					return mockNoteTagStore
				}(),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			noteID:                     1,
			page:                       p8n.NewPage(10, 0),
			expectedPaginatedNoteTags:  nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns note tags without error",
			noteTagService: NewNoteTagService(
				func() storage.NoteTagStore {
					mockNoteTagStore := mock.NewNoteTagStore(mockController)
					mockNoteTagStore.
						EXPECT().
						FetchManyByNotePaginated(
							gomock.Eq(1),
							gomock.Eq(10),
							gomock.Eq(0),
						).
						Return(
							[]storage.NoteTag{
								storage.NoteTag{
									NoteID:    1,
									TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
									CreatedAt: time.Time{},
									UpdatedAt: time.Time{},
								},
								storage.NoteTag{
									NoteID:    1,
									TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
									CreatedAt: time.Time{},
									UpdatedAt: time.Time{},
								},
								storage.NoteTag{
									NoteID:    1,
									TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
									CreatedAt: time.Time{},
									UpdatedAt: time.Time{},
								},
							},
							nil,
						)
					return mockNoteTagStore
				}(),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			noteID: 1,
			page:   p8n.NewPage(10, 0),
			expectedPaginatedNoteTags: &PaginatedNoteTags{
				NoteTags: []storage.NoteTag{
					storage.NoteTag{
						NoteID:    1,
						TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					},
					storage.NoteTag{
						NoteID:    1,
						TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					},
					storage.NoteTag{
						NoteID:    1,
						TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					},
				},
				Pagination: p8n.NewPagination(3, p8n.NewPage(10, 0)),
			},
			expectedErrorAssertionFunc: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actualPaginatedNoteTags, err := test.noteTagService.FetchNoteTagsByNote(
				test.noteID, test.page)
			assert.Equal(t, test.expectedPaginatedNoteTags, actualPaginatedNoteTags)
			test.expectedErrorAssertionFunc(t, err)
		})
	}
}

func TestNoteTagService_FetchNoteTagsByNotes(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	validator := validator.New()

	tests := []struct {
		description                string
		noteTagService             *NoteTagService
		noteIDs                    []int
		expectedNoteTagsMap        storage.NoteTagsMap
		expectedErrorAssertionFunc assert.ErrorAssertionFunc
	}{
		{
			description: "returns error because of note tag store's failure",
			noteTagService: NewNoteTagService(
				func() storage.NoteTagStore {
					mockNoteTagStore := mock.NewNoteTagStore(mockController)
					mockNoteTagStore.
						EXPECT().
						FetchManyByNotes(
							gomock.Eq([]int{1, 2, 3}),
						).
						Return(nil, errors.New("something went wrong"))
					return mockNoteTagStore
				}(),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			noteIDs:                    []int{1, 2, 3},
			expectedNoteTagsMap:        nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns note tags map without error",
			noteTagService: NewNoteTagService(
				func() storage.NoteTagStore {
					mockNoteTagStore := mock.NewNoteTagStore(mockController)
					mockNoteTagStore.
						EXPECT().
						FetchManyByNotes(gomock.Eq([]int{1, 2, 3})).
						Return(
							storage.NoteTagsMap{
								1: []storage.NoteTag{
									storage.NoteTag{
										NoteID:    1,
										TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
										CreatedAt: time.Time{},
										UpdatedAt: time.Time{},
									},
									storage.NoteTag{
										NoteID:    1,
										TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
										CreatedAt: time.Time{},
										UpdatedAt: time.Time{},
									},
									storage.NoteTag{
										NoteID:    1,
										TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
										CreatedAt: time.Time{},
										UpdatedAt: time.Time{},
									},
								},
								2: []storage.NoteTag{
									storage.NoteTag{
										NoteID:    2,
										TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
										CreatedAt: time.Time{},
										UpdatedAt: time.Time{},
									},
									storage.NoteTag{
										NoteID:    2,
										TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
										CreatedAt: time.Time{},
										UpdatedAt: time.Time{},
									},
									storage.NoteTag{
										NoteID:    2,
										TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
										CreatedAt: time.Time{},
										UpdatedAt: time.Time{},
									},
								},
								3: []storage.NoteTag{
									storage.NoteTag{
										NoteID:    3,
										TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
										CreatedAt: time.Time{},
										UpdatedAt: time.Time{},
									},
									storage.NoteTag{
										NoteID:    3,
										TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
										CreatedAt: time.Time{},
										UpdatedAt: time.Time{},
									},
									storage.NoteTag{
										NoteID:    3,
										TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
										CreatedAt: time.Time{},
										UpdatedAt: time.Time{},
									},
								},
							},
							nil,
						)
					return mockNoteTagStore
				}(),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			noteIDs: []int{1, 2, 3},
			expectedNoteTagsMap: storage.NoteTagsMap{
				1: []storage.NoteTag{
					storage.NoteTag{
						NoteID:    1,
						TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					},
					storage.NoteTag{
						NoteID:    1,
						TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					},
					storage.NoteTag{
						NoteID:    1,
						TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					},
				},
				2: []storage.NoteTag{
					storage.NoteTag{
						NoteID:    2,
						TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					},
					storage.NoteTag{
						NoteID:    2,
						TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					},
					storage.NoteTag{
						NoteID:    2,
						TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					},
				},
				3: []storage.NoteTag{
					storage.NoteTag{
						NoteID:    3,
						TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					},
					storage.NoteTag{
						NoteID:    3,
						TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					},
					storage.NoteTag{
						NoteID:    3,
						TagName:   "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					},
				},
			},
			expectedErrorAssertionFunc: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actualNoteTagsMap, err := test.noteTagService.FetchNoteTagsByNotes(
				test.noteIDs)
			assert.Equal(t, test.expectedNoteTagsMap, actualNoteTagsMap)
			test.expectedErrorAssertionFunc(t, err)
		})
	}
}

func TestNoteTagService_DeleteNoteTag(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	validator := validator.New()

	tests := []struct {
		description                string
		noteTagService             *NoteTagService
		noteID                     int
		tagName                    string
		expectedErrorAssertionFunc assert.ErrorAssertionFunc
	}{
		{
			description: "returns error because note tag store transaction could not be started",
			noteTagService: NewNoteTagService(
				func() storage.NoteTagStore {
					mockNoteTagStore := mock.NewNoteTagStore(mockController)
					mockNoteTagStore.
						EXPECT().
						BeginTx().
						Return(nil, errors.New("something went wrong"))
					return mockNoteTagStore
				}(),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			noteID:                     1,
			tagName:                    "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns error because of note tag store's failure and rollbacks transaction",
			noteTagService: NewNoteTagService(
				func() storage.NoteTagStore {
					mockTx := mock.NewTx(mockController)
					mockTx.EXPECT().Rollback().Times(1)

					mockNoteTagStore := mock.NewNoteTagStore(mockController)
					mockNoteTagStore.
						EXPECT().
						BeginTx().
						Return(mockTx, nil)
					mockNoteTagStore.
						EXPECT().
						DeleteOne(mockTx, storage.DeleteNoteTagParameters{
							NoteID:  1,
							TagName: "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
						}).
						Return(errors.New("something went wrong"))
					return mockNoteTagStore
				}(),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			noteID:                     1,
			tagName:                    "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "deletes note tag and commits transaction without error, but fails to produce event",
			noteTagService: NewNoteTagService(
				func() storage.NoteTagStore {
					mockTx := mock.NewTx(mockController)
					mockTx.EXPECT().Commit().Times(1)

					mockNoteTagStore := mock.NewNoteTagStore(mockController)
					mockNoteTagStore.
						EXPECT().
						BeginTx().
						Return(mockTx, nil)
					mockNoteTagStore.
						EXPECT().
						DeleteOne(mockTx, storage.DeleteNoteTagParameters{
							NoteID:  1,
							TagName: "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
						}).
						Return(nil)
					return mockNoteTagStore
				}(),
				validator,
				func() eventdriven.Producer {
					mockEventProducer := mock.NewEventProducer(mockController)
					mockEventProducer.
						EXPECT().
						Produce(
							eventdriven.EventTypeNoteTagDeleted,
							eventdriven.NoteTagDeletedEventPayload{
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
			tagName:                    "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
			expectedErrorAssertionFunc: assert.NoError,
		},
		{
			description: "deletes note tag and commits transaction without error, and produces event without error",
			noteTagService: NewNoteTagService(
				func() storage.NoteTagStore {
					mockTx := mock.NewTx(mockController)
					mockTx.EXPECT().Commit().Times(1)

					mockNoteTagStore := mock.NewNoteTagStore(mockController)
					mockNoteTagStore.
						EXPECT().
						BeginTx().
						Return(mockTx, nil)
					mockNoteTagStore.
						EXPECT().
						DeleteOne(mockTx, storage.DeleteNoteTagParameters{
							NoteID:  1,
							TagName: "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
						}).
						Return(nil)
					return mockNoteTagStore
				}(),
				validator,
				func() eventdriven.Producer {
					mockEventProducer := mock.NewEventProducer(mockController)
					mockEventProducer.
						EXPECT().
						Produce(
							eventdriven.EventTypeNoteTagDeleted,
							eventdriven.NoteTagDeletedEventPayload{
								NoteID: 1,
							},
						).
						Return(nil)
					return mockEventProducer
				}(),
				mock.NewLogger(mockController),
			),
			noteID:                     1,
			tagName:                    "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean m",
			expectedErrorAssertionFunc: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			err := test.noteTagService.DeleteNoteTag(test.noteID, test.tagName)
			test.expectedErrorAssertionFunc(t, err)
		})
	}
}
