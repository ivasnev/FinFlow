package maxflow

// Edge — ребро графа (для использования солверами).
type Edge struct {
	To   int
	Rev  int
	Cap  int
	Flow int
}

// Graph — ориентированный граф для задачи максимального потока.
// Вершины: 0..NumNodes()-1. Рёбра хранятся как списки смежности с обратными ссылками.
// Adj доступен солверам для вычисления потока.
type Graph struct {
	Adj [][]*Edge
}

// NewGraph создаёт граф с n вершинами (0..n-1).
func NewGraph(n int) *Graph {
	return &Graph{
		Adj: make([][]*Edge, n),
	}
}

// AddEdge добавляет ориентированное ребро (from, to) с пропускной способностью cap
// и обратное ребро с нулевой пропускной способностью.
func (g *Graph) AddEdge(from, to, cap int) {
	fwd := &Edge{To: to, Rev: len(g.Adj[to]), Cap: cap}
	rev := &Edge{To: from, Rev: len(g.Adj[from]), Cap: 0}
	g.Adj[from] = append(g.Adj[from], fwd)
	g.Adj[to] = append(g.Adj[to], rev)
}

// NumNodes возвращает число вершин.
func (g *Graph) NumNodes() int {
	return len(g.Adj)
}

// OutEdge — исходящее ребро для чтения потока и пропускной способности после решения.
type OutEdge struct {
	To   int
	Cap  int
	Flow int
}

// OutEdges возвращает исходящие рёбра из вершины from (целевая вершина, пропускная способность, поток).
func (g *Graph) OutEdges(from int) []OutEdge {
	if from < 0 || from >= len(g.Adj) {
		return nil
	}
	out := make([]OutEdge, 0, len(g.Adj[from]))
	for _, e := range g.Adj[from] {
		out = append(out, OutEdge{To: e.To, Cap: e.Cap, Flow: e.Flow})
	}
	return out
}
