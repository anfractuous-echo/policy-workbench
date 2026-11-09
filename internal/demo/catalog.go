package demo

import "policy-workbench/internal/domain"

func Catalog() domain.Catalog {
	return domain.Catalog{
		Profiles: []domain.Profile{{ID: "treasury", Name: "Northstar Treasury"}, {ID: "studio", Name: "Studio Sandbox"}},
		Members: []domain.ProfileMember{
			{ID: "alex", ProfileID: "treasury", Name: "Alex Morgan"},
			{ID: "sam", ProfileID: "treasury", Name: "Sam Lee"},
			{ID: "jordan", ProfileID: "treasury", Name: "Jordan Park"},
			{ID: "casey", ProfileID: "studio", Name: "Casey Brooks"},
			{ID: "riley", ProfileID: "studio", Name: "Riley Chen"},
		},
		Wallets: []domain.Wallet{
			{ID: "operations", ProfileID: "treasury", Name: "Operations", Currency: domain.CurrencyUSD, OpeningBalance: "100000"},
			{ID: "reserve", ProfileID: "treasury", Name: "Reserve", Currency: domain.CurrencyUSD, OpeningBalance: "250000"},
			{ID: "europe", ProfileID: "treasury", Name: "Europe", Currency: domain.CurrencyEUR, OpeningBalance: "80000"},
			{ID: "studio-wallet", ProfileID: "studio", Name: "Studio USD", Currency: domain.CurrencyUSD, OpeningBalance: "20000"},
			{ID: "studio-eur", ProfileID: "studio", Name: "Studio EUR", Currency: domain.CurrencyEUR, OpeningBalance: "15000"},
		},
		WalletGroups: []domain.Group{
			{ID: "daily-wallets", ProfileID: "treasury", Name: "Daily operations", Members: []string{"operations"}},
			{ID: "all-treasury", ProfileID: "treasury", Name: "All treasury wallets", Members: []string{"operations", "reserve", "europe"}},
			{ID: "studio-wallets", ProfileID: "studio", Name: "Studio wallets", Members: []string{"studio-wallet", "studio-eur"}},
		},
		Recipients: []domain.Recipient{
			{ID: "supplier", ProfileID: "treasury", Name: "Supplier"},
			{ID: "partner", ProfileID: "treasury", Name: "Partner"},
			{ID: "restricted", ProfileID: "treasury", Name: "Restricted recipient"},
			{ID: "studio-supplier", ProfileID: "studio", Name: "Studio supplier"},
		},
		RecipientGroups: []domain.Group{
			{ID: "suppliers", ProfileID: "treasury", Name: "Suppliers", Members: []string{"supplier", "partner"}},
			{ID: "restricted", ProfileID: "treasury", Name: "Restricted", Members: []string{"restricted"}},
			{ID: "studio-suppliers", ProfileID: "studio", Name: "Studio suppliers", Members: []string{"studio-supplier"}},
		},
	}
}
