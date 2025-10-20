package transaction

import (
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// extract преобразует модель транзакции БД в бизнес-модель
func extract(dbTransaction *Transaction) *models.Transaction {
	if dbTransaction == nil {
		return nil
	}

	return &models.Transaction{
		ID:                    dbTransaction.ID,
		EventID:               dbTransaction.EventID,
		Name:                  dbTransaction.Name,
		TransactionCategoryID: dbTransaction.TransactionCategoryID,
		Datetime:              dbTransaction.Datetime,
		TotalPaid:             dbTransaction.TotalPaid,
		PayerID:               dbTransaction.PayerID,
		SplitType:             dbTransaction.SplitType,
	}
}

// extractSlice преобразует слайс моделей транзакций БД в бизнес-модели
func extractSlice(dbTransactions []Transaction) []models.Transaction {
	transactions := make([]models.Transaction, len(dbTransactions))
	for i, dbTransaction := range dbTransactions {
		if extracted := extract(&dbTransaction); extracted != nil {
			transactions[i] = *extracted
		}
	}
	return transactions
}

// load преобразует бизнес-модель транзакции в модель БД
func load(transaction *models.Transaction) *Transaction {
	if transaction == nil {
		return nil
	}

	return &Transaction{
		ID:                    transaction.ID,
		EventID:               transaction.EventID,
		Name:                  transaction.Name,
		TransactionCategoryID: transaction.TransactionCategoryID,
		Datetime:              transaction.Datetime,
		TotalPaid:             transaction.TotalPaid,
		PayerID:               transaction.PayerID,
		SplitType:             transaction.SplitType,
	}
}

// extractTransactionCategory преобразует модель категории транзакции БД в бизнес-модель
func extractTransactionCategory(dbCategory *TransactionCategory) *models.TransactionCategory {
	if dbCategory == nil {
		return nil
	}

	return &models.TransactionCategory{
		ID:     dbCategory.ID,
		Name:   dbCategory.Name,
		IconID: dbCategory.IconID,
	}
}

// loadTransactionCategory преобразует бизнес-модель категории транзакции в модель БД
func loadTransactionCategory(category *models.TransactionCategory) *TransactionCategory {
	if category == nil {
		return nil
	}

	return &TransactionCategory{
		ID:     category.ID,
		Name:   category.Name,
		IconID: category.IconID,
	}
}

// extractTransactionShare преобразует модель доли в транзакции БД в бизнес-модель
func extractTransactionShare(dbShare *TransactionShare) *models.TransactionShare {
	if dbShare == nil {
		return nil
	}

	return &models.TransactionShare{
		ID:            dbShare.ID,
		TransactionID: dbShare.TransactionID,
		UserID:        dbShare.UserID,
		Value:         dbShare.Value,
	}
}

// loadTransactionShare преобразует бизнес-модель доли в транзакции в модель БД
func loadTransactionShare(share *models.TransactionShare) *TransactionShare {
	if share == nil {
		return nil
	}

	return &TransactionShare{
		ID:            share.ID,
		TransactionID: share.TransactionID,
		UserID:        share.UserID,
		Value:         share.Value,
	}
}

// extractDebt преобразует модель долга БД в бизнес-модель
func extractDebt(dbDebt *Debt) *models.Debt {
	if dbDebt == nil {
		return nil
	}

	return &models.Debt{
		ID:            dbDebt.ID,
		TransactionID: dbDebt.TransactionID,
		FromUserID:    dbDebt.FromUserID,
		ToUserID:      dbDebt.ToUserID,
		Amount:        dbDebt.Amount,
	}
}

// extractDebtSlice преобразует слайс моделей долгов БД в бизнес-модели
func extractDebtSlice(dbDebts []Debt) []models.Debt {
	debts := make([]models.Debt, len(dbDebts))
	for i, dbDebt := range dbDebts {
		if extracted := extractDebt(&dbDebt); extracted != nil {
			debts[i] = *extracted
		}
	}
	return debts
}

// loadDebt преобразует бизнес-модель долга в модель БД
func loadDebt(debt *models.Debt) *Debt {
	if debt == nil {
		return nil
	}

	return &Debt{
		ID:            debt.ID,
		TransactionID: debt.TransactionID,
		FromUserID:    debt.FromUserID,
		ToUserID:      debt.ToUserID,
		Amount:        debt.Amount,
	}
}

// extractOptimizedDebt преобразует модель оптимизированного долга БД в бизнес-модель
func extractOptimizedDebt(dbDebt *OptimizedDebt) *models.OptimizedDebt {
	if dbDebt == nil {
		return nil
	}

	return &models.OptimizedDebt{
		ID:         dbDebt.ID,
		EventID:    dbDebt.EventID,
		FromUserID: dbDebt.FromUserID,
		ToUserID:   dbDebt.ToUserID,
		Amount:     dbDebt.Amount,
		CreatedAt:  dbDebt.CreatedAt,
		UpdatedAt:  dbDebt.UpdatedAt,
	}
}

// extractOptimizedDebtSlice преобразует слайс моделей оптимизированных долгов БД в бизнес-модели
func extractOptimizedDebtSlice(dbDebts []OptimizedDebt) []models.OptimizedDebt {
	debts := make([]models.OptimizedDebt, len(dbDebts))
	for i, dbDebt := range dbDebts {
		if extracted := extractOptimizedDebt(&dbDebt); extracted != nil {
			debts[i] = *extracted
		}
	}
	return debts
}

// loadOptimizedDebt преобразует бизнес-модель оптимизированного долга в модель БД
func loadOptimizedDebt(debt *models.OptimizedDebt) *OptimizedDebt {
	if debt == nil {
		return nil
	}

	return &OptimizedDebt{
		ID:         debt.ID,
		EventID:    debt.EventID,
		FromUserID: debt.FromUserID,
		ToUserID:   debt.ToUserID,
		Amount:     debt.Amount,
		CreatedAt:  debt.CreatedAt,
		UpdatedAt:  debt.UpdatedAt,
	}
}
