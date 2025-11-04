package repository

import (
	"fmt"
	"lab1/internal/app/ds"
	"lab1/internal/app/role"
	"time"

	"gorm.io/gorm"
)

func (r *Repository) CreateUtilityApplication(userID uint) (ds.UtilityApplication, error) {
	app := ds.UtilityApplication{
		UserID:      userID,
		Status:      string(ds.DRAFT),
		DateCreated: time.Now(),
	}
	if err := r.db.Create(&app).Error; err != nil {
		return ds.UtilityApplication{}, err
	}
	return app, nil
}

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

	if app.TotalCost != total {
		if err := r.db.Model(&ds.UtilityApplication{}).
			Where("id = ?", appID).
			Update("total_cost", total).Error; err != nil {
			return ds.UtilityApplication{}, err
		}
		app.TotalCost = total
	}

	return app, nil
}

func (r *Repository) AddServiceToApplication(appID uint, serviceID uint32, quantity float32) error {
	var app ds.UtilityApplication
	if err := r.db.First(&app, "id = ?", appID).Error; err != nil {
		return err
	}

	var svc ds.UtilityService
	if err := r.db.Where("id = ?", serviceID).First(&svc).Error; err != nil {
		return err
	}

	var existing ds.UtilityApplicationService
	check := r.db.Where(
		"utility_application_id = ? AND utility_service_id = ?",
		appID, serviceID,
	).First(&existing)

	if check.Error == nil {
		existing.Quantity += quantity
		existing.Total = existing.Quantity * svc.Tariff
		return r.db.Save(&existing).Error
	}
	if check.Error != nil && check.Error != gorm.ErrRecordNotFound {
		return check.Error
	}

	item := ds.UtilityApplicationService{
		UtilityApplicationID: appID,
		UtilityServiceID:     serviceID,
		Quantity:             quantity,
		Total:                quantity * svc.Tariff,
	}
	return r.db.Create(&item).Error
}

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

func (r *Repository) GetDraftApplication(userID uint) (ds.UtilityApplication, error) {
	var app ds.UtilityApplication
	err := r.db.Where("user_id = ? AND status = ?", userID, ds.DRAFT).First(&app).Error
	if err != nil {
		return ds.UtilityApplication{}, err
	}
	return app, nil
}

func (r *Repository) GetApplicationWithServices(appID uint) (ds.UtilityApplication, error) {
	var app ds.UtilityApplication
	err := r.db.Preload("Services.Service").Preload("User").Preload("Moderator").First(&app, appID).Error
	if err != nil {
		return ds.UtilityApplication{}, err
	}

	if app.Status == string(ds.DELETED) {
		return ds.UtilityApplication{}, gorm.ErrRecordNotFound
	}

	return app, nil
}

func (r *Repository) GetApplicationsFiltered(status, from, to string) ([]ds.UtilityApplication, error) {
	var applications []ds.UtilityApplication
	query := r.db.Preload("User").Preload("Moderator")

	query = query.Where("status != ?", ds.DELETED)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if from != "" {
		if fromTime, err := time.Parse("2006-01-02", from); err == nil {
			query = query.Where("(status = ? AND date_created >= ?) OR (status != ? AND date_formed >= ?)",
				ds.DRAFT, fromTime, ds.DRAFT, fromTime)
		}
	}

	if to != "" {
		if toTime, err := time.Parse("2006-01-02", to); err == nil {
			query = query.Where("(status = ? AND date_created <= ?) OR (status != ? AND date_formed <= ?)",
				ds.DRAFT, toTime, ds.DRAFT, toTime)
		}
	}

	if err := query.Order("id DESC").Find(&applications).Error; err != nil {
		return nil, err
	}

	return applications, nil
}

