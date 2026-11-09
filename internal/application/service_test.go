package application_test

import (
	"errors"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"policy-workbench/internal/application"
	"policy-workbench/internal/demo"
	"policy-workbench/internal/domain"
	"policy-workbench/internal/storage/jsonfile"
)

func testService(t *testing.T, path string) (*application.Service, *jsonfile.Repository) {
	t.Helper()
	catalog := demo.Catalog()
	seed := func(now time.Time) application.WorkspaceState { return demo.Workspace(catalog, now) }
	repository, err := jsonfile.Open(path, seed(time.Date(2026, 1, 12, 12, 0, 0, 0, time.UTC)))
	if err != nil {
		t.Fatal(err)
	}
	service, err := application.NewService(repository, catalog, seed)
	if err != nil {
		t.Fatal(err)
	}
	return service, repository
}

func transferRequest() domain.TransferRequest {
	return domain.TransferRequest{ProfileID: "treasury", InitiatorID: "alex", WalletID: "operations", RecipientID: "supplier", Amount: "100"}
}

func TestDecisionSurvivesPolicyDeletionAndRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	service, _ := testService(t, path)
	request := transferRequest()
	request.WalletID = "reserve"
	transfer, err := service.CreateTransfer(request)
	if err != nil {
		t.Fatal(err)
	}
	if transfer.Decision.PolicyID != "policy-review" || transfer.Status != domain.StatusPendingSignature {
		t.Fatal("fixture did not select the fallback policy for immediate signing")
	}
	decision := transfer.Decision.Clone()
	if err := service.DeletePolicy("treasury", transfer.Decision.PolicyID); err != nil {
		t.Fatal(err)
	}
	reloaded, _ := testService(t, path)
	for _, action := range []domain.TransferAction{domain.ActionSign, domain.ActionConfirm} {
		transfer, err = reloaded.ApplyTransferAction(transfer.ID, application.TransferActionCommand{ProfileID: "treasury", Action: action})
		if err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(transfer.Decision, decision) {
		t.Fatal("historical decision changed after policy deletion and restart")
	}
}

func TestConcurrentCreatesAndDuplicateSigning(t *testing.T) {
	service, repository := testService(t, filepath.Join(t.TempDir(), "state.json"))
	before := len(repository.Snapshot().Transfers)
	var workers sync.WaitGroup
	for index := 0; index < 16; index++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			if _, err := service.CreateTransfer(transferRequest()); err != nil {
				t.Error(err)
			}
		}()
	}
	workers.Wait()
	state := repository.Snapshot()
	if len(state.Transfers) != before+16 {
		t.Fatal("concurrent creates lost records")
	}
	ids := map[string]bool{}
	for _, transfer := range state.Transfers {
		if ids[transfer.ID] {
			t.Fatal("duplicate ID")
		}
		ids[transfer.ID] = true
	}
	request := transferRequest()
	request.WalletID = "reserve"
	transfer, err := service.CreateTransfer(request)
	if err != nil {
		t.Fatal(err)
	}
	results := make(chan error, 2)
	for index := 0; index < 2; index++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			_, err := service.ApplyTransferAction(transfer.ID, application.TransferActionCommand{ProfileID: "treasury", Action: domain.ActionSign})
			results <- err
		}()
	}
	workers.Wait()
	close(results)
	successes, conflicts := 0, 0
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, domain.ErrActionUnavailable):
			conflicts++
		default:
			t.Fatal(err)
		}
	}
	state = repository.Snapshot()
	latest := state.Transfers[len(state.Transfers)-1]
	if successes != 1 || conflicts != 1 || latest.Status != domain.StatusSent {
		t.Fatal("duplicate signing was not refused")
	}
}

func TestFinalStatesAndProfileScope(t *testing.T) {
	service, _ := testService(t, filepath.Join(t.TempDir(), "state.json"))
	transfer, err := service.CreateTransfer(transferRequest())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ApplyTransferAction(transfer.ID, application.TransferActionCommand{ProfileID: "studio", Action: domain.ActionSign}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("another profile changed transfer")
	}
	if _, err := service.ApplyTransferAction(transfer.ID, application.TransferActionCommand{ProfileID: "treasury", Action: domain.ActionConfirm}); !errors.Is(err, domain.ErrActionUnavailable) {
		t.Fatal("unsigned transfer confirmed")
	}
	for _, action := range []domain.TransferAction{domain.ActionSign, domain.ActionConfirm} {
		if _, err := service.ApplyTransferAction(transfer.ID, application.TransferActionCommand{ProfileID: "treasury", Action: action}); err != nil {
			t.Fatal(err)
		}
	}
	for _, action := range []domain.TransferAction{domain.ActionFail, domain.ActionCancel, domain.ActionSign, domain.ActionReject, domain.ActionConfirm} {
		if _, err := service.ApplyTransferAction(transfer.ID, application.TransferActionCommand{ProfileID: "treasury", Action: action}); !errors.Is(err, domain.ErrActionUnavailable) {
			t.Errorf("final transfer accepted %s", action)
		}
	}
}
