package domain

import "unicode/utf8"

func ValidateTransfer(request TransferRequest, catalog Catalog) error {
	if !catalog.HasProfile(request.ProfileID) || !catalog.HasMember(request.ProfileID, request.InitiatorID) {
		return InvalidRequest("choose an initiator in this profile")
	}
	if _, ok := catalog.FindWallet(request.ProfileID, request.WalletID); !ok {
		return InvalidRequest("choose a source wallet in this profile")
	}
	if _, ok := catalog.FindRecipient(request.ProfileID, request.RecipientID); !ok {
		return InvalidRequest("choose a recipient in this profile")
	}
	amount, err := ParseMoney(request.Amount)
	if err != nil || amount.coefficient.Sign() <= 0 {
		return InvalidRequest("amount must be positive with at most 2 decimal places")
	}
	if utf8.RuneCountInString(request.Note) > 200 {
		return InvalidRequest("note must contain at most 200 characters")
	}
	return nil
}
