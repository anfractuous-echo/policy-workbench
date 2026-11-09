package application

import "policy-workbench/internal/domain"

func (service *Service) CreateTransfer(request domain.TransferRequest) (domain.Transfer, error) {
	var transfer domain.Transfer
	err := service.repository.Update(func(state *WorkspaceState) error {
		decision, err := domain.EvaluatePolicies(state.Policies, service.catalog, request)
		if err != nil {
			return err
		}
		wallet, _ := service.catalog.FindWallet(request.ProfileID, request.WalletID)
		if decision.Outcome == domain.OutcomeAllow {
			balance, err := domain.CalculateWalletBalance(wallet, state.Transfers)
			if err != nil {
				return err
			}
			available, err := domain.ParseMoney(balance.Available)
			if err != nil {
				return err
			}
			amount, err := domain.ParseMoney(request.Amount)
			if err != nil {
				return err
			}
			if amount.Compare(available) > 0 {
				return domain.InvalidRequest("insufficient %s balance: available %s", wallet.Currency, balance.Available)
			}
		}
		transfer = domain.NewTransfer(state.AllocateID("tx"), request, wallet.Currency, decision, state.SimulationTime)
		state.Transfers = append(state.Transfers, transfer)
		return nil
	})
	return transfer, err
}

func (service *Service) ApplyTransferAction(transferID string, command TransferActionCommand) (domain.Transfer, error) {
	var result domain.Transfer
	err := service.repository.Update(func(state *WorkspaceState) error {
		for index := range state.Transfers {
			transfer := &state.Transfers[index]
			if transfer.ID == transferID && transfer.ProfileID == command.ProfileID {
				if err := transfer.ApplyAction(command.Action, state.SimulationTime); err != nil {
					return err
				}
				result = transfer.Clone()
				return nil
			}
		}
		return domain.ErrNotFound
	})
	return result, err
}
