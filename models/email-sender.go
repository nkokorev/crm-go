package models

type EmailSender interface {

	Entity

	// email_queues, email_campaigns, email_notifications
	GetType() string

	// Возвращает состояние сендера
	IsEnabled() bool
}

type EmailSenderType = string

const (
	EmailSenderQueue 			EmailSenderType = "email_queues"
	EmailSenderCampaign 		EmailSenderType = "email_campaigns"
	EmailSenderNotification 	EmailSenderType = "email_notifications"
)