package ds

// Статусы заявки (храним как SMALLINT 1..5)
const (
	StatusDraft     = 1 // Черновик
	StatusDeleted   = 2 // Удалён
	StatusFormed    = 3 // Сформирован
	StatusCompleted = 4 // Завершён
	StatusRejected  = 5 // Отклонён
)
