package server

type ConnectionSlotStatus struct {
	maxSlots int
	slotsInUse int
}

func (p *ConnectionSlotStatus) UseFreeSlot() bool {
	if p.hasFreeSlot() {
		p.slotsInUse++

		return true
	}

	return false
}

func (p *ConnectionSlotStatus) hasFreeSlot() bool {
	return p.slotsInUse < p.maxSlots
}

func (p *ConnectionSlotStatus) FreesASlot() {
	if p.slotsInUse == 0 {
		return
	}
	p.slotsInUse--
}

func NewConnectionSlotStatus(maxSlots int) *ConnectionSlotStatus {
	return &ConnectionSlotStatus{maxSlots: maxSlots, slotsInUse: 0}
}

