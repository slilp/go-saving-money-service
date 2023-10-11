package repository

import "gorm.io/gorm"

type PlannerEntity struct {
	gorm.Model
	Name   string `gorm:"not null;"`
	Target uint
}

func (PlannerEntity) TableName() string {
	return "planner"
}

type PlannerRepository interface {
	Create(PlannerEntity) error
	Delete(id uint) error
	Update(PlannerEntity) error
}
