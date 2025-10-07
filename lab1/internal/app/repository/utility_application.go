package repository

import (
	"lab1/internal/app/ds"
	"time"

	"gorm.io/gorm"
)

// Создать "черновик" заявки (по аналогии с CreateSystemCalc)
func (r *Repository) CreateUtilityApplication(userID uint) (ds.UtilityApplication, error) {
	app := ds.UtilityApplication{
		UserID:      userID,
		Status:      string(ds.DRAFT), // можно и "DRAFT" строкой, но лучше константой, как у тебя в SystemCalc
		DateCreated: time.Now(),
	}
	if err := r.db.Create(&app).Error; err != nil {
		return ds.UtilityApplication{}, err
	}
	return app, nil
}

// Найти черновик заявки пользователя (как GetSystemCalc)
func (r *Repository) GetUtilityApplicationDraft(userID uint) (ds.UtilityApplication, error) {
	var app ds.UtilityApplication
	err := r.db.
		Where("user_id = ? AND status = ?", userID, ds.DRAFT).
		First(&app).Error
	if err != nil {
		return ds.UtilityApplication{}, err
	}
	return app, nil
}

// Создать или получить существующий черновик (как CreateOrGetSystemCalc)
func (r *Repository) CreateOrGetUtilityApplication(userID uint) (ds.UtilityApplication, error) {
	app, err := r.GetUtilityApplicationDraft(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return r.CreateUtilityApplication(userID)
		}
		return ds.UtilityApplication{}, err
	}
	return app, nil
}

// Получить заявку по ID со связанными услугами и посчитать итог (как GetUtility__Application)
func (r *Repository) GetUtilityApplicationByID(appID uint) (ds.UtilityApplication, error) {
	var app ds.UtilityApplication
	if err := r.db.
		Where("id = ? AND status <> ?", appID, ds.DELETED).
		Preload("Services.Service").
		First(&app).Error; err != nil {
		return ds.UtilityApplication{}, err
	}

	var total float32
	for _, s := range app.Services {
		total += s.Total
	}
	app.TotalCost = total
	return app, nil
}

// Добавить услугу в заявку (аналог AddComponentInSystemCalc)
func (r *Repository) AddServiceToApplication(appID uint, serviceID uint32, quantity float32) error {
	// Проверим саму заявку
	var app ds.UtilityApplication
	if err := r.db.First(&app, "id = ?", appID).Error; err != nil {
		return err
	}

	// Проверим справочник услуг
	var svc ds.UtilityService
	if err := r.db.Where("id = ? AND is_deleted = ?", serviceID, false).First(&svc).Error; err != nil {
		return err
	}

	// Проверим, нет ли уже записи (композитный PK)
	var existing ds.UtilityApplicationService
	check := r.db.Where(
		"utility_application_id = ? AND utility_service_id = ?",
		appID, serviceID,
	).First(&existing)

	if check.Error == nil {
		// уже добавлено — ничего не делаем (как в AddComponentInSystemCalc)
		return nil
	}
	if check.Error != nil && check.Error != gorm.ErrRecordNotFound {
		return check.Error
	}

	// Создаём строку заявки
	item := ds.UtilityApplicationService{
		UtilityApplicationID: appID,
		UtilityServiceID:     serviceID,
		Quantity:             quantity,
		Total:                quantity * svc.Tariff,
	}
	return r.db.Create(&item).Error
}

// Удалить услугу из заявки (по двум ключам)
// В твоём примере для компонентов была попытка удалить через поиск, я сделал просто и надёжно.
func (r *Repository) RemoveServiceFromApplication(appID uint, serviceID uint32) error {
	return r.db.Delete(&ds.UtilityApplicationService{},
		"utility_application_id = ? AND utility_service_id = ?",
		appID, serviceID,
	).Error
}

func (r *Repository) DeleteUtilityApplication(appID uint) error {
	res := r.db.Exec(
		"UPDATE utility_applications SET status = ? WHERE id = ? AND status <> ?",
		ds.DELETED, appID, ds.DELETED,
	)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
