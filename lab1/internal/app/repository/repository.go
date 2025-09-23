package repository

import (
	"context"
	"errors"

	"lab1/internal/app/ds"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, err
	}
	return &Repository{db: db}, nil
}

func (r *Repository) DB() *gorm.DB { return r.db }

// ---------- УСЛУГИ (ORM) ----------

func (r *Repository) ListServices(search string) ([]ds.Service, error) {
	q := r.db.Model(&ds.Service{}).Where("is_active = true")
	if search != "" {
		ilike := "%" + search + "%"
		q = q.Where("title ILIKE ? OR description ILIKE ?", ilike, ilike)
	}
	var out []ds.Service
	return out, q.Order("id ASC").Find(&out).Error
}

func (r *Repository) GetServiceByID(id uint) (ds.Service, error) {
	var s ds.Service
	err := r.db.Where("id = ? AND is_active = true", id).First(&s).Error
	return s, err
}

// ---------- ТЕКУЩАЯ ЗАЯВКА (draft) ----------

func (r *Repository) GetDraftRequestID(userID uint) (uint, error) {
	var req ds.Request
	err := r.db.Where("creator_id = ? AND status = ?", userID, ds.StatusDraft).First(&req).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	return req.ID, err
}

func (r *Repository) CreateDraftRequest(userID uint) (uint, error) {
	req := ds.Request{
		Status:    ds.StatusDraft,
		CreatorID: userID,
	}
	if err := r.db.Create(&req).Error; err != nil {
		return 0, err
	}
	return req.ID, nil
}

func (r *Repository) CountDraftItems(userID uint) (int64, error) {
	draftID, err := r.GetDraftRequestID(userID)
	if err != nil || draftID == 0 {
		return 0, err
	}
	var n int64
	if err := r.db.Table("request_services").Where("request_id = ?", draftID).Count(&n).Error; err != nil {
		return 0, err
	}
	return n, nil
}

// ---------- ДОБАВЛЕНИЕ УСЛУГИ В ЗАЯВКУ (ORM) ----------

func (r *Repository) AddServiceToDraft(userID, serviceID uint, quantity float64) (uint, error) {
	if quantity <= 0 {
		quantity = 1
	}

	// найдём/создадим черновик
	draftID, err := r.GetDraftRequestID(userID)
	if err != nil {
		return 0, err
	}
	if draftID == 0 {
		draftID, err = r.CreateDraftRequest(userID)
		if err != nil {
			return 0, err
		}
	}

	// вытаскиваем тариф услуги
	svc, err := r.GetServiceByID(serviceID)
	if err != nil {
		return 0, err
	}

	item := ds.RequestService{
		RequestID:      draftID,
		ServiceID:      serviceID,
		Quantity:       quantity,
		TariffSnapshot: svc.Tariff,
		TotalLine:      round2(quantity * svc.Tariff),
	}

	// upsert по составному ключу: если позиция уже есть — увеличиваем количество и пересчитываем сумму
	err = r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "request_id"}, {Name: "service_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"quantity":   gorm.Expr("request_services.quantity + EXCLUDED.quantity"),
			"total_line": gorm.Expr("(request_services.quantity + EXCLUDED.quantity) * request_services.tariff_snapshot"),
		}),
	}).Create(&item).Error
	if err != nil {
		return 0, err
	}

	return draftID, nil
}

// ---------- ПРОСМОТР КОРЗИНЫ (ORM) ----------
// Возвращаем строки m2m с подгруженной услугой (Service),
// чтобы в шаблоне использовать .Service.Title / .Service.Description и т.д.

func (r *Repository) GetDraftCart(userID uint) (uint, []ds.RequestService, float64, error) {
	draftID, err := r.GetDraftRequestID(userID)
	if err != nil || draftID == 0 {
		return 0, nil, 0, err
	}

	var lines []ds.RequestService
	err = r.db.
		Preload("Service").
		Where("request_id = ?", draftID).
		Order("COALESCE(position, 999999), service_id ASC").
		Find(&lines).Error
	if err != nil {
		return 0, nil, 0, err
	}

	var total float64
	for _, l := range lines {
		total += l.TotalLine
	}
	total = round2(total)

	return draftID, lines, total, nil
}

// ---------- ЛОГИЧЕСКОЕ УДАЛЕНИЕ ЗАЯВКИ (RAW SQL) ----------

func (r *Repository) SoftDeleteDraft(ctx context.Context, userID, requestID uint) (int64, error) {
	sql := `
	  UPDATE requests
	     SET status = $1
	   WHERE id = $2 AND creator_id = $3 AND status = $4
	`
	tx := r.db.WithContext(ctx).Exec(sql, ds.StatusDeleted, requestID, userID, ds.StatusDraft)
	return tx.RowsAffected, tx.Error
}

// ---------- утилиты ----------

func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100.0
}

// Просмотр любой АКТИВНОЙ заявки пользователя (не Deleted).
// Возвращает 0, nil, 0, nil — если заявка не найдена/не его/удалена.
func (r *Repository) GetCartByID(userID, requestID uint) (uint, []ds.RequestService, float64, error) {
	// сначала проверим, что заявка принадлежит пользователю и не удалена
	var req ds.Request
	err := r.db.
		Where("id = ? AND creator_id = ? AND status <> ?", requestID, userID, ds.StatusDeleted).
		First(&req).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil, 0, nil
	}
	if err != nil {
		return 0, nil, 0, err
	}

	// тянем строки m2m вместе с услугами
	var lines []ds.RequestService
	if err := r.db.
		Preload("Service").
		Where("request_id = ?", requestID).
		Order("COALESCE(position, 999999), service_id ASC").
		Find(&lines).Error; err != nil {
		return 0, nil, 0, err
	}

	var total float64
	for _, l := range lines {
		total += l.TotalLine
	}
	return requestID, lines, round2(total), nil
}
