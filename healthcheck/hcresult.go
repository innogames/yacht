package healthcheck

// HCResultMsg is used to send result back to LB Node
type HCResultMsg struct {
	result HCResult
	HC     HealthCheck
}

// HCResult represents result of a healthcheck, see below for know enum variables
type HCResult int

const (
	// HCUnknown is the default value
	HCUnknown HCResult = iota
	// HCError means failure in check itself
	HCError
	// HCBad means the check has failed
	HCBad
	// HCGood means the check has succeeded
	HCGood
)

// HCResultError is used to send data from a specific HC class to HCBase class.
// It combines HCResult with additional error code
type HCResultError struct {
	res HCResult
	err error
}

// HCsResults is used to store last result of many checks.
type HCsResults map[HealthCheck]HCResult

// GoodHCs counts good and all HCs.
func (hcsrs *HCsResults) GoodHCs() (int, int) {
	var allHCs int
	var goodHCs int
	for _, result := range *hcsrs {
		if result == HCGood {
			goodHCs++
		}
		allHCs++
	}
	return goodHCs, allHCs
}

// Update updates HCsResults with information from a HCResultMsg
func (hcsrs *HCsResults) Update(hcrm HCResultMsg) {
	(*hcsrs)[hcrm.HC] = hcrm.result
}
