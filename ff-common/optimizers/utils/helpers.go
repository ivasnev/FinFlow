package utils

import (
	"errors"
	"sort"

	"github.com/ivasnev/FinFlow/ff-common/optimizers"
	"github.com/ivasnev/FinFlow/ff-common/optimizers/maxflow"
)

// Users возвращает отсортированный список всех пользователей, участвующих в переводах.
func Users(transfers []optimizers.Transfer) []string {
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
func Balances(transfers []optimizers.Transfer) (map[string]int, error) {
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
func TransferMatrix(transfers []optimizers.Transfer) map[string]map[string]int {
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
func CollapseTransfers(transfers []optimizers.Transfer) ([]optimizers.Transfer, error) {
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

	var out []optimizers.Transfer
	for from, toMap := range net {
		for to, amount := range toMap {
			if from == to {
				continue
			}
			if amount <= 0 {
				continue
			}
			out = append(out, optimizers.Transfer{From: from, To: to, Amount: amount})
		}
	}
	return out, nil
}

// ResidualGraph строит остаточную сеть по графу g с уже найденным потоком.
// Для каждого ребра (u,v) с пропускной способностью cap и потоком flow:
// в остаточном графе добавляется ребро (u,v) с пропускной способностью cap-flow (если > 0)
// и ребро (v,u) с пропускной способностью flow (если > 0).
// Возвращается новый граф; g не изменяется.
func ResidualGraph(g *maxflow.Graph) *maxflow.Graph {
	n := g.NumNodes()
	res := maxflow.NewGraph(n)
	for from := 0; from < n; from++ {
		for _, e := range g.OutEdges(from) {
			if e.Cap-e.Flow > 0 {
				res.AddEdge(from, e.To, e.Cap-e.Flow)
			}
			if e.Flow > 0 {
				res.AddEdge(e.To, from, e.Flow)
			}
		}
	}
	return res
}

// TransfersFromFlowGraph по графу потока (схема node splitting: 0..n-1 in, n..2n-1 out)
// и списку users строит список переводов по рёбрам fromOut -> toIn с положительным потоком.
func TransfersFromFlowGraph(g *maxflow.Graph, n int, users []string) []optimizers.Transfer {
	var out []optimizers.Transfer
	for fromIdx := 0; fromIdx < n; fromIdx++ {
		fromOut := fromIdx + n
		for _, e := range g.OutEdges(fromOut) {
			if e.To < 0 || e.To >= n || e.Flow <= 0 {
				continue
			}
			out = append(out, optimizers.Transfer{
				From:   users[fromIdx],
				To:     users[e.To],
				Amount: e.Flow,
			})
		}
	}
	return out
}
