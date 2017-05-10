package healthcheck

// Result represents result of a healthcheck, see below for know enum variables
type Result int

const (
	// HCUnknown is the default value
	HCUnknown Result = iota
	// HCError means failure in check itself
	HCError
	// HCBad means the check has failed
	HCBad
	// HCGood means the check has succeeded
	HCGood
)

// ResultError is used to send data from a specific HC class to this master class.
// It combines Result with additional error code
type ResultError struct {
	res Result
	err error
}
