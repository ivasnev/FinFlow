package debt_calculator

import (
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"

	"github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

// Константы для типов распределения
const (
	PercentType = "percent" // Процентное распределение
	AmountType  = "amount"  // Фиксированные суммы
	UnitsType   = "units"   // Распределение по долям
)

// DebtCalculator интерфейс для стратегий расчета долгов
type DebtCalculator interface {
	// Calculate рассчитывает доли и долги для транзакции
	Calculate(req *service.TransactionRequest, eventID int64) ([]Share, []Debt, error)
}

// Share представляет долю пользователя в транзакции (для внутреннего использования)
type Share struct {
	UserID int64
	Value  float64
}

// Debt представляет долг между пользователями (для внутреннего использования)
type Debt struct {
	FromUserID int64
	ToUserID   int64
	Amount     float64
}

// GetCalculator возвращает стратегию расчета по типу
func GetCalculator(calculationType string) (DebtCalculator, error) {
	switch calculationType {
	case PercentType:
		return &PercentStrategy{}, nil
	case AmountType:
		return &AmountStrategy{}, nil
	case UnitsType:
		return &UnitsStrategy{}, nil
	default:
		return nil, errors.NewValidationError("type", "неизвестный тип расчета долгов")
	}
}

// PercentStrategy стратегия распределения на основе процентов
type PercentStrategy struct{}

// Calculate рассчитывает доли и долги по процентам
func (s *PercentStrategy) Calculate(req *service.TransactionRequest, eventID int64) ([]Share, []Debt, error) {
	totalCents, err := validateRequest(req, PercentType)
	if err != nil {
		return nil, nil, err
	}
	// Проверка суммы процентов (должно быть 100%)
	var totalPercent float64
	for _, p := range req.Portion {
		totalPercent += p
	}

	if math.IsInf(totalPercent, 0) || math.Abs(totalPercent-100.0) > 0.0100000001 {
		return nil, nil, errors.NewValidationError("portion", "сумма процентов должна быть равна 100")
	}

	// Расчет долей на основе процентов
	shares := make([]Share, 0, len(req.Portion))
	for userIDStr, percent := range req.Portion {
		userID, err := parseUserID(userIDStr)
		if err != nil {
			return nil, nil, err
		}

		shares = append(shares, Share{
			UserID: userID,
			Value:  percent,
		})
	}
	shares = allocateCents(totalCents, shares)

	// Расчет долгов
	debts := calculateDebts(shares, req.FromUser)
	return shares, debts, nil
}

// AmountStrategy стратегия фиксированных сумм
type AmountStrategy struct{}

// Calculate рассчитывает доли и долги по фиксированным суммам
func (s *AmountStrategy) Calculate(req *service.TransactionRequest, eventID int64) ([]Share, []Debt, error) {
	totalCents, err := validateRequest(req, AmountType)
	if err != nil {
		return nil, nil, err
	}

	// Расчет долей
	shares := make([]Share, 0, len(req.Portion))
	var allocated int64
	for userIDStr, amount := range req.Portion {
		cents, err := moneyCents(amount, "portion")
		if err != nil {
			return nil, nil, err
		}
		// Каждая доля ограничена суммой расхода, поэтому накопление не переполнится.
		if cents > totalCents-allocated {
			return nil, nil, errors.NewValidationError("portion", "сумма долей превышает расход")
		}
		allocated += cents
		userID, err := parseUserID(userIDStr)
		if err != nil {
			return nil, nil, err
		}

		shares = append(shares, Share{
			UserID: userID,
			Value:  float64(cents) / 100,
		})
	}
	if allocated != totalCents {
		return nil, nil, errors.NewValidationError("portion", "сумма долей должна совпадать с расходом до копейки")
	}
	sort.Slice(shares, func(i, j int) bool { return shares[i].UserID < shares[j].UserID })

	// Расчет долгов
	debts := calculateDebts(shares, req.FromUser)
	return shares, debts, nil
}

// UnitsStrategy стратегия распределения на основе долей
type UnitsStrategy struct{}

// Calculate рассчитывает доли и долги по единицам долей
func (s *UnitsStrategy) Calculate(req *service.TransactionRequest, eventID int64) ([]Share, []Debt, error) {
	totalCents, err := validateRequest(req, UnitsType)
	if err != nil {
		return nil, nil, err
	}
	// Подготовка карты долей
	unitMap := make(map[int64]float64)
	redistributeValue := 0.0
	redistribute := false
	redistributeCount := 0

	// Сначала заполняем карту долей для всех пользователей значением 1.0
	for _, userID := range req.Users {
		unitMap[userID] = 1.0
	}

	// Затем обрабатываем явно указанные доли
	for userIDStr, units := range req.Portion {
		// Проверка на распределение остатка (-1)
		if userIDStr == "-1" {
			redistributeValue = units
			redistribute = true
			continue
		}

		userID, err := parseUserID(userIDStr)
		if err != nil {
			return nil, nil, err
		}

		unitMap[userID] = units
	}

	// Подсчет пользователей, на которых нужно распределить остаток (не указанные явно)
	for _, userID := range req.Users {
		_, exists := req.Portion[fmt.Sprintf("%d", userID)]
		if !exists {
			redistributeCount++
		}
	}

	// Распределение остатка, если указан параметр -1
	if redistribute && redistributeCount > 0 {
		valuePerUser := redistributeValue / float64(redistributeCount)
		for _, userID := range req.Users {
			_, exists := req.Portion[fmt.Sprintf("%d", userID)]
			if !exists {
				unitMap[userID] = valuePerUser
			}
		}
	}

	// Вычисляем общее количество долей
	var totalUnits float64
	for _, units := range unitMap {
		totalUnits += units
	}
	if totalUnits <= 0 || math.IsInf(totalUnits, 0) {
		return nil, nil, errors.NewValidationError("portion", "сумма долей должна быть положительной и конечной")
	}

	// Создаем итоговый список долей
	shares := make([]Share, 0, len(unitMap))
	for userID, units := range unitMap {
		shares = append(shares, Share{
			UserID: userID,
			Value:  units,
		})
	}
	shares = allocateCents(totalCents, shares)

	// Расчет долгов
	debts := calculateDebts(shares, req.FromUser)
	return shares, debts, nil
}

// validateRequest проверяет денежную сумму и допустимость весов до расчета.
func validateRequest(req *service.TransactionRequest, calculationType string) (int64, error) {
	if req == nil {
		return 0, errors.NewValidationError("transaction", "запрос отсутствует")
	}
	cents, err := moneyCents(req.Amount, "amount")
	if err != nil {
		return 0, err
	}
	if cents <= 0 {
		return 0, errors.NewValidationError("amount", "сумма должна быть положительной")
	}
	for _, value := range req.Portion {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
			return 0, errors.NewValidationError("portion", "доли должны быть конечными и неотрицательными")
		}
	}
	if req.FromUser <= 0 || len(req.Users) == 0 {
		return 0, errors.NewValidationError("users", "плательщик и участники должны быть указаны")
	}
	users := make(map[int64]struct{}, len(req.Users))
	for _, id := range req.Users {
		if _, exists := users[id]; id <= 0 || exists {
			return 0, errors.NewValidationError("users", "ID участников должны быть положительными и уникальными")
		}
		users[id] = struct{}{}
	}
	for key := range req.Portion {
		if calculationType == UnitsType && key == "-1" {
			continue
		}
		id, err := parseUserID(key)
		if err != nil {
			return 0, err
		}
		if _, exists := users[id]; !exists {
			return 0, errors.NewValidationError("portion", "доля указана для пользователя вне списка users")
		}
	}
	if calculationType != UnitsType && len(req.Portion) != len(users) {
		return 0, errors.NewValidationError("portion", "доли должны быть указаны для всех участников расхода")
	}
	return cents, nil
}

