package errors

import "errors"

// ErrOCRUnavailable обозначает занятый или прогревающийся обработчик, а не испорченный чек.
var ErrOCRUnavailable = errors.New("OCR temporarily unavailable")

// ErrReceiptStorageUnavailable обозначает временный сбой сохранения фотографии.
var ErrReceiptStorageUnavailable = errors.New("receipt image storage unavailable")

// ErrReceiptReadBudget denies further paid storage reads until the quota recovers.
var ErrReceiptReadBudget = errors.New("достигнут лимит просмотра фотографий, попробуйте позже")
