package dinic

import (
	"context"
	"fmt"

	"github.com/ivasnev/FinFlow/ff-common/optimizers"
	"github.com/ivasnev/FinFlow/ff-common/optimizers/maxflow"
	"github.com/ivasnev/FinFlow/ff-common/optimizers/maxflow/dinic"
	"github.com/ivasnev/FinFlow/ff-common/optimizers/utils"
)

// Optimizer реализует алгоритм Диница для оптимизации долгов.
// Сложность: O(V^2 * E), где V - количество пользователей, E - количество долговых связей.
type Optimizer struct {
	maxFlowSolver dinic.Solver
}

// New создаёт новый экземпляр оптимизатора на основе алгоритма Диница.
func New() *Optimizer {
	return &Optimizer{maxFlowSolver: dinic.Solver{}}
}

// Optimize выполняет расчёт без ограничения срока.
func (o *Optimizer) Optimize(debts []optimizers.Transfer) ([]optimizers.Transfer, error) {
	return o.OptimizeContext(context.Background(), debts)
}

// OptimizeContext проверяет отмену между этапами и внутри расчёта максимального потока.
func (o *Optimizer) OptimizeContext(ctx context.Context, debts []optimizers.Transfer) ([]optimizers.Transfer, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(debts) == 0 {
		return nil, nil
	}

	debts, err := utils.CollapseTransfers(debts)
	if err != nil {
		return nil, err
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	balances, err := utils.Balances(debts)
	if err != nil {
		return nil, err
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	users := utils.Users(debts)
	index := make(map[string]int, len(users))
	for i, u := range users {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		index[u] = i
	}

	originalOut := make(map[string]int, len(users))
	for _, tr := range debts {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if tr.Amount > 0 {
			originalOut[tr.From] += tr.Amount
		}
	}

	totalDemand := 0
	for _, user := range users {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
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
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		in := i
		out := i + n
		cap := originalOut[user]
		if cap > 0 {
			graph.AddEdge(in, out, cap)
		}
	}

	for i, user := range users {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
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
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		fromIdx := index[from]
		fromOut := fromIdx + n
		for to := range toMap {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			toIdx := index[to]
			if fromIdx == toIdx {
				continue
			}
			toIn := toIdx
			graph.AddEdge(fromOut, toIn, infCap)
		}
	}

	maxFlow, err := o.maxFlowSolver.GetMaxFlowContext(ctx, graph, source, sink)
	if err != nil {
		return nil, err
	}
	if maxFlow != totalDemand {
		return nil, fmt.Errorf("insufficient max flow: %d of %d", maxFlow, totalDemand)
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	result := utils.TransfersFromFlowGraph(graph, n, users)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
