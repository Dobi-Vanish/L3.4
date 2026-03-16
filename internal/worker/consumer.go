package worker

import (
	"L3.4/internal/service"
	"context"
	"github.com/segmentio/kafka-go"
	"github.com/wb-go/wbf/logger"
)

type KafkaConsumer struct {
	reader *kafka.Reader
	svc    *service.ImageService
	log    logger.Logger
}

func NewKafkaConsumer(reader *kafka.Reader, svc *service.ImageService, log logger.Logger) *KafkaConsumer {
	return &KafkaConsumer{
		reader: reader,
		svc:    svc,
		log:    log,
	}
}

func (kc *KafkaConsumer) Run(ctx context.Context) error {
	for {
		msg, err := kc.reader.ReadMessage(ctx)
		if err != nil {
			kc.log.Error("failed to read message", "error", err)
			return err
		}
		imageID := string(msg.Value)
		kc.log.Info("processing image", "id", imageID)

		img, err := kc.svc.GetImage(ctx, imageID)
		if err != nil || img == nil {
			kc.log.Error("image not found", "id", imageID, "error", err)
			continue
		}
		if img.OriginalPath == "" {
			kc.log.Error("original path is empty", "id", imageID)
			continue
		}

		if err := kc.svc.ProcessImage(ctx, imageID); err != nil {
			kc.log.Error("failed to process image", "id", imageID, "error", err)
		}
	}
}
