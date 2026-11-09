package domain

import "time"

type TransferStatus string

const (
	StatusPendingSignature TransferStatus = "PENDING_SIGNATURE"
	StatusSent             TransferStatus = "SENT"
	StatusConfirmed        TransferStatus = "CONFIRMED"
	StatusFailed           TransferStatus = "FAILED"
	StatusRejected         TransferStatus = "REJECTED"
	StatusRejectedByPolicy TransferStatus = "REJECTED_BY_POLICY"
	StatusNoPolicyFound    TransferStatus = "NO_POLICY_FOUND"
	StatusCancelled        TransferStatus = "CANCELLED"
)

type TransferAction string

const (
	ActionReject  TransferAction = "reject"
	ActionCancel  TransferAction = "cancel"
	ActionSign    TransferAction = "sign"
	ActionConfirm TransferAction = "confirm"
	ActionFail    TransferAction = "fail"
)

type TransferRequest struct {
	ProfileID   string `json:"profile_id"`
	InitiatorID string `json:"initiator_id"`
	WalletID    string `json:"wallet_id"`
	RecipientID string `json:"recipient_id"`
	Amount      string `json:"amount"` // Exact amount in the source wallet currency.
	Note        string `json:"note"`
}

type Transfer struct {
	ID string `json:"id"`
	TransferRequest
	Currency  Currency       `json:"currency"`
	Status    TransferStatus `json:"status"`
	Decision  PolicyDecision `json:"decision"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

func NewTransfer(id string, request TransferRequest, currency Currency, decision PolicyDecision, now time.Time) Transfer {
	status := StatusPendingSignature
	switch decision.Outcome {
	case OutcomeNoMatch:
		status = StatusNoPolicyFound
	case OutcomeReject:
		status = StatusRejectedByPolicy
	}
	return Transfer{
		ID: id, TransferRequest: request, Currency: currency, Status: status,
		Decision: decision.Clone(), CreatedAt: now, UpdatedAt: now,
	}
}

func (transfer Transfer) Clone() Transfer {
	transfer.Decision = transfer.Decision.Clone()
	return transfer
}

func (transfer *Transfer) ApplyAction(action TransferAction, now time.Time) error {
	switch action {
	case ActionReject:
		if transfer.Status != StatusPendingSignature {
			return ErrActionUnavailable
		}
		transfer.Status = StatusRejected
	case ActionCancel:
		if transfer.Status != StatusPendingSignature {
			return ErrActionUnavailable
		}
		transfer.Status = StatusCancelled
	case ActionSign:
		if transfer.Status != StatusPendingSignature {
			return ErrActionUnavailable
		}
		transfer.Status = StatusSent
	case ActionConfirm:
		if transfer.Status != StatusSent {
			return ErrActionUnavailable
		}
		transfer.Status = StatusConfirmed
	case ActionFail:
		if transfer.Status != StatusPendingSignature && transfer.Status != StatusSent {
			return ErrActionUnavailable
		}
		transfer.Status = StatusFailed
	default:
		return InvalidRequest("unknown transfer action")
	}
	transfer.UpdatedAt = now
	return nil
}
