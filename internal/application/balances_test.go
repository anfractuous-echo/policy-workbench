package application_test

import (
	"errors"
	"path/filepath"
	"sync"
	"testing"

	"policy-workbench/internal/application"
	"policy-workbench/internal/domain"
)

func walletBalance(t *testing.T, service *application.Service, id string) domain.WalletBalance {
	t.Helper()
	view, err := service.Workspace("treasury")
	if err != nil {
		t.Fatal(err)
	}
	for _, wallet := range view.Wallets {
		if wallet.ID == id {
			return wallet
		}
	}
	t.Fatal("wallet missing", id)
	return domain.WalletBalance{}
}
func TestBalancesFollowTransferStatesAndRestart(t *testing.T) {
	for _, tc := range []struct {
		action                       domain.TransferAction
		balance, reserved, available string
	}{
		{domain.ActionCancel, "94000", "0", "94000"},
		{domain.ActionReject, "94000", "0", "94000"},
		{domain.ActionFail, "94000", "0", "94000"},
		{domain.ActionSign, "94000", "100", "93900"},
	} {
		t.Run(string(tc.action), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "state.json")
			service, _ := testService(t, path)
			transfer, err := service.CreateTransfer(transferRequest())
			if err != nil {
				t.Fatal(err)
			}
			if got := walletBalance(t, service, "operations"); got.Available != "93900" || got.Reserved != "100" {
				t.Fatal("creation did not reserve funds", got)
			}
			if _, err = service.ApplyTransferAction(transfer.ID, application.TransferActionCommand{ProfileID: "treasury", Action: tc.action}); err != nil {
				t.Fatal(err)
			}
			service, _ = testService(t, path)
			got := walletBalance(t, service, "operations")
			if got.Balance != tc.balance || got.Reserved != tc.reserved || got.Available != tc.available {
				t.Fatal("unexpected balance", got)
			}
			if tc.action == domain.ActionSign {
				if _, err = service.ApplyTransferAction(transfer.ID, application.TransferActionCommand{ProfileID: "treasury", Action: domain.ActionConfirm}); err != nil {
					t.Fatal(err)
				}
				got = walletBalance(t, service, "operations")
				if got.Balance != "93900" || got.Available != "93900" || got.Reserved != "0" {
					t.Fatal("confirmation double-counted funds", got)
				}
			}
		})
	}
}
func TestConcurrentRequestsCannotOverspend(t *testing.T) {
	service, repository := testService(t, filepath.Join(t.TempDir(), "state.json"))
	request := transferRequest()
	request.Amount = "60000"
	before := len(repository.Snapshot().Transfers)
	results := make(chan error, 2)
	var workers sync.WaitGroup
	for i := 0; i < 2; i++ {
		workers.Add(1)
		go func() { defer workers.Done(); _, err := service.CreateTransfer(request); results <- err }()
	}
	workers.Wait()
	close(results)
	accepted, rejected := 0, 0
	for err := range results {
		if err == nil {
			accepted++
		} else if errors.Is(err, domain.ErrInvalidRequest) {
			rejected++
		} else {
			t.Fatal(err)
		}
	}
	if accepted != 1 || rejected != 1 || len(repository.Snapshot().Transfers) != before+1 {
		t.Fatal("concurrent overspend", accepted, rejected)
	}
	if got := walletBalance(t, service, "operations"); got.Available != "34000" {
		t.Fatal(got)
	}
}
func TestPolicyRejectionDoesNotReserveFunds(t *testing.T) {
	service, _ := testService(t, filepath.Join(t.TempDir(), "state.json"))
	request := transferRequest()
	request.RecipientID = "restricted"
	request.Amount = "999999"
	transfer, err := service.CreateTransfer(request)
	if err != nil || transfer.Status != domain.StatusRejectedByPolicy {
		t.Fatal("policy rejection lost", transfer, err)
	}
	if got := walletBalance(t, service, "operations"); got.Available != "94000" {
		t.Fatal("rejection reserved funds", got)
	}
}
