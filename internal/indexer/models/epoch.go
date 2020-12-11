package models

type Epoch struct {
	Epoch                   uint64 `gorm:"primaryKey"`
	Slot1                   Slot   `gorm:"foreignKey:Slot"`
	Slot2                   Slot   `gorm:"foreignKey:Slot"`
	Slot3                   Slot   `gorm:"foreignKey:Slot"`
	Slot4                   Slot   `gorm:"foreignKey:Slot"`
	Slot5                   Slot   `gorm:"foreignKey:Slot"`
	ParticipationPercentage uint64
	Finalized               bool
	Justified               bool
	Randao                  string
}
