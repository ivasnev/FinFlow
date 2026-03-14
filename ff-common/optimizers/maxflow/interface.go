package maxflow

// MaxFlowSolver решает задачу максимального потока в графе между source и sink.
// Мутирует поток в графе; возвращает величину максимального потока.
type MaxFlowSolver interface {
	GetMaxFlow(g *Graph, source, sink int) int
}
