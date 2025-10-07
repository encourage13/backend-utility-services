package repository

import (
	"fmt"
	"strings"

	"lab1/internal/app/ds"
)

// Список услуг (всё подряд, без флага is_deleted)
func (r *Repository) GetUtilityServices() ([]ds.UtilityService, error) {
	var services []ds.UtilityService
	if err := r.db.
		Model(&ds.UtilityService{}).
		Order("id").
		Find(&services).Error; err != nil {
		return nil, err
	}
	if len(services) == 0 {
		return nil, fmt.Errorf("список услуг пуст")
	}
	return services, nil
}

// Одна услуга по ID
func (r *Repository) GetUtilityServiceByID(id uint32) (ds.UtilityService, error) {
	var s ds.UtilityService
	if err := r.db.
		Model(&ds.UtilityService{}).
		Where("id = ?", id).
		First(&s).Error; err != nil {
		return ds.UtilityService{}, err
	}
	return s, nil
}

// Поиск услуг по названию (ILIKE для Postgres, без is_deleted)
func (r *Repository) SearchUtilityServices(title string) ([]ds.UtilityService, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return r.GetUtilityServices()
	}

	var services []ds.UtilityService
	if err := r.db.
		Model(&ds.UtilityService{}).
		Where("title ILIKE ?", "%"+title+"%").
		Order("id").
		Find(&services).Error; err != nil {
		return nil, err
	}
	return services, nil
}
