package optimizers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCollapseTransfers(t *testing.T) {
	testcases := []struct {
		name               string
		transfers          []Transfer
		optimizedTransfers []Transfer
	}{
		{
			name: "reverse transfer",
			transfers: []Transfer{
				{From: "a", To: "b", Amount: 10},
				{From: "a", To: "b", Amount: 20},
				{From: "b", To: "a", Amount: 40},
			},
			optimizedTransfers: []Transfer{
				{From: "b", To: "a", Amount: 10},
			},
		},
		{
			name: "zero net",
			transfers: []Transfer{
				{From: "a", To: "b", Amount: 10},
				{From: "a", To: "b", Amount: 20},
				{From: "b", To: "a", Amount: 30},
			},
			optimizedTransfers: nil,
		},
		{
			name: "simple",
			transfers: []Transfer{
				{From: "a", To: "b", Amount: 10},
				{From: "a", To: "b", Amount: 20},
				{From: "b", To: "a", Amount: 20},
			},
			optimizedTransfers: []Transfer{
				{From: "a", To: "b", Amount: 10},
			},
		},
		{
			name:               "empty",
			transfers:          nil,
			optimizedTransfers: nil,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := CollapseTransfers(tc.transfers)
			require.NoError(t, err)
			require.Equal(t, tc.optimizedTransfers, got)
		})
	}
}
