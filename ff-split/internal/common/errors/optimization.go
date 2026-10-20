package errors

import "errors"

// ErrOptimizationBusy позволяет отклонять лишние вычисления без очереди.
var ErrOptimizationBusy = errors.New("debt optimization is busy; retry later")

// ErrDebtRevisionChanged запрещает публиковать результат устаревшего снимка долгов.
var ErrDebtRevisionChanged = errors.New("debts changed during optimization; retry")
