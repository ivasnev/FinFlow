package debs_optimizer

import (
	"container/list"
	"fmt"
)

// Dinics реализует алгоритм максимального потока Диница
type Dinics struct {
	N, S, T      int
	MaxFlow      int
	Graph        [][]Edge
	VertexLabels []string
	Edges        []*Edge
	Level        []int
	Solved       bool
}

// NewDinics создает новый экземпляр решателя максимального потока
func NewDinics(n int, vertexLabels []string) *Dinics {
	d := &Dinics{
		N:            n,
		Graph:        make([][]Edge, n),
		VertexLabels: vertexLabels,
		Edges:        make([]*Edge, 0),
		Level:        make([]int, n),
		Solved:       false,
	}
	return d
}

// AddEdge добавляет ребро в граф
func (d *Dinics) AddEdge(from, to, capacity int) {
	if capacity < 0 {
		panic("Capacity < 0")
	}

	e1 := Edge{From: from, To: to, Capacity: capacity, Flow: 0}
	e2 := Edge{From: to, To: from, Capacity: 0, Flow: 0}

	e1.Residual = &e2
	e2.Residual = &e1

	d.Graph[from] = append(d.Graph[from], e1)
	d.Graph[to] = append(d.Graph[to], e2)

	// Сохраняем прямое ребро для доступа
	copied := e1
	d.Edges = append(d.Edges, &copied)
}

// AddEdges добавляет множество ребер в граф
func (d *Dinics) AddEdges(edges []*Edge) {
	for _, edge := range edges {
		d.AddEdge(edge.From, edge.To, edge.Capacity)
	}
}

// SetSource устанавливает исток для графа потока
func (d *Dinics) SetSource(s int) {
	d.S = s
}

// SetSink устанавливает сток для графа потока
func (d *Dinics) SetSink(t int) {
	d.T = t
}

// GetSource возвращает исток
func (d *Dinics) GetSource() int {
	return d.S
}

// GetSink возвращает сток
func (d *Dinics) GetSink() int {
	return d.T
}

// GetEdges возвращает список ребер
func (d *Dinics) GetEdges() []*Edge {
	return d.Edges
}

// GetGraph возвращает граф после выполнения алгоритма
func (d *Dinics) GetGraph() [][]Edge {
	d.Execute()
	return d.Graph
}

// Recompute сбрасывает флаг решения для повторного вычисления
func (d *Dinics) Recompute() {
	d.Solved = false
}

// Execute выполняет алгоритм, если он еще не выполнен
func (d *Dinics) Execute() {
	if !d.Solved {
		d.Solved = true
		d.Solve()
	}
}

// GetMaxFlow возвращает максимальный поток
func (d *Dinics) GetMaxFlow() int {
	d.Execute()
	return d.MaxFlow
}

// Solve решает задачу о максимальном потоке
func (d *Dinics) Solve() {
	next := make([]int, d.N)

	for d.BFS() {
		for i := 0; i < d.N; i++ {
			next[i] = 0
		}

		// Находим максимальный поток, добавляя все пути улучшения
		f := d.DFS(d.S, next, INF)
		for f != 0 {
			d.MaxFlow += f
			f = d.DFS(d.S, next, INF)
		}
	}
}

// BFS выполняет поиск в ширину, чтобы создать послойный граф
func (d *Dinics) BFS() bool {
	// Инициализируем массив уровней
	for i := 0; i < d.N; i++ {
		d.Level[i] = -1
	}
	d.Level[d.S] = 0

	// Используем очередь для BFS
	queue := list.New()
	queue.PushBack(d.S)

	for queue.Len() > 0 {
		node := queue.Front().Value.(int)
		queue.Remove(queue.Front())

		for i := 0; i < len(d.Graph[node]); i++ {
			edge := &d.Graph[node][i]
			cap := edge.RemainingCapacity()

			if cap > 0 && d.Level[edge.To] == -1 {
				d.Level[edge.To] = d.Level[node] + 1
				queue.PushBack(edge.To)
			}
		}
	}

	// Возвращаем true, если сток был достигнут
	return d.Level[d.T] != -1
}

// DFS выполняет поиск в глубину для нахождения потока через послойный граф
func (d *Dinics) DFS(at int, next []int, flow int) int {
	if at == d.T {
		return flow
	}

	numEdges := len(d.Graph[at])

	for next[at] < numEdges {
		edge := &d.Graph[at][next[at]]
		cap := edge.RemainingCapacity()

		if cap > 0 && d.Level[edge.To] == d.Level[at]+1 {
			bottleneck := d.DFS(edge.To, next, min(flow, cap))

			if bottleneck > 0 {
				edge.Augment(bottleneck)
				return bottleneck
			}
		}

		next[at]++
	}

	return 0
}

// PrintEdges выводит все ребра
func (d *Dinics) PrintEdges() {
	for _, edge := range d.Edges {
		fmt.Printf("%s ----%d----> %s\n", d.VertexLabels[edge.From], edge.Capacity, d.VertexLabels[edge.To])
	}
}

// min возвращает минимальное из двух чисел
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