// moneyCents принимает суммы в пределах numeric(10,2) без дробных копеек.
func moneyCents(value float64, field string) (int64, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 99999999.99 {
		return 0, errors.NewValidationError(field, "сумма должна быть конечной и находиться в диапазоне 0..99999999.99")
	}
	scaled := value * 100
	if math.Abs(scaled-math.Round(scaled)) > 0.00001 {
		return 0, errors.NewValidationError(field, "дробные копейки недопустимы")
	}
	return int64(math.Round(scaled)), nil
}

// allocateCents распределяет целые копейки методом наибольших остатков.
// Рациональные числа исключают потерю копеек при крайних весах; равные остатки
// разрешаются по ID, поэтому порядок обхода map не влияет на результат.
func allocateCents(totalCents int64, shares []Share) []Share {
	sort.Slice(shares, func(i, j int) bool { return shares[i].UserID < shares[j].UserID })
	totalWeight := new(big.Rat)
	for _, share := range shares {
		totalWeight.Add(totalWeight, new(big.Rat).SetFloat64(share.Value))
	}
	remainders := make([]*big.Rat, len(shares))
	order := make([]int, len(shares))
	remaining := totalCents
	for i := range shares {
		quota := new(big.Rat).SetFloat64(shares[i].Value)
		quota.Mul(quota, new(big.Rat).SetInt64(totalCents))
		quota.Quo(quota, totalWeight)
		cents := new(big.Int).Quo(quota.Num(), quota.Denom()).Int64()
		remainders[i] = new(big.Rat).Sub(quota, new(big.Rat).SetInt64(cents))
		shares[i].Value = float64(cents) / 100
		remaining -= cents
		order[i] = i
	}
	sort.SliceStable(order, func(i, j int) bool { return remainders[order[i]].Cmp(remainders[order[j]]) > 0 })
	for _, index := range order[:int(remaining)] {
		shares[index].Value = float64(int64(math.Round(shares[index].Value*100))+1) / 100
	}
	return shares
}

// parseUserID преобразует строковый ID пользователя в int64
func parseUserID(userIDStr string) (int64, error) {
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 || strconv.FormatInt(userID, 10) != userIDStr {
		return 0, errors.NewValidationError("portion", "ключ должен быть положительным целым ID пользователя без префиксов и суффиксов")
	}
	return userID, nil
}

// calculateDebts рассчитывает долги на основе долей
func calculateDebts(shares []Share, payerID int64) []Debt {
	debts := make([]Debt, 0)

	// Рассчитываем долги для каждого участника
	for _, share := range shares {
		// Пропускаем плательщика
		if share.UserID == payerID {
			continue
		}

		// Создаем долг только если сумма больше нуля
		if share.Value > 0 {
			debts = append(debts, Debt{
				FromUserID: share.UserID,
				ToUserID:   payerID,
				Amount:     share.Value,
			})
		}
	}

	return debts
}
