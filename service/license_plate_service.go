package service

import (
	"errors"
	"shared-charge/models"

	"gorm.io/gorm"
)

// LicensePlateService 车牌号服务
type LicensePlateService struct{}

// NewLicensePlateService 创建车牌号服务实例
func NewLicensePlateService() *LicensePlateService {
	return &LicensePlateService{}
}

// GetUserLicensePlates 获取用户的车牌号列表
func (s *LicensePlateService) GetUserLicensePlates(userID uint) ([]models.LicensePlate, error) {
	var licensePlates []models.LicensePlate
	
	err := models.DB.Where("user_id = ?", userID).
		Order("is_default DESC, created_at ASC").
		Find(&licensePlates).Error
	
	return licensePlates, err
}

// CreateLicensePlate 创建车牌号
func (s *LicensePlateService) CreateLicensePlate(userID uint, plateNumber string, isDefault bool) (*models.LicensePlate, error) {
	// 检查车牌号格式（简单验证）
	if len(plateNumber) < 6 || len(plateNumber) > 10 {
		return nil, errors.New("车牌号格式不正确")
	}

	// 检查是否已存在相同车牌号
	var existingPlate models.LicensePlate
			err := models.DB.Where("user_id = ? AND plate_number = ?", userID, plateNumber).First(&existingPlate).Error
	if err == nil {
		return nil, errors.New("该车牌号已存在")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 如果用户还没有车牌号，自动设为默认
	var count int64
			models.DB.Model(&models.LicensePlate{}).Where("user_id = ?", userID).Count(&count)
	if count == 0 {
		isDefault = true
	}

	licensePlate := &models.LicensePlate{
		UserID:      userID,
		PlateNumber: plateNumber,
		IsDefault:   isDefault,
	}

	err = models.DB.Create(licensePlate).Error
	return licensePlate, err
}

// UpdateLicensePlate 更新车牌号
func (s *LicensePlateService) UpdateLicensePlate(userID uint, plateID uint, plateNumber string) error {
	// 检查车牌号格式
	if len(plateNumber) < 6 || len(plateNumber) > 10 {
		return errors.New("车牌号格式不正确")
	}

	// 检查车牌号是否属于当前用户
	var licensePlate models.LicensePlate
	err := models.DB.Where("id = ? AND user_id = ?", plateID, userID).First(&licensePlate).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("车牌号不存在")
		}
		return err
	}

	// 检查是否与其他车牌号重复
	var existingPlate models.LicensePlate
			err = models.DB.Where("user_id = ? AND plate_number = ? AND id != ?", userID, plateNumber, plateID).First(&existingPlate).Error
	if err == nil {
		return errors.New("该车牌号已存在")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return models.DB.Model(&licensePlate).Update("plate_number", plateNumber).Error
}

// DeleteLicensePlate 删除车牌号
func (s *LicensePlateService) DeleteLicensePlate(userID uint, plateID uint) error {
	// 检查车牌号是否属于当前用户
	var licensePlate models.LicensePlate
	err := models.DB.Where("id = ? AND user_id = ?", plateID, userID).First(&licensePlate).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("车牌号不存在")
		}
		return err
	}

	// 检查是否有关联的预约或记录
	var reservationCount, recordCount int64
			models.DB.Model(&models.Reservation{}).Where("license_plate_id = ?", plateID).Count(&reservationCount)
		models.DB.Model(&models.Record{}).Where("license_plate_id = ?", plateID).Count(&recordCount)
	
	if reservationCount > 0 || recordCount > 0 {
		return errors.New("该车牌号已有关联的预约或记录，无法删除")
	}

	// 如果删除的是默认车牌号，需要设置其他车牌号为默认
	if licensePlate.IsDefault {
		var otherPlate models.LicensePlate
		err = models.DB.Where("user_id = ? AND id != ?", userID, plateID).First(&otherPlate).Error
		if err == nil {
			models.DB.Model(&otherPlate).Update("is_default", true)
		}
	}

	return models.DB.Delete(&licensePlate).Error
}

// SetDefaultLicensePlate 设置默认车牌号
func (s *LicensePlateService) SetDefaultLicensePlate(userID uint, plateID uint) error {
	// 检查车牌号是否属于当前用户
	var licensePlate models.LicensePlate
	err := models.DB.Where("id = ? AND user_id = ?", plateID, userID).First(&licensePlate).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("车牌号不存在")
		}
		return err
	}

	// 使用事务确保数据一致性
	return models.DB.Transaction(func(tx *gorm.DB) error {
		// 将该用户的所有车牌号设为非默认
		if err := tx.Model(&models.LicensePlate{}).
			Where("user_id = ?", userID).
			Update("is_default", false).Error; err != nil {
			return err
		}

		// 设置指定车牌号为默认
		return tx.Model(&licensePlate).Update("is_default", true).Error
	})
}

// GetDefaultLicensePlate 获取用户的默认车牌号
func (s *LicensePlateService) GetDefaultLicensePlate(userID uint) (*models.LicensePlate, error) {
	var licensePlate models.LicensePlate
	err := models.DB.Where("user_id = ? AND is_default = ?", userID, true).First(&licensePlate).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 没有默认车牌号
		}
		return nil, err
	}
	return &licensePlate, nil
}

// GetLicensePlateByID 根据ID获取车牌号
func (s *LicensePlateService) GetLicensePlateByID(userID uint, plateID uint) (*models.LicensePlate, error) {
	var licensePlate models.LicensePlate
	err := models.DB.Where("id = ? AND user_id = ?", plateID, userID).First(&licensePlate).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("车牌号不存在")
		}
		return nil, err
	}
	return &licensePlate, nil
}
