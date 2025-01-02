package enums

type SubscriptionAction string

const (
	LogsSubscribe   SubscriptionAction = "logsSubscribe"
	LogsUnsubscribe SubscriptionAction = "logsUnsubscribe"

	ProgramSubscribe   SubscriptionAction = "programSubscribe"
	ProgramUnsubscribe SubscriptionAction = "programUnsubscribe"
)

func IsSubscribe(action SubscriptionAction) bool {
	return action == LogsSubscribe || action == ProgramSubscribe
}

func IsUnsubscribe(action SubscriptionAction) bool {
	return action == LogsUnsubscribe || action == ProgramUnsubscribe
}
