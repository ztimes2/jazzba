package main

import (
	"log"

	"github.com/sirupsen/logrus"
	"github.com/ztimes2/jazzba/pkg/eventdriven/rabbit"
	"github.com/ztimes2/jazzba/pkg/eventer"
	"github.com/ztimes2/jazzba/pkg/search/elastic"
	"github.com/ztimes2/jazzba/pkg/storage/postgres"
)

func main() {
	cfg, err := eventer.LoadConfig()
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	db, err := postgres.NewDB(postgres.Config{
		Host:     cfg.PostgresConfig.Host,
		Port:     cfg.PostgresConfig.Port,
		User:     cfg.PostgresConfig.User,
		Password: cfg.PostgresConfig.Password,
		DBName:   cfg.PostgresConfig.Name,
		SSLMode:  postgres.SSLMode(cfg.PostgresConfig.SSLMode),
	})
	if err != nil {
		log.Fatalf("could not init postgres db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("could not ping postgres db: %v", err)
	}

	esClient, err := elastic.NewClient(elastic.Config{
		Host:     cfg.ElasticSearchConfig.Host,
		Port:     cfg.ElasticSearchConfig.Port,
		Username: cfg.ElasticSearchConfig.Username,
		Password: cfg.ElasticSearchConfig.Password,
	})
	if err != nil {
		log.Fatalf("could not init elasticsearch client: %v", err)
	}
	defer esClient.Stop()

	if err := elastic.InitIndices(esClient); err != nil {
		log.Fatalf("could not init elasticsearch indices: %v", err)
	}

	amqpConnection, err := rabbit.NewConnection(rabbit.Config{
		Host:     cfg.RabbitMQConfig.Host,
		Port:     cfg.RabbitMQConfig.Port,
		Username: cfg.RabbitMQConfig.Username,
		Password: cfg.RabbitMQConfig.Password,
	})
	if err != nil {
		log.Fatalf("could not init rabbitmq connection: %v", err)
	}
	defer amqpConnection.Close()

	if err := rabbit.DeclareQueues(amqpConnection); err != nil {
		log.Fatalf("could not declare rabbitmq queues: %v", err)
	}

	eventer.New(eventer.Dependencies{
		EventConsumer:   rabbit.NewEventConsumer(amqpConnection),
		NoteStore:       postgres.NewNoteStore(db),
		NoteTagStore:    postgres.NewNoteTagStore(db),
		NotebookStore:   postgres.NewNotebookStore(db),
		NoteUpdater:     elastic.NewNoteUpdater(esClient),
		NoteTagUpdater:  elastic.NewNoteTagUpdater(esClient),
		NotebookUpdater: elastic.NewNotebookUpdater(esClient),
		Logger:          logrus.New(),
	}).Run()
}
