package extrinsic

// ExtrinsicController is a struct that contains all the extrinsics
// need to take care of the pointers
type ExtrinsicController struct {
	TicketController         *TicketController
	PreimageController       *PreimageController
	GuaranteeController      *GuaranteeController
	DisputeController        *DisputeController
	AvailAssuranceController *AvailAssuranceController
}

// NewExtrinsicController returns a new ExtrinsicController
func NewExtrinsicController(ticketController *TicketController, preimageController *PreimageController,
	guaranteeController *GuaranteeController, disputeController *DisputeController, availAssuranceController *AvailAssuranceController) *ExtrinsicController {
	return &ExtrinsicController{
		TicketController:         ticketController,
		PreimageController:       preimageController,
		GuaranteeController:      guaranteeController,
		DisputeController:        disputeController,
		AvailAssuranceController: availAssuranceController,
	}
}
