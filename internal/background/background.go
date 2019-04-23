package background

import (
	"time"

	"gitlab.com/asciishell/tfs-go-auction/internal/storage"
	"gitlab.com/asciishell/tfs-go-auction/pkg/log"
)

type Background struct {
	logger  log.Logger
	storage storage.Storage
}

func (b Background) RunCloseLots() {
	go func() {
		for {
			count, err := b.storage.CloseLots()
			if err != nil {
				b.logger.Errorf("error during closing: %+v", err)
			}
			if count != 0 {
				b.logger.Infof("closed %d lots", count)
			}
			time.Sleep(time.Second)
		}
	}()
}
func NewBackground(logger log.Logger, storage storage.Storage) Background {
	result := Background{logger: logger, storage: storage}
	result.RunCloseLots()
	return result
}
