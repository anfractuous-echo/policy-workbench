package domain

import "slices"

type Currency string

const (
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
)

func (currency Currency) Valid() bool { return currency == CurrencyUSD || currency == CurrencyEUR }

type Profile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ProfileMember struct {
	ID        string `json:"id"`
	ProfileID string `json:"profile_id"`
	Name      string `json:"name"`
}

type Wallet struct {
	ID             string   `json:"id"`
	ProfileID      string   `json:"profile_id"`
	Name           string   `json:"name"`
	Currency       Currency `json:"currency"`
	OpeningBalance string   `json:"opening_balance"`
}

type Group struct {
	ID        string   `json:"id"`
	ProfileID string   `json:"profile_id"`
	Name      string   `json:"name"`
	Members   []string `json:"members"`
}

type Recipient struct {
	ID        string `json:"id"`
	ProfileID string `json:"profile_id"`
	Name      string `json:"name"`
}

type Catalog struct {
	Profiles        []Profile       `json:"profiles"`
	Members         []ProfileMember `json:"people"`
	Wallets         []Wallet        `json:"wallets"`
	WalletGroups    []Group         `json:"wallet_groups"`
	Recipients      []Recipient     `json:"recipients"`
	RecipientGroups []Group         `json:"recipient_groups"`
}

func (c Catalog) HasProfile(id string) bool {
	return slices.ContainsFunc(c.Profiles, func(p Profile) bool { return p.ID == id })
}
func (c Catalog) HasMember(profile, id string) bool {
	return slices.ContainsFunc(c.Members, func(p ProfileMember) bool { return p.ProfileID == profile && p.ID == id })
}
func (c Catalog) FindWallet(profile, id string) (Wallet, bool) {
	for _, wallet := range c.Wallets {
		if wallet.ProfileID == profile && wallet.ID == id {
			return wallet, true
		}
	}
	return Wallet{}, false
}
func (c Catalog) FindRecipient(profile, id string) (Recipient, bool) {
	for _, recipient := range c.Recipients {
		if recipient.ProfileID == profile && recipient.ID == id {
			return recipient, true
		}
	}
	return Recipient{}, false
}
func hasGroup(groups []Group, profile, id string) bool {
	return slices.ContainsFunc(groups, func(group Group) bool { return group.ProfileID == profile && group.ID == id })
}
func groupContains(groups []Group, profile, id, member string) bool {
	for _, group := range groups {
		if group.ProfileID == profile && group.ID == id {
			return slices.Contains(group.Members, member)
		}
	}
	return false
}
