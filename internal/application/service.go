package application

import (
	"fmt"
	"slices"
	"time"

	"policy-workbench/internal/domain"
)

type Service struct {
	repository    Repository
	catalog       domain.Catalog
	seedWorkspace func(time.Time) WorkspaceState
}

func NewService(repository Repository, catalog domain.Catalog, seedWorkspace func(time.Time) WorkspaceState) (*Service, error) {
	state := repository.Snapshot()
	for _, wallet := range catalog.Wallets {
		if !wallet.Currency.Valid() {
			return nil, fmt.Errorf("invalid wallet currency: %s", wallet.ID)
		}
		if _, err := domain.CalculateWalletBalance(wallet, state.Transfers); err != nil {
			return nil, err
		}
	}
	for _, policy := range state.Policies {
		if err := domain.ValidatePolicy(policy, catalog); err != nil {
			return nil, fmt.Errorf("invalid saved policy %s: %w", policy.ID, err)
		}
	}
	return &Service{repository: repository, catalog: catalog, seedWorkspace: seedWorkspace}, nil
}

func (service *Service) Workspace(profileID string) (WorkspaceView, error) {
	if !service.catalog.HasProfile(profileID) {
		return WorkspaceView{}, domain.InvalidRequest("choose a known profile")
	}
	state := service.repository.Snapshot()
	view := WorkspaceView{SimulationTime: state.SimulationTime, ProfileID: profileID, Catalog: service.catalog, Wallets: []domain.WalletBalance{}, Policies: []domain.Policy{}, Transfers: []domain.Transfer{}}
	for _, wallet := range service.catalog.Wallets {
		if wallet.ProfileID != profileID {
			continue
		}
		balance, err := domain.CalculateWalletBalance(wallet, state.Transfers)
		if err != nil {
			return WorkspaceView{}, err
		}
		view.Wallets = append(view.Wallets, balance)
	}
	for _, policy := range state.Policies {
		if policy.ProfileID == profileID {
			view.Policies = append(view.Policies, policy)
		}
	}
	for _, transfer := range state.Transfers {
		if transfer.ProfileID == profileID {
			view.Transfers = append(view.Transfers, transfer)
		}
	}
	slices.Reverse(view.Transfers)
	return view, nil
}
