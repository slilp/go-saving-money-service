package repository

import (
	"gorm.io/gorm"
)

type plannerRepositoryDB struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) PlannerRepository {
	db.AutoMigrate(&PlannerEntity{})
	return plannerRepositoryDB{db}
}

func (r plannerRepositoryDB) Create(planner PlannerEntity) error {
	return r.db.Create(&planner).Error
}

func (r plannerRepositoryDB) Delete(id uint) error {
	return r.db.Delete(&PlannerEntity{}, id).Error
}

func (r plannerRepositoryDB) Update(planner PlannerEntity) error {
	return r.db.Model(&planner).Updates(planner).Error
}
