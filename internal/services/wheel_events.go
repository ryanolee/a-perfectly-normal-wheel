package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"go.uber.org/zap"
)

const (
	NewCandidateAddedToWheelEventType = "NewCandidateAddedToWheelEvent"
)

type (
	WheelEventsService struct {
		pub    message.Publisher
		sub    message.Subscriber
		logger *zap.Logger
	}

	WheelEventEnvelope struct {
		EventType string          `json:"event_type"`
		Data      json.RawMessage `json:"data"`
	}

	NewCandidateAddedToWheelEvent struct {
		Candidate Candidate `json:"candidate"`
	}
)

func NewWheelEventsService(logger *zap.Logger, watermillLogger watermill.LoggerAdapter) *WheelEventsService {
	ps := gochannel.NewGoChannel(gochannel.Config{}, watermillLogger)
	return &WheelEventsService{
		pub:    ps,
		sub:    ps,
		logger: logger,
	}
}

func (s *WheelEventsService) PublishNewCandidateAddedToWheelEvent(wheelId int64, candidate Candidate) error {
	innerPayload, err := json.Marshal(NewCandidateAddedToWheelEvent{
		Candidate: candidate,
	})
	if err != nil {
		return err
	}

	payload, err := json.Marshal(WheelEventEnvelope{
		EventType: NewCandidateAddedToWheelEventType,
		Data:      json.RawMessage(innerPayload),
	})
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	return s.pub.Publish(getWheelEventsTopic(wheelId), msg)
}

func (s *WheelEventsService) SubscribeToWheelEvents(ctx context.Context, wheelId int64) (<-chan interface{}, error) {
	messageChannel, err := s.sub.Subscribe(ctx, getWheelEventsTopic(wheelId))
	eventChannel := make(chan interface{})
	if err != nil {
		return nil, err
	}

	wheelLogger := s.logger.With(zap.Int64("wheel_id", wheelId))

	go func() {
		defer close(eventChannel)

		for {
			select {
			case msg, ok := <-messageChannel:
				if !ok {
					return
				}

				var envelope WheelEventEnvelope
				if err := json.Unmarshal(msg.Payload, &envelope); err != nil {
					wheelLogger.Error("Failed to unmarshal event envelope", zap.Int64("wheel_id", wheelId), zap.Error(err))
					msg.Nack()
					continue
				}

				event, err := interfaceToEvent(envelope.EventType, envelope.Data)
				if err != nil {
					wheelLogger.Error("Failed to convert event to interface", zap.Int64("wheel_id", wheelId), zap.Error(err))
					msg.Nack()
					continue
				}

				msg.Ack()
				eventChannel <- event
			case <-ctx.Done():
				return
			}
		}
	}()

	return eventChannel, nil
}

func interfaceToEvent(eventType string, rawData json.RawMessage) (interface{}, error) {
	switch eventType {
	case NewCandidateAddedToWheelEventType:
		var event NewCandidateAddedToWheelEvent
		if err := json.Unmarshal(rawData, &event); err != nil {
			return nil, err
		}
		return event, nil
	default:
		return nil, fmt.Errorf("unknown event type: %s", eventType)
	}
}

func getWheelEventsTopic(wheelId int64) string {
	return "wheel-events-" + strconv.FormatInt(wheelId, 10)
}
