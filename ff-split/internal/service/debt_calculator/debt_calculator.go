package debt_calculator

import (
	"errors"
	"fmt"
	"math"

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
		return nil, fmt.Errorf("неизвестный тип расчета долгов: %s", calculationType)
	}
}

// PercentStrategy стратегия распределения на основе процентов
type PercentStrategy struct{}

// Calculate рассчитывает доли и долги по процентам
func (s *PercentStrategy) Calculate(req *service.TransactionRequest, eventID int64) ([]Share, []Debt, error) {
	// Проверка суммы процентов (должно быть 100%)
	var totalPercent float64
	for _, p := range req.Portion {
		totalPercent += p
	}

	if math.Abs(totalPercent-100.0) > 0.01 {
		return nil, nil, errors.New("сумма процентов должна быть равна 100")
	}

	// Расчет долей на основе процентов
	shares := make([]Share, 0, len(req.Portion))
	for userIDStr, percent := range req.Portion {
		userID, err := parseUserID(userIDStr)
		if err != nil {
			return nil, nil, err
		}

		value := req.Amount * percent / 100.0
		shares = append(shares, Share{
			UserID: userID,
			Value:  value,
		})
	}

	// Расчет долгов
	debts := calculateDebts(shares, req.FromUser)
	return shares, debts, nil
}

// AmountStrategy стратегия фиксированных сумм
type AmountStrategy struct{}

// Calculate рассчитывает доли и долги по фиксированным суммам
func (s *AmountStrategy) Calculate(req *service.TransactionRequest, eventID int64) ([]Share, []Debt, error) {
	// Проверка общей суммы
	var totalAmount float64
	for _, amount := range req.Portion {
		totalAmount += amount
	}

	if math.Abs(totalAmount-req.Amount) > 0.01 {
		return nil, nil, errors.New("сумма распределенных долей должна быть равна общей сумме")
	}

	// Расчет долей
	shares := make([]Share, 0, len(req.Portion))
	for userIDStr, amount := range req.Portion {
		userID, err := parseUserID(userIDStr)
		if err != nil {
			return nil, nil, err
		}

		shares = append(shares, Share{
			UserID: userID,
			Value:  amount,
		})
	}

	// Расчет долгов
	debts := calculateDebts(shares, req.FromUser)
	return shares, debts, nil
}

// UnitsStrategy стратегия распределения на основе долей
type UnitsStrategy struct{}

// Calculate рассчитывает доли и долги по единицам долей
func (s *UnitsStrategy) Calculate(req *service.TransactionRequest, eventID int64) ([]Share, []Debt, error) {
	// Подготовка карты долей
	unitMap := make(map[int64]float64)
	redistributeValue := 0.0
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
	if redistributeValue > 0 && redistributeCount > 0 {
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

	// Рассчитываем стоимость одной доли
	unitCost := req.Amount / totalUnits

	// Создаем итоговый список долей
	shares := make([]Share, 0, len(unitMap))
	for userID, units := range unitMap {
		value := units * unitCost
		shares = append(shares, Share{
			UserID: userID,
			Value:  value,
		})
	}

	// Расчет долгов
	debts := calculateDebts(shares, req.FromUser)
	return shares, debts, nil
}

// Вспомогательные функции

// parseUserID преобразует строковый ID пользователя в int64
func parseUserID(userIDStr string) (int64, error) {
	var userID int64
	_, err := fmt.Sscanf(userIDStr, "%d", &userID)
	if err != nil {
		return 0, fmt.Errorf("неверный формат ID пользователя: %s", userIDStr)
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