func (r *Repository) UpdateApplicationUserFields(id uint, req ds.UtilityApplicationUpdateRequest) error {
	updates := make(map[string]interface{})

	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.TotalCost != nil {
		updates["total_cost"] = *req.TotalCost
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}
	if req.ModeratorID != nil {
		updates["moderator_id"] = *req.ModeratorID
	}

	if len(updates) == 0 {
		return nil
	}

	return r.db.Model(&ds.UtilityApplication{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) FormApplication(id uint, userID uint) error {
	var app ds.UtilityApplication
	if err := r.db.First(&app, id).Error; err != nil {
		return err
	}

	if app.UserID != userID {
		return fmt.Errorf("only creator can form application")
	}

	if app.Status != string(ds.DRAFT) {
		return fmt.Errorf("only draft applications can be formed")
	}

	now := time.Now()
	return r.db.Model(&app).Updates(map[string]interface{}{
		"status":      ds.FORMED,
		"date_formed": now,
	}).Error
}

func (r *Repository) ResolveApplication(id uint, moderatorID uint, action string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var app ds.UtilityApplication
		if err := tx.Preload("Services.Service").First(&app, id).Error; err != nil {
			return err
		}

		if app.Status != string(ds.FORMED) {
			return fmt.Errorf("only formed applications can be resolved")
		}

		var moderator ds.User
		if err := tx.First(&moderator, moderatorID).Error; err != nil {
			return err
		}
		if moderator.Role != role.Manager && moderator.Role != role.Admin {
			return fmt.Errorf("only moderators can resolve applications")
		}

		now := time.Now()
		updates := map[string]interface{}{
			"moderator_id":  moderatorID,
			"date_accepted": now,
		}

		switch action {
		case "COMPLETED":
			updates["status"] = ds.COMPLETED

			var totalCost float32
			for _, service := range app.Services {
				totalCost += service.Total
			}
			updates["total_cost"] = totalCost

		case "REJECTED":
			updates["status"] = ds.REJECTED
		default:
			return fmt.Errorf("invalid action, must be COMPLETED or REJECTED")
		}

		return tx.Model(&app).Updates(updates).Error
	})
}

func (r *Repository) LogicallyDeleteApplication(appID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var app ds.UtilityApplication

		if err := tx.Preload("Services").First(&app, appID).Error; err != nil {
			return err
		}

		updates := map[string]interface{}{
			"status":      ds.DELETED,
			"date_formed": time.Now(),
		}

		if err := tx.Model(&ds.UtilityApplication{}).Where("id = ?", appID).Updates(updates).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *Repository) UpdateApplicationService(appID uint, serviceID uint32, req ds.UtilityApplicationServiceUpdateRequest) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var link ds.UtilityApplicationService
		if err := tx.Where("utility_application_id = ? AND utility_service_id = ?", appID, serviceID).
			First(&link).Error; err != nil {
			return fmt.Errorf("service not found in application: %w", err)
		}

		var service ds.UtilityService
		if err := tx.First(&service, serviceID).Error; err != nil {
			return fmt.Errorf("service not found: %w", err)
		}

		updates := make(map[string]interface{})

		if req.Quantity != nil {
			updates["quantity"] = *req.Quantity
		}

		if req.Tariff != nil {
			updates["custom_tariff"] = *req.Tariff
		}

		finalTariff := service.Tariff
		if req.Tariff != nil {
			finalTariff = *req.Tariff
		}

		finalQuantity := link.Quantity
		if req.Quantity != nil {
			finalQuantity = *req.Quantity
		}

		updates["total"] = finalTariff * finalQuantity

		if len(updates) == 0 {
			return fmt.Errorf("no valid fields to update")
		}

		if err := tx.Model(&link).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update service in application: %w", err)
		}

		var totalCost float32
		if err := tx.Model(&ds.UtilityApplicationService{}).
			Where("utility_application_id = ?", appID).
			Select("COALESCE(SUM(total), 0)").
			Scan(&totalCost).Error; err != nil {
			return fmt.Errorf("failed to calculate total cost: %w", err)
		}

		if err := tx.Model(&ds.UtilityApplication{}).
			Where("id = ?", appID).
			Update("total_cost", totalCost).Error; err != nil {
			return fmt.Errorf("failed to update application total cost: %w", err)
		}

		return nil
	})
}
