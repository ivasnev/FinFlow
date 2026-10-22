package errors

import "errors"

// ErrOCRUnavailable обозначает занятый или прогревающийся обработчик, а не испорченный чек.
var ErrOCRUnavailable = errors.New("OCR temporarily unavailable")

// ErrReceiptStorageUnavailable обозначает временный сбой сохранения фотографии.
var ErrReceiptStorageUnavailable = errors.New("receipt image storage unavailable")
