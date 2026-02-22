package greedy

import (
	"fmt"

	"github.com/ivasnev/FinFlow/ff-common/optimizers"
)

// Optimizer реализует жадный алгоритм оптимизации долгов.
// Проталкивает поток по целым путям от должников к кредиторам (увеличивающие пути в остаточной сети),
// чтобы корректно обрабатывать любые графы, включая сложные.
type Optimizer struct{}

// New создаёт новый экземпляр жадного оптимизатора.
func New() *Optimizer {
	return &Optimizer{}
}

func (o *Optimizer) Optimize(debts []optimizers.Transfer) ([]optimizers.Transfer, error) {
	if len(debts) == 0 {
		return nil, nil
	}

	debts, err := optimizers.CollapseTransfers(debts)
	if err != nil {
		return nil, err
	}

	balances, err := optimizers.Balances(debts)
	if err != nil {
		return nil, err
	}

	allowed := optimizers.TransferMatrix(debts)
	remaining := make(map[string]int, len(balances))
	for k, v := range balances {
		remaining[k] = v
	}

	capacity := make(map[string]map[string]int)
	for from, toMap := range allowed {
		capacity[from] = make(map[string]int, len(toMap))
		for to, amount := range toMap {
			capacity[from][to] = amount
		}
	}

	users := optimizers.Users(debts)
	var transfers []optimizers.Transfer

	for {
		path, amount := findPath(users, remaining, capacity)
		if path == nil || amount <= 0 {
			for _, user := range users {
				if remaining[user] != 0 {
					return nil, fmt.Errorf("unable to settle: no path for remaining balances")
				}
			}
			return transfers, nil
		}

		for i := 0; i < len(path)-1; i++ {
			from, to := path[i], path[i+1]
			transfers = append(transfers, optimizers.Transfer{From: from, To: to, Amount: amount})
			remaining[from] += amount
			remaining[to] -= amount
			capacity[from][to] -= amount
		}
	}
}

// findPath ищет путь от любого узла с отрицательным балансом к узлу с положительным по рёбрам с capacity > 0.
// Возвращает путь и величину потока (min(|balance| на концах, min capacity на пути)).
func findPath(users []string, remaining map[string]int, capacity map[string]map[string]int) ([]string, int) {
	// BFS от всех узлов с отрицательным балансом
	type state struct {
		node string
		path []string
	}

	for _, user := range users {
		if remaining[user] >= 0 {
			continue
		}

		visited := make(map[string]bool)
		queue := []state{{user, []string{user}}}

		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]

			if remaining[cur.node] > 0 {
				// нашли сток — путь от user до cur.node
				path := cur.path
				amount := -remaining[user]
				if remaining[cur.node] < amount {
					amount = remaining[cur.node]
				}
				for i := 0; i < len(path)-1; i++ {
					cap := capacity[path[i]][path[i+1]]
					if cap < amount {
						amount = cap
					}
				}
				if amount > 0 {
					return path, amount
				}
				continue
			}

			for to, cap := range capacity[cur.node] {
				if cap <= 0 || visited[to] {
					continue
				}
				visited[to] = true
				newPath := make([]string, len(cur.path)+1)
				copy(newPath, cur.path)
				newPath[len(cur.path)] = to
				queue = append(queue, state{to, newPath})
			}
		}
	}

	return nil, 0
}
