package pushrelabel

import (
	"fmt"

	"github.com/ivasnev/FinFlow/ff-common/optimizers"
	"github.com/ivasnev/FinFlow/ff-common/optimizers/maxflow"
	mfpr "github.com/ivasnev/FinFlow/ff-common/optimizers/maxflow/pushrelabel"
	"github.com/ivasnev/FinFlow/ff-common/optimizers/utils"
)

// Optimizer реализует оптимизацию долгов на основе алгоритма Черкасского (push-relabel с highest label).
// Сложность: O(V·E + V²√E).
type Optimizer struct {
	maxFlowSolver maxflow.MaxFlowSolver
}

// New создаёт новый экземпляр оптимизатора на основе алгоритма Черкасского.
func New() *Optimizer {
	return &Optimizer{maxFlowSolver: mfpr.Solver{}}
}

func (o *Optimizer) Optimize(debts []optimizers.Transfer) ([]optimizers.Transfer, error) {
	if len(debts) == 0 {
		return nil, nil
	}

	debts, err := utils.CollapseTransfers(debts)
	if err != nil {
		return nil, err
	}

	balances, err := utils.Balances(debts)
	if err != nil {
		return nil, err
	}

	users := utils.Users(debts)
	index := make(map[string]int, len(users))
	for i, u := range users {
		index[u] = i
	}

	originalOut := make(map[string]int, len(users))
	for _, tr := range debts {
		if tr.Amount > 0 {
			originalOut[tr.From] += tr.Amount
		}
	}

	totalDemand := 0
	for _, user := range users {
		if balances[user] > 0 {
			totalDemand += balances[user]
		}
	}

	if totalDemand == 0 {
		return nil, nil
	}

	n := len(users)
	source := 2 * n
	sink := 2*n + 1
	graph := maxflow.NewGraph(2*n + 2)

	for i, user := range users {
		in := i
		out := i + n
		cap := originalOut[user]
		if cap > 0 {
			graph.AddEdge(in, out, cap)
		}
	}

	for i, user := range users {
		in := i
		balance := balances[user]
		if balance < 0 {
			graph.AddEdge(source, in, -balance)
		} else if balance > 0 {
			graph.AddEdge(in, sink, balance)
		}
	}

	allowed := utils.TransferMatrix(debts)
	infCap := totalDemand
	for from, toMap := range allowed {
		fromIdx := index[from]
		fromOut := fromIdx + n
		for to := range toMap {
			toIdx := index[to]
			if fromIdx == toIdx {
				continue
			}
			toIn := toIdx
			graph.AddEdge(fromOut, toIn, infCap)
		}
	}

	maxFlow := o.maxFlowSolver.GetMaxFlow(graph, source, sink)
	if maxFlow != totalDemand {
		return nil, fmt.Errorf("insufficient max flow: %d of %d", maxFlow, totalDemand)
	}

	return utils.TransfersFromFlowGraph(graph, n, users), nil
}
