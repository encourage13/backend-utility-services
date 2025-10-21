package repository

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/url"
	"strings"

	"lab1/internal/app/ds"

	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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

func (r *Repository) GetUtilityServicesFiltered(title string) ([]ds.UtilityService, int64, error) {
	var services []ds.UtilityService
	var total int64

	query := r.db.Model(&ds.UtilityService{})
	if title != "" {
		query = query.Where("title ILIKE ?", "%"+title+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	servicesQuery := query.Order("id asc")
	if err := servicesQuery.Find(&services).Error; err != nil {
		return nil, 0, err
	}

	return services, total, nil
}

func (r *Repository) CreateUtilityService(service *ds.UtilityService) error {
	return r.db.Create(service).Error
}

func (r *Repository) UpdateUtilityService(id uint32, req ds.UtilityServiceUpdateRequest) (ds.UtilityService, error) {
	var service ds.UtilityService
	if err := r.db.First(&service, id).Error; err != nil {
		return ds.UtilityService{}, err
	}

	if req.Title != nil {
		service.Title = *req.Title
	}
	if req.Description != nil {
		service.Description = *req.Description
	}
	if req.ImageURL != nil {
		service.ImageURL = *req.ImageURL
	}
	if req.Unit != nil {
		service.Unit = *req.Unit
	}
	if req.Tariff != nil {
		service.Tariff = *req.Tariff
	}

	if err := r.db.Save(&service).Error; err != nil {
		return ds.UtilityService{}, err
	}

	return service, nil
}

func (r *Repository) DeleteUtilityService(id uint32) error {
	var service ds.UtilityService
	var imageURLToDelete string

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&service, id).Error; err != nil {
			return err
		}
		if service.ImageURL != "" {
			imageURLToDelete = service.ImageURL
		}
		if err := tx.Delete(&ds.UtilityService{}, id).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	if imageURLToDelete != "" && r.minioClient != nil {
		parsedURL, err := url.Parse(imageURLToDelete)
		if err != nil {
			return nil
		}

		objectName := strings.TrimPrefix(parsedURL.Path, fmt.Sprintf("/%s/", r.bucketName))
		err = r.minioClient.RemoveObject(context.Background(), r.bucketName, objectName, minio.RemoveObjectOptions{})
		if err != nil {

			fmt.Printf("ERROR: failed to delete object '%s' from MinIO: %v\n", objectName, err)
		}
	}

	return nil
}

func (r *Repository) UploadUtilityServiceImage(serviceID uint32, fileHeader *multipart.FileHeader) (string, error) {
	if r.minioClient == nil {
		return "", fmt.Errorf("MinIO client not initialized")
	}

	var finalImageURL string
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var service ds.UtilityService
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&service, serviceID).Error; err != nil {
			return fmt.Errorf("service with id %d not found: %w", serviceID, err)
		}

		if service.ImageURL != "" {
			oldImageURL, err := url.Parse(service.ImageURL)
			if err == nil {
				oldObjectName := strings.TrimPrefix(oldImageURL.Path, fmt.Sprintf("/%s/", r.bucketName))
				r.minioClient.RemoveObject(context.Background(), r.bucketName, oldObjectName, minio.RemoveObjectOptions{})
			}
		}

		fileName := fileHeader.Filename
		objectName := fileName

		file, err := fileHeader.Open()
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = r.minioClient.PutObject(context.Background(), r.bucketName, objectName, file, fileHeader.Size, minio.PutObjectOptions{
			ContentType: fileHeader.Header.Get("Content-Type"),
		})

		if err != nil {
			return fmt.Errorf("failed to upload to minio: %w", err)
		}

		imageURL := fmt.Sprintf("http://%s/%s/%s", r.minioEndpoint, r.bucketName, objectName)

		if err := tx.Model(&service).Update("image_url", imageURL).Error; err != nil {
			return fmt.Errorf("failed to update service image url in db: %w", err)
		}

		finalImageURL = imageURL
		return nil
	})
	if err != nil {
		return "", err
	}
	return finalImageURL, nil
}
