package domain

import "slices"

func EvaluatePolicies(policies []Policy, catalog Catalog, request TransferRequest) (PolicyDecision, error) {
	if err := ValidateTransfer(request, catalog); err != nil {
		return PolicyDecision{}, err
	}
	wallet, _ := catalog.FindWallet(request.ProfileID, request.WalletID)
	amount, err := ParseMoney(request.Amount)
	if err != nil {
		return PolicyDecision{}, err
	}
	decision := PolicyDecision{Outcome: OutcomeNoMatch, Trace: []RuleEvaluation{}}
	for _, policy := range policies {
		if policy.ProfileID != request.ProfileID {
			continue
		}
		sender := policy.SenderType == SenderAny || policy.SenderType == SenderWallet && policy.SenderID == request.WalletID ||
			policy.SenderType == SenderWalletGroup && groupContains(catalog.WalletGroups, request.ProfileID, policy.SenderID, request.WalletID)
		receiver := policy.ReceiverType == ReceiverAny || policy.ReceiverType == ReceiverRecipient && policy.ReceiverID == request.RecipientID ||
			policy.ReceiverType == ReceiverRecipientGroup && groupContains(catalog.RecipientGroups, request.ProfileID, policy.ReceiverID, request.RecipientID)
		checks := []ConditionResult{
			{"Initiator", policy.InitiatorID == "" || policy.InitiatorID == request.InitiatorID, "Any or selected profile member"},
			{"Sender", sender, "Wallet or wallet-group membership"},
			{"Receiver", receiver, "Recipient or recipient-group membership"},
			{"Currency", policy.Currency == wallet.Currency, string(wallet.Currency) + " compared with " + string(policy.Currency)},
		}
		if policy.MaxAmount == "" {
			checks = append(checks, ConditionResult{"Amount", true, "No amount restriction"})
		} else {
			maximum, err := ParseMoney(policy.MaxAmount)
			if err != nil {
				return PolicyDecision{}, InvalidRequest("amount limit: %v", err)
			}
			checks = append(checks, ConditionResult{"Amount", amount.Compare(maximum) <= 0, amount.String() + " <= " + policy.MaxAmount})
		}
		matched := !slices.ContainsFunc(checks, func(check ConditionResult) bool { return !check.Matched })
		decision.Trace = append(decision.Trace, RuleEvaluation{policy.ID, policy.Name, matched, checks})
		if matched {
			decision.PolicyID, decision.PolicyName, decision.Outcome = policy.ID, policy.Name, DecisionOutcome(policy.Action)
			matchedPolicy := policy
			decision.MatchedPolicy = &matchedPolicy
			return decision, nil
		}
	}
	return decision, nil
}
