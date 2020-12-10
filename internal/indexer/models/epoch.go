package models

type Epoch struct {
	Epoch                   uint64 `gorm:"primarykey"`
	Slot1                   Slot   `gorm:"foreignkey:Slot"`
	Slot2                   Slot   `gorm:"foreignkey:Slot"`
	Slot3                   Slot   `gorm:"foreignkey:Slot"`
	Slot4                   Slot   `gorm:"foreignkey:Slot"`
	Slot5                   Slot   `gorm:"foreignkey:Slot"`
	ParticipationPercentage uint64
	Finalized               bool
	Justified               bool
	Randao                  [32]byte
}
