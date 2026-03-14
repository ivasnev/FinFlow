package testutil

import "github.com/ivasnev/FinFlow/ff-common/optimizers"

// DebtsDirectSimple возвращает простой случай с прямым долгом.
func DebtsDirectSimple() []optimizers.Transfer {
	return []optimizers.Transfer{
		{From: "A", To: "B", Amount: 10},
	}
}

// DebtsNeedsIntermediate возвращает случай, требующий промежуточную вершину.
func DebtsNeedsIntermediate() []optimizers.Transfer {
	return []optimizers.Transfer{
		{From: "A", To: "B", Amount: 5},
		{From: "B", To: "C", Amount: 5},
	}
}

// DebtsComplexGraph возвращает сложный граф долгов.
// Original graph:
// - Fred -> Bob: $10, Fred -> Charlie: $30, Fred -> David: $10, Fred -> Ema: $10
// - Gabe -> Bob: $30, Gabe -> David: $10
// - Bob -> Charlie: $40
// - Charlie -> David: $20
// - David -> Ema: $50
func DebtsComplexGraph() []optimizers.Transfer {
	return []optimizers.Transfer{
		{From: "Fred", To: "Bob", Amount: 10},
		{From: "Fred", To: "Charlie", Amount: 30},
		{From: "Fred", To: "David", Amount: 10},
		{From: "Fred", To: "Ema", Amount: 10},
		{From: "Gabe", To: "Bob", Amount: 30},
		{From: "Gabe", To: "David", Amount: 10},
		{From: "Bob", To: "Charlie", Amount: 40},
		{From: "Charlie", To: "David", Amount: 20},
		{From: "David", To: "Ema", Amount: 50},
	}
}

// DebtsTriangle возвращает треугольный граф долгов для тестирования.
func DebtsTriangle() []optimizers.Transfer {
	return []optimizers.Transfer{
		{From: "A", To: "B", Amount: 10},
		{From: "B", To: "C", Amount: 10},
		{From: "C", To: "A", Amount: 10},
	}
}

// DebtsEmpty возвращает пустой список долгов.
func DebtsEmpty() []optimizers.Transfer {
	return nil
}
