package optimizers

import (
	"errors"
	"sort"
)

// Users возвращает отсортированный список всех пользователей, участвующих в переводах.
func Users(transfers []Transfer) []string {
	unique := make(map[string]struct{})
	for _, tr := range transfers {
		unique[tr.From] = struct{}{}
		unique[tr.To] = struct{}{}
	}
	users := make([]string, 0, len(unique))
	for u := range unique {
		users = append(users, u)
	}
	sort.Strings(users)
	return users
}

// Balances вычисляет баланс для каждого пользователя: входящие - исходящие.
// Положительный баланс означает, что пользователь - кредитор (ему должны).
// Отрицательный баланс означает, что пользователь - должник.
func Balances(transfers []Transfer) (map[string]int, error) {
	balances := make(map[string]int)
	for _, tr := range transfers {
		if tr.Amount < 0 {
			return nil, ValidationError{Message: "negative transfer amount"}
		}
		if tr.Amount == 0 {
			continue
		}
		balances[tr.From] -= tr.Amount
		balances[tr.To] += tr.Amount
	}
	for _, user := range Users(transfers) {
		if _, ok := balances[user]; !ok {
			balances[user] = 0
		}
	}
	return balances, nil
}

// TransferMatrix агрегирует переводы по парам отправитель->получатель.
func TransferMatrix(transfers []Transfer) map[string]map[string]int {
	matrix := make(map[string]map[string]int)
	for _, tr := range transfers {
		if tr.Amount <= 0 {
			continue
		}
		if _, ok := matrix[tr.From]; !ok {
			matrix[tr.From] = make(map[string]int)
		}
		matrix[tr.From][tr.To] += tr.Amount
	}
	return matrix
}

// CollapseTransfers схлопывает долговые обязательства
func CollapseTransfers(transfers []Transfer) ([]Transfer, error) {
	net := make(map[string]map[string]int)
	for _, tr := range transfers {
		if tr.Amount <= 0 {
			return nil, errors.New("negative transfer amount")
		}
		if net[tr.From] == nil {
			net[tr.From] = make(map[string]int)
		}
		if net[tr.To] == nil {
			net[tr.To] = make(map[string]int)
		}
		net[tr.From][tr.To] += tr.Amount
		net[tr.To][tr.From] -= tr.Amount
	}

	var out []Transfer
	for from, toMap := range net {
		for to, amount := range toMap {
			if from == to {
				continue
			}
			if amount <= 0 {
				continue
			}
			out = append(out, Transfer{From: from, To: to, Amount: amount})
		}
	}
	return out, nil
}
