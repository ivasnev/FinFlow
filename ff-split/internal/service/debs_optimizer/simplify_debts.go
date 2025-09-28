package debs_optimizer

import (
	"fmt"
)

const INF int = 1e9             // "бесконечность" для избежания переполнения
const OFFSET int64 = 1000000000 // смещение для хеш-функции ребер

// Debt представляет долг от одного человека другому
type Debt struct {
	Creditor string
	Debtor   string
	Amount   int
}

// SimplifyDebts упрощает долги между людьми
func SimplifyDebts(debts map[string]map[string]int) map[string]map[string]int {
	// Создаем список всех людей
	people := make(map[string]bool)
	for creditor, debtors := range debts {
		people[creditor] = true
		for debtor := range debtors {
			people[debtor] = true
		}
	}

	// Преобразуем множество в список
	personList := make([]string, 0, len(people))
	for person := range people {
		personList = append(personList, person)
	}

	// Создаем отображение имен на индексы
	nameToIndex := make(map[string]int)
	for i, name := range personList {
		nameToIndex[name] = i
	}

	n := len(personList)
	solver := NewDinics(n, personList)

	// Добавляем все транзакции в граф
	for creditor, debtors := range debts {
		for debtor, amount := range debtors {
			fromIdx := nameToIndex[creditor]
			toIdx := nameToIndex[debtor]
			solver.AddEdge(fromIdx, toIdx, amount)
		}
	}

	// Множество для отслеживания посещенных ребер
	visitedEdges := make(map[int64]bool)

	getHashKeyForEdge := func(u, v int) int64 {
		return int64(u)*OFFSET + int64(v)
	}

	getNonVisitedEdge := func(edges []*Edge) int {
		for i, edge := range edges {
			if !visitedEdges[getHashKeyForEdge(edge.From, edge.To)] {
				return i
			}
		}
		return -1
	}

	for {
		edgePos := getNonVisitedEdge(solver.GetEdges())
		if edgePos == -1 {
			break
		}

		// Пересчитываем последующие потоки в графе
		solver.Recompute()

		// Устанавливаем источник и сток в графе потока
		firstEdge := solver.GetEdges()[edgePos]
		solver.SetSource(firstEdge.From)
		solver.SetSink(firstEdge.To)

		// Инициализируем остаточный граф как копию исходного графа
		residualGraph := solver.GetGraph()
		newEdges := make([]*Edge, 0)

		for node, edges := range residualGraph {
			for _, edge := range edges {
				remainingFlow := 0
				if edge.Flow < 0 {
					remainingFlow = edge.Capacity
				} else {
					remainingFlow = edge.Capacity - edge.Flow
				}

				// Если в графе осталась пропускная способность, добавляем ее
				if remainingFlow > 0 {
					newEdges = append(newEdges, &Edge{
						From:     node,
						To:       edge.To,
						Capacity: remainingFlow,
					})
				}
			}
		}

		// Получаем максимальный поток между источником и стоком
		maxFlow := solver.GetMaxFlow()

		// Отмечаем ребро от источника к стоку как посещенное
		source := solver.GetSource()
		sink := solver.GetSink()
		visitedEdges[getHashKeyForEdge(source, sink)] = true

		// Создаем новый граф
		solver = NewDinics(n, personList)

		// Добавляем ребра с оставшейся пропускной способностью
		solver.AddEdges(newEdges)

		// Добавляем ребро от источника к стоку в новый граф
		if maxFlow > 0 {
			solver.AddEdge(source, sink, maxFlow)
		}
	}

	// Формируем результат в виде словаря
	result := make(map[string]map[string]int)

	for _, edge := range solver.GetEdges() {
		if edge.Capacity > 0 { // Берем только ненулевые транзакции
			creditor := personList[edge.From]
			debtor := personList[edge.To]
			amount := edge.Capacity

			if _, exists := result[creditor]; !exists {
				result[creditor] = make(map[string]int)
			}
			result[creditor][debtor] = amount
		}
	}

	return result
}

// PrintTransactions выводит все упрощенные транзакции
func PrintTransactions(transactions map[string]map[string]int) {
	fmt.Println("Упрощенные транзакции:")
	for creditor, debtors := range transactions {
		for debtor, amount := range debtors {
			fmt.Printf("%s -> %s: %d\n", creditor, debtor, amount)
		}
	}
}
