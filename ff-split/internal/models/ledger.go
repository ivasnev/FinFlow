package models

import (
	"math"
	"sort"
)

// OutstandingDebts сохраняет исходную разбивку без возвратов, иначе строит остатки по чистым балансам.
// Линейное сведение с сортировкой ID закрывает также циклы после возврата непрямому кредитору.
func OutstandingDebts(debts []Debt) []Debt { return netDebts(debts, false) }

// NetDebts сводит неизменяемый архивный ledger без запуска оптимизатора и записи кэша.
func NetDebts(debts []Debt) []Debt { return netDebts(debts, true) }

// netDebts строит остатки в копейках, сохраняя прежнюю разбивку при необходимости.
func netDebts(debts []Debt, force bool) []Debt {
	hasSettlement := force
	balances := map[int64]int64{}
	for _, debt := range debts {
		hasSettlement = hasSettlement || debt.Settlement
		cents := int64(math.Round(debt.Amount * 100))
		balances[debt.FromUserID] -= cents
		balances[debt.ToUserID] += cents
	}
	if !hasSettlement {
		return debts
	}
	var senders, receivers []int64
	for id, balance := range balances {
		if balance < 0 {
			senders = append(senders, id)
		}
		if balance > 0 {
			receivers = append(receivers, id)
		}
	}
	sort.Slice(senders, func(i, j int) bool { return senders[i] < senders[j] })
	sort.Slice(receivers, func(i, j int) bool { return receivers[i] < receivers[j] })
	result := []Debt{}
	for i, j := 0, 0; i < len(senders) && j < len(receivers); {
		from, to := senders[i], receivers[j]
		amount := min(-balances[from], balances[to])
		result = append(result, Debt{FromUserID: from, ToUserID: to, Amount: float64(amount) / 100})
		balances[from] += amount
		balances[to] -= amount
		if balances[from] == 0 {
			i++
		}
		if balances[to] == 0 {
			j++
		}
	}
	return result
}
