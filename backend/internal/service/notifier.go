package service

import (
	"fmt"
	"time"

	"goal-manager/backend/internal/domain"
)

type NotificationChannel string

const (
	ChannelEmail  NotificationChannel = "email"
	ChannelFeishu NotificationChannel = "feishu"
)

type NotificationAdapter interface {
	Send(event domain.NotificationEvent) error
	Channel() NotificationChannel
}

type EmailAdapter struct{}

func (a *EmailAdapter) Send(event domain.NotificationEvent) error {
	return nil
}

func (a *EmailAdapter) Channel() NotificationChannel {
	return ChannelEmail
}

type FeishuAdapter struct{}

func (a *FeishuAdapter) Send(event domain.NotificationEvent) error {
	return nil
}

func (a *FeishuAdapter) Channel() NotificationChannel {
	return ChannelFeishu
}

type Notifier struct {
	adapters map[NotificationChannel]NotificationAdapter
	store    *Store
}

func NewNotifier(store *Store) *Notifier {
	return &Notifier{
		adapters: map[NotificationChannel]NotificationAdapter{
			ChannelEmail:  &EmailAdapter{},
			ChannelFeishu: &FeishuAdapter{},
		},
		store: store,
	}
}

func (n *Notifier) Notify(eventType, target, title, message string, channels ...NotificationChannel) []domain.NotificationEvent {
	if len(channels) == 0 {
		channels = []NotificationChannel{ChannelEmail, ChannelFeishu}
	}
	var events []domain.NotificationEvent
	for _, ch := range channels {
		adapter, ok := n.adapters[ch]
		if !ok {
			continue
		}
		now := time.Now().UTC()
		event := domain.NotificationEvent{
			EventType:  eventType,
			Target:     target,
			Channel:    string(ch),
			Title:      title,
			Message:    message,
			Delivered:  false,
			AuditFields: domain.AuditFields{CreatedAt: now, UpdatedAt: now},
		}
		err := adapter.Send(event)
		event.Delivered = err == nil
		if err != nil {
			event.Message = fmt.Sprintf("%s (delivery failed: %v)", message, err)
		}
		saved := n.store.SaveNotification(event)
		events = append(events, saved)
	}
	return events
}

func (n *Notifier) NotifyAlert(alert domain.Alert, channels ...NotificationChannel) []domain.NotificationEvent {
	return n.Notify(
		alert.AlertType,
		alert.TargetID,
		alert.AlertType,
		alert.Message,
		channels...,
	)
}
