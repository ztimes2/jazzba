package service

import (
	"errors"
	"testing"
	"time"

	"github.com/ztimes2/jazzba/pkg/api/p8n"
	"github.com/ztimes2/jazzba/pkg/eventdriven"
	"github.com/ztimes2/jazzba/pkg/mock"
	"github.com/ztimes2/jazzba/pkg/storage"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var _ Notebooker = (*NotebookService)(nil)

func TestNotebookService_CreateNotebook(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	validator := validator.New()

	tests := []struct {
		description                string
		notebookService            *NotebookService
		params                     CreateNotebookParameters
		expectedNotebook           *storage.Notebook
		expectedErrorAssertionFunc assert.ErrorAssertionFunc
	}{
		{
			description: "returns error because the name parameter is empty",
			notebookService: NewNotebookService(
				mock.NewNotebookStore(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			params:                     CreateNotebookParameters{},
			expectedNotebook:           nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns error because the name parameter exceeds the character length limit",
			notebookService: NewNotebookService(
				mock.NewNotebookStore(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			params: CreateNotebookParameters{
				Name: "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec qua",
			},
			expectedNotebook:           nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns error because notebook store transaction could not be started",
			notebookService: NewNotebookService(
				func() storage.NotebookStore {
					mockNotebookStore := mock.NewNotebookStore(mockController)
					mockNotebookStore.EXPECT().
						BeginTx().
						Return(nil, errors.New("something went wrong"))
					return mockNotebookStore
				}(),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			params: CreateNotebookParameters{
				Name: "Lorem ipsum dolor sit amet",
			},
			expectedNotebook:           nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns error because of notebook store's failure and rollbacks transaction",
			notebookService: NewNotebookService(
				func() storage.NotebookStore {
					mockTx := mock.NewTx(mockController)
					mockTx.EXPECT().Rollback().Times(1)

					mockNotebookStore := mock.NewNotebookStore(mockController)
					mockNotebookStore.
						EXPECT().
						BeginTx().
						Return(mockTx, nil)
					mockNotebookStore.
						EXPECT().
						CreateOne(mockTx, "Lorem ipsum dolor sit amet").
						Return(nil, errors.New("something went wrong"))

					return mockNotebookStore
				}(),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			params: CreateNotebookParameters{
				Name: "Lorem ipsum dolor sit amet",
			},
			expectedNotebook:           nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns created notebook and commits transaction without error",
			notebookService: NewNotebookService(
				func() storage.NotebookStore {
					mockTx := mock.NewTx(mockController)
					mockTx.EXPECT().Commit().Times(1)

					mockNotebookStore := mock.NewNotebookStore(mockController)
					mockNotebookStore.
						EXPECT().
						BeginTx().
						Return(mockTx, nil)
					mockNotebookStore.
						EXPECT().
						CreateOne(mockTx, "Lorem ipsum dolor sit amet").
						Return(
							&storage.Notebook{
								ID:        1,
								Name:      "Lorem ipsum dolor sit amet",
								CreatedAt: time.Time{},
								UpdatedAt: time.Time{},
							},
							nil,
						)

					return mockNotebookStore
				}(),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			params: CreateNotebookParameters{
				Name: "Lorem ipsum dolor sit amet",
			},
			expectedNotebook: &storage.Notebook{
				ID:        1,
				Name:      "Lorem ipsum dolor sit amet",
				CreatedAt: time.Time{},
				UpdatedAt: time.Time{},
			},
			expectedErrorAssertionFunc: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actualNotebook, err := test.notebookService.CreateNotebook(test.params)
			assert.Equal(t, test.expectedNotebook, actualNotebook)
			test.expectedErrorAssertionFunc(t, err)
		})
	}
}

func TestNotebookService_FetchNotebook(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	validator := validator.New()

	tests := []struct {
		description                string
		notebookService            *NotebookService
		notebookID                 int
		expectedNotebook           *storage.Notebook
		expectedErrorAssertionFunc assert.ErrorAssertionFunc
	}{
		{
			description: "returns error because of notebook store's failure",
			notebookService: NewNotebookService(
				func() storage.NotebookStore {
					mockNotebookStore := mock.NewNotebookStore(mockController)
					mockNotebookStore.
						EXPECT().
						FetchOne(gomock.Eq(1)).
						Return(nil, errors.New("something went wrong"))
					return mockNotebookStore
				}(),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			notebookID:                 1,
			expectedNotebook:           nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns notebook without error",
			notebookService: NewNotebookService(
				func() storage.NotebookStore {
					mockNotebookStore := mock.NewNotebookStore(mockController)
					mockNotebookStore.
						EXPECT().
						FetchOne(gomock.Eq(1)).
						Return(
							&storage.Notebook{
								ID:        1,
								Name:      "Lorem ipsum dolor sit amet",
								CreatedAt: time.Time{},
								UpdatedAt: time.Time{},
							},
							nil,
						)
					return mockNotebookStore
				}(),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			notebookID: 1,
			expectedNotebook: &storage.Notebook{
				ID:        1,
				Name:      "Lorem ipsum dolor sit amet",
				CreatedAt: time.Time{},
				UpdatedAt: time.Time{},
			},
			expectedErrorAssertionFunc: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actualNotebook, err := test.notebookService.FetchNotebook(test.notebookID)
			assert.Equal(t, test.expectedNotebook, actualNotebook)
			test.expectedErrorAssertionFunc(t, err)
		})
	}
}

func TestNotebookService_FetchNotebooks(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	validator := validator.New()

	tests := []struct {
		description                string
		notebookService            *NotebookService
		page                       p8n.Page
		expectedPaginatedNotebooks *PaginatedNotebooks
		expectedErrorAssertionFunc assert.ErrorAssertionFunc
	}{
		{
			description: "returns error because of notebook store's failure",
			notebookService: NewNotebookService(
				func() storage.NotebookStore {
					mockNotebookStore := mock.NewNotebookStore(mockController)
					mockNotebookStore.
						EXPECT().
						FetchAllPaginated(gomock.Eq(10), gomock.Eq(0)).
						Return(nil, errors.New("something went wrong"))
					return mockNotebookStore
				}(),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			page:                       p8n.NewPage(10, 0),
			expectedPaginatedNotebooks: nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns notebooks without error",
			notebookService: NewNotebookService(
				func() storage.NotebookStore {
					mockNotebookStore := mock.NewNotebookStore(mockController)
					mockNotebookStore.
						EXPECT().
						FetchAllPaginated(gomock.Eq(10), gomock.Eq(0)).
						Return(
							[]storage.Notebook{
								storage.Notebook{
									ID:        1,
									Name:      "Lorem ipsum dolor sit amet",
									CreatedAt: time.Time{},
									UpdatedAt: time.Time{},
								},
								storage.Notebook{
									ID:        2,
									Name:      "Lorem ipsum dolor sit amet",
									CreatedAt: time.Time{},
									UpdatedAt: time.Time{},
								},
								storage.Notebook{
									ID:        3,
									Name:      "Lorem ipsum dolor sit amet",
									CreatedAt: time.Time{},
									UpdatedAt: time.Time{},
								},
							},
							nil,
						)
					return mockNotebookStore
				}(),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			page: p8n.NewPage(10, 0),
			expectedPaginatedNotebooks: &PaginatedNotebooks{
				Notebooks: []storage.Notebook{
					storage.Notebook{
						ID:        1,
						Name:      "Lorem ipsum dolor sit amet",
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					},
					storage.Notebook{
						ID:        2,
						Name:      "Lorem ipsum dolor sit amet",
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					},
					storage.Notebook{
						ID:        3,
						Name:      "Lorem ipsum dolor sit amet",
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
			actualPaginatedNotebooks, err := test.notebookService.FetchNotebooks(test.page)
			assert.Equal(t, test.expectedPaginatedNotebooks, actualPaginatedNotebooks)
			test.expectedErrorAssertionFunc(t, err)
		})
	}
}

func TestNotebookService_UpdateNotebook(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	validator := validator.New()

	tests := []struct {
		description                string
		notebookService            *NotebookService
		params                     UpdateNotebookParameters
		expectedNotebook           *storage.Notebook
		expectedErrorAssertionFunc assert.ErrorAssertionFunc
	}{
		{
			description: "returns error because the name parameter is empty",
			notebookService: NewNotebookService(
				mock.NewNotebookStore(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			params: UpdateNotebookParameters{
				NotebookID: 1,
				Name:       "",
			},
			expectedNotebook:           nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns error because the name parameter exceeds the character length limit",
			notebookService: NewNotebookService(
				mock.NewNotebookStore(mockController),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			params: UpdateNotebookParameters{
				NotebookID: 1,
				Name:       "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec qua",
			},
			expectedNotebook:           nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns error because notebook store transaction could not be started",
			notebookService: NewNotebookService(
				func() storage.NotebookStore {
					mockNotebookStore := mock.NewNotebookStore(mockController)
					mockNotebookStore.EXPECT().
						BeginTx().
						Return(nil, errors.New("something went wrong"))
					return mockNotebookStore
				}(),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			params: UpdateNotebookParameters{
				NotebookID: 1,
				Name:       "Lorem ipsum dolor sit amet",
			},
			expectedNotebook:           nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns error because of notebook store's failure and rollbacks transaction",
			notebookService: NewNotebookService(
				func() storage.NotebookStore {
					mockTx := mock.NewTx(mockController)
					mockTx.EXPECT().Rollback().Times(1)

					mockNotebookStore := mock.NewNotebookStore(mockController)
					mockNotebookStore.
						EXPECT().
						BeginTx().
						Return(mockTx, nil)
					mockNotebookStore.
						EXPECT().
						UpdateOne(storage.UpdateNotebookParameters{
							Transaction: mockTx,
							NotebookID:  1,
							Name:        "Lorem ipsum dolor sit amet",
						}).
						Return(nil, errors.New("something went wrong"))

					return mockNotebookStore
				}(),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			params: UpdateNotebookParameters{
				NotebookID: 1,
				Name:       "Lorem ipsum dolor sit amet",
			},
			expectedNotebook:           nil,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns updated notebook and commits transaction without error, but fails to produce event",
			notebookService: NewNotebookService(
				func() storage.NotebookStore {
					mockTx := mock.NewTx(mockController)
					mockTx.EXPECT().Commit().Times(1)

					mockNotebookStore := mock.NewNotebookStore(mockController)
					mockNotebookStore.
						EXPECT().
						BeginTx().
						Return(mockTx, nil)
					mockNotebookStore.
						EXPECT().
						UpdateOne(storage.UpdateNotebookParameters{
							Transaction: mockTx,
							NotebookID:  1,
							Name:        "Lorem ipsum dolor sit amet",
						}).
						Return(
							&storage.Notebook{
								ID:        1,
								Name:      "Lorem ipsum dolor sit amet",
								CreatedAt: time.Time{},
								UpdatedAt: time.Time{},
							},
							nil,
						)

					return mockNotebookStore
				}(),
				validator,
				func() eventdriven.Producer {
					mockEventProducer := mock.NewEventProducer(mockController)
					mockEventProducer.
						EXPECT().
						Produce(
							eventdriven.EventTypeNotebookUpdated,
							eventdriven.NotebookUpdatedEventPayload{
								NotebookID: 1,
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
			params: UpdateNotebookParameters{
				NotebookID: 1,
				Name:       "Lorem ipsum dolor sit amet",
			},
			expectedNotebook: &storage.Notebook{
				ID:        1,
				Name:      "Lorem ipsum dolor sit amet",
				CreatedAt: time.Time{},
				UpdatedAt: time.Time{},
			},
			expectedErrorAssertionFunc: assert.NoError,
		},
		{
			description: "returns updated notebook and commits transaction without error, and produces event without error",
			notebookService: NewNotebookService(
				func() storage.NotebookStore {
					mockTx := mock.NewTx(mockController)
					mockTx.EXPECT().Commit().Times(1)

					mockNotebookStore := mock.NewNotebookStore(mockController)
					mockNotebookStore.
						EXPECT().
						BeginTx().
						Return(mockTx, nil)
					mockNotebookStore.
						EXPECT().
						UpdateOne(storage.UpdateNotebookParameters{
							Transaction: mockTx,
							NotebookID:  1,
							Name:        "Lorem ipsum dolor sit amet",
						}).
						Return(
							&storage.Notebook{
								ID:        1,
								Name:      "Lorem ipsum dolor sit amet",
								CreatedAt: time.Time{},
								UpdatedAt: time.Time{},
							},
							nil,
						)

					return mockNotebookStore
				}(),
				validator,
				func() eventdriven.Producer {
					mockEventProducer := mock.NewEventProducer(mockController)
					mockEventProducer.
						EXPECT().
						Produce(
							eventdriven.EventTypeNotebookUpdated,
							eventdriven.NotebookUpdatedEventPayload{
								NotebookID: 1,
							},
						).
						Return(nil)
					return mockEventProducer
				}(),
				mock.NewLogger(mockController),
			),
			params: UpdateNotebookParameters{
				NotebookID: 1,
				Name:       "Lorem ipsum dolor sit amet",
			},
			expectedNotebook: &storage.Notebook{
				ID:        1,
				Name:      "Lorem ipsum dolor sit amet",
				CreatedAt: time.Time{},
				UpdatedAt: time.Time{},
			},
			expectedErrorAssertionFunc: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actualNotebook, err := test.notebookService.UpdateNotebook(test.params)
			assert.Equal(t, test.expectedNotebook, actualNotebook)
			test.expectedErrorAssertionFunc(t, err)
		})
	}
}

func TestNotebookService_DeleteNotebook(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	validator := validator.New()

	tests := []struct {
		description                string
		notebookService            *NotebookService
		notebookID                 int
		expectedErrorAssertionFunc assert.ErrorAssertionFunc
	}{
		{
			description: "returns error because notebook store transaction could not be started",
			notebookService: NewNotebookService(
				func() storage.NotebookStore {
					mockNotebookStore := mock.NewNotebookStore(mockController)
					mockNotebookStore.EXPECT().
						BeginTx().
						Return(nil, errors.New("something went wrong"))
					return mockNotebookStore
				}(),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			notebookID:                 1,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "returns error because of notebook store's failure and rollbacks transaction",
			notebookService: NewNotebookService(
				func() storage.NotebookStore {
					mockTx := mock.NewTx(mockController)
					mockTx.EXPECT().Rollback().Times(1)

					mockNotebookStore := mock.NewNotebookStore(mockController)
					mockNotebookStore.
						EXPECT().
						BeginTx().
						Return(mockTx, nil)
					mockNotebookStore.
						EXPECT().
						DeleteOne(mockTx, 1).
						Return(errors.New("something went wrong"))

					return mockNotebookStore
				}(),
				validator,
				mock.NewEventProducer(mockController),
				mock.NewLogger(mockController),
			),
			notebookID:                 1,
			expectedErrorAssertionFunc: assert.Error,
		},
		{
			description: "deletes notebook and commits transaction without error, but fails to produce event",
			notebookService: NewNotebookService(
				func() storage.NotebookStore {
					mockTx := mock.NewTx(mockController)
					mockTx.EXPECT().Commit().Times(1)

					mockNotebookStore := mock.NewNotebookStore(mockController)
					mockNotebookStore.
						EXPECT().
						BeginTx().
						Return(mockTx, nil)
					mockNotebookStore.
						EXPECT().
						DeleteOne(mockTx, 1).
						Return(nil)

					return mockNotebookStore
				}(),
				validator,
				func() eventdriven.Producer {
					mockEventProducer := mock.NewEventProducer(mockController)
					mockEventProducer.
						EXPECT().
						Produce(
							eventdriven.EventTypeNotebookDeleted,
							eventdriven.NotebookDeletedEventPayload{
								NotebookID: 1,
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
			notebookID:                 1,
			expectedErrorAssertionFunc: assert.NoError,
		},
		{
			description: "deletes notebook and commits transaction without error, but fails to produce event",
			notebookService: NewNotebookService(
				func() storage.NotebookStore {
					mockTx := mock.NewTx(mockController)
					mockTx.EXPECT().Commit().Times(1)

					mockNotebookStore := mock.NewNotebookStore(mockController)
					mockNotebookStore.
						EXPECT().
						BeginTx().
						Return(mockTx, nil)
					mockNotebookStore.
						EXPECT().
						DeleteOne(mockTx, 1).
						Return(nil)

					return mockNotebookStore
				}(),
				validator,
				func() eventdriven.Producer {
					mockEventProducer := mock.NewEventProducer(mockController)
					mockEventProducer.
						EXPECT().
						Produce(
							eventdriven.EventTypeNotebookDeleted,
							eventdriven.NotebookDeletedEventPayload{
								NotebookID: 1,
							},
						).
						Return(nil)
					return mockEventProducer
				}(),
				mock.NewLogger(mockController),
			),
			notebookID:                 1,
			expectedErrorAssertionFunc: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			err := test.notebookService.DeleteNotebook(test.notebookID)
			test.expectedErrorAssertionFunc(t, err)
		})
	}
}
