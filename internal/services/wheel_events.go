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
	NewCandidateAddedToWheelEventType  = "NewCandidateAddedToWheelEvent"
	CandidateRemovedFromWheelEventType = "CandidateRemovedFromWheelEvent"
	WheelStatusChangedEventType        = "WheelStatusChangedEvent"
	WheelDeletedEventType              = "WheelDeletedEvent"
	WheelAddedEventType                = "WheelAddedEvent"
	WinnerDeclaredEventType            = "WinnerDeclaredEvent"

	globalWheelEventsTopic = "global-wheel-events"
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

	CandidateRemovedFromWheelEvent struct {
		CandidateID int64 `json:"candidate_id"`
	}

	StatusChangedEvent struct {
		WheelID int64  `json:"wheel_id"`
		Status  string `json:"status"`
	}

	WheelDeletedEvent struct {
		WheelID int64 `json:"wheel_id"`
	}

	WheelAddedEvent struct {
		Wheel Wheel `json:"wheel"`
	}

	WinnerDeclaredEvent struct {
		WheelID int64     `json:"wheel_id"`
		Winner  Candidate `json:"winner"`
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
	msg, err := eventToMessage(NewCandidateAddedToWheelEventType, NewCandidateAddedToWheelEvent{
		Candidate: candidate,
	})

	if err != nil {
		return err
	}

	return s.publishWheelEvent(wheelId, NewCandidateAddedToWheelEventType, msg)
}

func (s *WheelEventsService) PublishCandidateRemovedFromWheelEvent(wheelId int64, candidateId int64) error {
	msg, err := eventToMessage(CandidateRemovedFromWheelEventType, CandidateRemovedFromWheelEvent{
		CandidateID: candidateId,
	})

	if err != nil {
		return err
	}

	return s.publishWheelEvent(wheelId, CandidateRemovedFromWheelEventType, msg)
}

func (s *WheelEventsService) PublishWheelStatusChangedEvent(wheelId int64, status WheelStatus) error {
	msg, err := eventToMessage(WheelStatusChangedEventType, StatusChangedEvent{
		WheelID: wheelId,
		Status:  status.String(),
	})

	if err != nil {
		return err
	}

	return s.publishWheelEvent(wheelId, WheelStatusChangedEventType, msg)
}

func (s *WheelEventsService) PublishWinnerDeclaredEvent(wheelId int64, winner Candidate) error {
	msg, err := eventToMessage(WinnerDeclaredEventType, WinnerDeclaredEvent{
		WheelID: wheelId,
		Winner:  winner,
	})

	if err != nil {
		return err
	}

	return s.publishWheelEvent(wheelId, WinnerDeclaredEventType, msg)
}

func (s *WheelEventsService) PublishWheelDeletedEvent(wheelId int64) error {
	msg, err := eventToMessage(WheelDeletedEventType, WheelDeletedEvent{
		WheelID: wheelId,
	})

	if err != nil {
		return err
	}

	if err := s.PublishGlobalWheelEvent(WheelDeletedEventType, msg); err != nil {
		return err
	}

	return s.publishWheelEvent(wheelId, WheelDeletedEventType, msg)
}

func (s *WheelEventsService) PublishWheelAddedEvent(wheel Wheel) error {
	msg, err := eventToMessage(WheelAddedEventType, WheelAddedEvent{
		Wheel: wheel,
	})

	if err != nil {
		return err
	}

	return s.PublishGlobalWheelEvent(WheelAddedEventType, msg)
}

func (s *WheelEventsService) publishWheelEvent(wheelId int64, eventType string, msg *message.Message) error {
	return s.pub.Publish(getWheelEventsTopic(wheelId), msg)
}

func (s *WheelEventsService) PublishGlobalWheelEvent(eventType string, msg *message.Message) error {
	return s.pub.Publish(globalWheelEventsTopic, msg)
}

func (s *WheelEventsService) SubscribeToWheelEvents(ctx context.Context, wheelId int64) (<-chan interface{}, error) {
	return s.subscribeToTopic(ctx, getWheelEventsTopic(wheelId), s.logger.With(zap.Int64("wheel_id", wheelId)))
}

func (s *WheelEventsService) SubscribeToGlobalWheelEvents(ctx context.Context) (<-chan interface{}, error) {
	return s.subscribeToTopic(ctx, globalWheelEventsTopic, s.logger)
}

func (s *WheelEventsService) subscribeToTopic(ctx context.Context, topic string, logger *zap.Logger) (<-chan interface{}, error) {
	messageChannel, err := s.sub.Subscribe(ctx, topic)
	eventChannel := make(chan interface{})
	if err != nil {
		return nil, err
	}

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
					logger.Error("Failed to unmarshal event envelope", zap.Error(err))
					msg.Nack()
					continue
				}

				event, err := interfaceToEvent(envelope.EventType, envelope.Data)
				if err != nil {
					logger.Error("Failed to convert event to interface", zap.Error(err))
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

func eventToMessage(eventType string, event interface{}) (*message.Message, error) {
	innerPayload, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(WheelEventEnvelope{
		EventType: eventType,
		Data:      json.RawMessage(innerPayload),
	})
	if err != nil {
		return nil, err
	}

	return message.NewMessage(watermill.NewUUID(), payload), nil
}

func unmarshalEvent[T any](rawData json.RawMessage) (interface{}, error) {
	var event T
	if err := json.Unmarshal(rawData, &event); err != nil {
		return nil, err
	}
	return event, nil
}

var eventDecoders = map[string]func(json.RawMessage) (interface{}, error){
	NewCandidateAddedToWheelEventType:  unmarshalEvent[NewCandidateAddedToWheelEvent],
	CandidateRemovedFromWheelEventType: unmarshalEvent[CandidateRemovedFromWheelEvent],
	WheelStatusChangedEventType:        unmarshalEvent[StatusChangedEvent],
	WheelDeletedEventType:              unmarshalEvent[WheelDeletedEvent],
	WheelAddedEventType:                unmarshalEvent[WheelAddedEvent],
	WinnerDeclaredEventType:            unmarshalEvent[WinnerDeclaredEvent],
}

func interfaceToEvent(eventType string, rawData json.RawMessage) (interface{}, error) {
	decode, ok := eventDecoders[eventType]
	if !ok {
		return nil, fmt.Errorf("unknown event type: %s", eventType)
	}
	return decode(rawData)
}

func getWheelEventsTopic(wheelId int64) string {
	return "wheel-events-" + strconv.FormatInt(wheelId, 10)
}
