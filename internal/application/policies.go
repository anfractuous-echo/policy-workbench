package application

import (
	"strings"

	"policy-workbench/internal/domain"
)

func (service *Service) CreatePolicy(policy domain.Policy) (domain.Policy, error) {
	policy.Name = strings.TrimSpace(policy.Name)
	if err := domain.ValidatePolicy(policy, service.catalog); err != nil {
		return domain.Policy{}, err
	}
	err := service.repository.Update(func(state *WorkspaceState) error {
		policy.ID, policy.CreatedAt = state.AllocateID("policy"), state.SimulationTime
		state.Policies = append(state.Policies, policy)
		return nil
	})
	return policy, err
}

func (service *Service) DeletePolicy(profileID, policyID string) error {
	return service.repository.Update(func(state *WorkspaceState) error {
		for index, policy := range state.Policies {
			if policy.ID == policyID && policy.ProfileID == profileID {
				state.Policies = append(state.Policies[:index], state.Policies[index+1:]...)
				return nil
			}
		}
		return domain.ErrNotFound
	})
}

func (service *Service) ReorderPolicies(command ReorderPoliciesCommand) error {
	return service.repository.Update(func(state *WorkspaceState) error {
		if !service.catalog.HasProfile(command.ProfileID) {
			return domain.InvalidRequest("unknown profile")
		}
		byID := map[string]domain.Policy{}
		for _, policy := range state.Policies {
			if policy.ProfileID == command.ProfileID {
				byID[policy.ID] = policy
			}
		}
		if len(command.IDs) != len(byID) {
			return domain.InvalidRequest("order must contain every policy in this profile exactly once; refresh and retry")
		}
		ordered := make([]domain.Policy, 0, len(command.IDs))
		for _, id := range command.IDs {
			policy, exists := byID[id]
			if !exists {
				return domain.InvalidRequest("unknown or repeated policy ID; refresh and retry")
			}
			ordered = append(ordered, policy)
			delete(byID, id)
		}
		position := 0
		for index := range state.Policies {
			if state.Policies[index].ProfileID == command.ProfileID {
				state.Policies[index] = ordered[position]
				position++
			}
		}
		return nil
	})
}
