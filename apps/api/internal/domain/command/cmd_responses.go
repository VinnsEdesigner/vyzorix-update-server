package command

// CommandResponse is the server's response to a command dispatch.
type CommandResponse struct {
	DispatchID string `json:"dispatchId"`
	Delivery   string `json:"delivery"`
	ServerTime int64  `json:"serverTime"`
}

// CommandDeliveryStatus represents the delivery status of a command.
type CommandDeliveryStatus string

const (
	DeliveryPending   CommandDeliveryStatus = "pending"
	DeliveryDelivered CommandDeliveryStatus = "delivered"
	DeliveryFailed    CommandDeliveryStatus = "failed"
	DeliveryTimeout   CommandDeliveryStatus = "timeout"
)
