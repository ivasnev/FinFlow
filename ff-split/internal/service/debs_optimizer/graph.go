package debs_optimizer

// Edge представляет ребро в графе потока
type Edge struct {
	From, To       int
	Capacity, Flow int
	Residual       *Edge
}

// RemainingCapacity возвращает оставшуюся пропускную способность ребра
func (e *Edge) RemainingCapacity() int {
	return e.Capacity - e.Flow
}

// IsResidual проверяет, является ли ребро обратным
func (e *Edge) IsResidual() bool {
	return e.Capacity == 0
}

// Augment увеличивает поток в ребре
func (e *Edge) Augment(bottleneck int) {
	e.Flow += bottleneck
	e.Residual.Flow -= bottleneck
}
