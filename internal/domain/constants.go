package domain

const (
	SetStatusDraft    = "draft"
	SetStatusOnSale   = "on_sale"
	SetStatusReserved = "reserved"
	SetStatusSold     = "sold"
	SetStatusHidden   = "hidden"

	TxStatusReserved        = "reserved"
	TxStatusHandoverWaiting = "handover_waiting"
	TxStatusShipped         = "shipped"
	TxStatusReceived        = "received"
	TxStatusCompleted       = "completed"
	TxStatusCancelled       = "cancelled"

	SessionTypeListing   = "listing_support"
	SessionTypeQuestion  = "set_question"
	SessionTypeSearch    = "search"
	SessionTypeStartPlan = "start_plan"

	ItemConditionNew     = "new"
	ItemConditionLikeNew = "like_new"
	ItemConditionGood    = "good"
	ItemConditionFair    = "fair"
	ItemConditionUnknown = "unknown"
)
