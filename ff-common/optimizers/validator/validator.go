package validator

import (
	"fmt"

	"github.com/ivasnev/FinFlow/ff-common/optimizers"
)

const (
	CriterionBalances     = "balances"
	CriterionNoNewDebts   = "no_new_debts"
	CriterionTotalNotMore = "total_not_more"
)

// Violation описывает нарушение критерия валидации.
type Violation struct {
	Criterion string
	Message   string
}

// Report содержит результат валидации.
type Report struct {
	Valid      bool
	Violations []Violation
}

// Option определяет опцию конфигурации валидатора.
type Option func(*config)

type config struct {
	checkBalances     bool
	checkNoNewDebts   bool
	checkTotalNotMore bool
}

// WithBalancesCheck включает проверку сохранения балансов.
// Итоговый баланс каждого участника должен быть одинаков до и после оптимизации.
func WithBalancesCheck() Option {
	return func(c *config) {
		c.checkBalances = true
	}
}

// WithNoNewDebtsCheck включает проверку отсутствия новых долгов.
// Нельзя создавать долг от A к B, если изначально A не был должен B.
func WithNoNewDebtsCheck() Option {
	return func(c *config) {
		c.checkNoNewDebts = true
	}
}

// WithTotalNotMoreCheck включает проверку суммарного долга.
// Никто не должен платить в сумме больше, чем до упрощения.
func WithTotalNotMoreCheck() Option {
	return func(c *config) {
		c.checkTotalNotMore = true
	}
}

// WithAllChecks включает все проверки.
func WithAllChecks() Option {
	return func(c *config) {
		c.checkBalances = true
		c.checkNoNewDebts = true
		c.checkTotalNotMore = true
	}
}

// Validator выполняет валидацию результата оптимизации.
type Validator struct {
	cfg config
}

// NewValidator создаёт новый валидатор с указанными опциями.
func NewValidator(opts ...Option) *Validator {
	cfg := config{}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Validator{cfg: cfg}
}

// Validate проверяет, что optimized корректно оптимизирует original.
func (v *Validator) Validate(original, optimized []optimizers.Transfer) Report {
	report := Report{Valid: true}
	addViolation := func(criterion, message string) {
		report.Valid = false
		report.Violations = append(report.Violations, Violation{
			Criterion: criterion,
			Message:   message,
		})
	}

	if v.cfg.checkBalances {
		v.validateBalances(original, optimized, addViolation)
	}

	if v.cfg.checkNoNewDebts {
		v.validateNoNewDebts(original, optimized, addViolation)
	}

	if v.cfg.checkTotalNotMore {
		v.validateTotalNotMore(original, optimized, addViolation)
	}

	return report
}

// validateBalances проверяет, что балансы всех участников одинаковы до и после оптимизации.
func (v *Validator) validateBalances(original, optimized []optimizers.Transfer, addViolation func(string, string)) {
	origBalances, err := optimizers.Balances(original)
	if err != nil {
		addViolation(CriterionBalances, fmt.Sprintf("ошибка вычисления балансов original: %v", err))
		return
	}

	optBalances, err := optimizers.Balances(optimized)
	if err != nil {
		addViolation(CriterionBalances, fmt.Sprintf("ошибка вычисления балансов optimized: %v", err))
		return
	}

	allUsers := make(map[string]struct{})
	for user := range origBalances {
		allUsers[user] = struct{}{}
	}
	for user := range optBalances {
		allUsers[user] = struct{}{}
	}

	for user := range allUsers {
		origBal := origBalances[user]
		optBal := optBalances[user]
		if origBal != optBal {
			addViolation(
				CriterionBalances,
				fmt.Sprintf("баланс пользователя %s изменился: было %d, стало %d", user, origBal, optBal),
			)
		}
	}
}

// validateNoNewDebts проверяет, что не создаются новые долговые связи.
func (v *Validator) validateNoNewDebts(original, optimized []optimizers.Transfer, addViolation func(string, string)) {
	origMatrix := optimizers.TransferMatrix(original)
	optMatrix := optimizers.TransferMatrix(optimized)

	for from, toMap := range optMatrix {
		for to, amount := range toMap {
			if amount <= 0 {
				continue
			}
			if origMatrix[from] == nil || origMatrix[from][to] == 0 {
				addViolation(
					CriterionNoNewDebts,
					fmt.Sprintf("нельзя создавать новый долг: %s -> %s", from, to),
				)
			}
		}
	}
}

// validateTotalNotMore проверяет, что никто не платит в сумме больше, чем до оптимизации.
func (v *Validator) validateTotalNotMore(original, optimized []optimizers.Transfer, addViolation func(string, string)) {
	origOut := make(map[string]int)
	for _, tr := range original {
		if tr.Amount > 0 {
			origOut[tr.From] += tr.Amount
		}
	}

	optOut := make(map[string]int)
	for _, tr := range optimized {
		if tr.Amount > 0 {
			optOut[tr.From] += tr.Amount
		}
	}

	for from, out := range optOut {
		if out > origOut[from] {
			addViolation(
				CriterionTotalNotMore,
				fmt.Sprintf("суммарный долг увеличился для %s: было %d, стало %d", from, origOut[from], out),
			)
		}
	}
}
