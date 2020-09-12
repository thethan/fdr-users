package entities

type BroadcastType int

const (
	BroadCastTypeDraftOpen BroadcastType = iota
	BroadCastTypeDraftOrder
	BroadCastTypePlayerDrafted
)
