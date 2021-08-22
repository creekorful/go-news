package scheduler

import (
	"context"
	"fmt"
	"github.com/creekorful/go-news/internal/database"
	"github.com/mmcdole/gofeed"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Scheduler struct {
	db *database.Database

	processing      bool
	processingMutex sync.RWMutex
}

func NewScheduler(db *database.Database) *Scheduler {
	return &Scheduler{db: db}
}

func (s *Scheduler) Schedule(ctx context.Context, interval time.Duration) error {
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ticker.C:
			if err := s.process(); err != nil {
				log.Printf("error while processing feeds: %s", err)
			}
		case <-ctx.Done():
			ticker.Stop()
			return nil
		}
	}
}

func (s *Scheduler) process() error {
	if s.isProcessing() {
		// execution already in progress
		return nil
	}

	s.processingMutex.Lock()
	defer s.processingMutex.Unlock()

	s.processing = true
	defer func() {
		s.processing = false
	}()

	feeds, err := s.db.GetFeeds()
	if err != nil {
		return err
	}

	fp := gofeed.NewParser()
	for _, feed := range feeds {
		log.Printf("processing feed `%s`", feed.Link)

		res, err := fp.ParseURL(feed.Link)
		if err != nil {
			log.Printf("skipping feed `%s`: %s", feed.Link, err)
			continue
		}

		for _, newItem := range res.Items {
			alreadyExists := false

			for _, existingItem := range feed.Items {
				if existingItem.GUID == newItem.GUID {
					alreadyExists = true
					break
				}
			}

			if !alreadyExists {
				log.Printf("saving new item `%s`", newItem.Title)

				item := &database.Item{
					Title:       newItem.Title,
					Description: newItem.Description,
					GUID:        newItem.GUID,
					FeedID:      feed.ID,
					PublishDate: newItem.PublishedParsed,
				}

				if strings.HasPrefix(newItem.Link, "http") {
					item.Link = newItem.Link
				} else {
					// relative URL
					u, err := url.Parse(feed.Link)
					if err != nil {
						return err
					}

					item.Link = fmt.Sprintf("%s:%s%s", u.Scheme, u.Host, newItem.Link)
				}

				if err := s.db.CreateItem(&feed, item); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *Scheduler) isProcessing() bool {
	s.processingMutex.RLock()
	defer s.processingMutex.RUnlock()

	return s.processing
}
