package server

type ConnectionSlots struct {
	maxSlots int
	slotsInUse int
}

func (p *ConnectionSlots) HasFreeSlot() bool {
	return p.slotsInUse < p.maxSlots
}

func (p *ConnectionSlots) UseFreeSlot() {
	p.slotsInUse++
}

func NewConnectionSlots(maxSlots int) *ConnectionSlots {
	return &ConnectionSlots{maxSlots: maxSlots, slotsInUse: 0}
}

