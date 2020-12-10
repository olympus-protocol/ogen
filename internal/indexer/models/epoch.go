package models

type Epoch struct {
	Epoch                   int
	Slot1                   Slot
	Slot2                   Slot
	Slot3                   Slot
	Slot4                   Slot
	Slot5                   Slot
	ParticipationPercentage int
	Finalized               bool
	Justified               bool
	Randao                  string
}
