package domain

import "time"

type PolicyAction string

const (
	PolicyAllow  PolicyAction = "ALLOW"
	PolicyReject PolicyAction = "REJECT"
)

type SenderType string

const (
	SenderAny         SenderType = "any"
	SenderWallet      SenderType = "wallet"
	SenderWalletGroup SenderType = "wallet_group"
)

type ReceiverType string

const (
	ReceiverAny            ReceiverType = "any"
	ReceiverRecipient      ReceiverType = "recipient"
	ReceiverRecipientGroup ReceiverType = "recipient_group"
)

type Policy struct {
	ID           string       `json:"id"`
	ProfileID    string       `json:"profile_id"`
	Name         string       `json:"name"`
	InitiatorID  string       `json:"initiator_id"`
	SenderType   SenderType   `json:"sender_type"`
	SenderID     string       `json:"sender_id"`
	ReceiverType ReceiverType `json:"receiver_type"`
	ReceiverID   string       `json:"receiver_id"`
	MaxAmount    string       `json:"amount_limit"` // Inclusive upper bound; empty means unbounded.
	Currency     Currency     `json:"currency"`
	Action       PolicyAction `json:"action"`
	CreatedAt    time.Time    `json:"created_at"`
}
