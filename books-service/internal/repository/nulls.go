package repository

import "database/sql"

// Конвертирует ссылочный тип в sql.Null
func ToNull[T any](v *T) sql.Null[T] {
	if v == nil {
		return sql.Null[T]{}
	}
	return sql.Null[T]{V: *v, Valid: true}
}

// Конвертирует sql.Null в ссылочный тип
func FromNull[T any](v sql.Null[T]) *T {
	if !v.Valid {
		return nil
	}
	return &v.V
}
