package optimizers

// Transfer представляет один денежный перевод от одного пользователя к другому.
type Transfer struct {
	From   string
	To     string
	Amount int
}

// Optimizer - общий интерфейс для всех алгоритмов оптимизации долгов.
type Optimizer interface {
	Optimize(debts []Transfer) ([]Transfer, error)
}
