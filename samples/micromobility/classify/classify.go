// Package classify holds assertions that automated classifiers make about
// modem identity: which operator runs a scooter and what hardware model it is.
//
// The observation platform runs multiple independent classifiers (DNS
// fingerprinting, IMSI-range analysis, payload-structure recognition) that
// may disagree. Both an [Operator] and a [Model] assertion embed
// [fingerprint.Classification], which provides the classifier's [fingerprint.Signature]
// (type and deployment version), an Active flag, and a Confidence score so
// that downstream consumers can reconcile conflicts without forcing premature
// resolution.
package classify

import "github.com/go-digitaltwin/v2-experiment/fingerprint"

// Operator records one classifier's assertion about which fleet operator
// runs the scooter behind a given modem.
//
// Multiple classifiers may assert different operators for the same IMEI
// (e.g., DNS sees lime.com while the IMSI-range maps to Bird). All
// assertions coexist; the composite key (IMEI + Source) keeps them
// separate.
type Operator struct {
	IMEI     string // Key. Modem identity being classified.
	Operator string // Asserted operator name (e.g. "Lime", "Bird", "Dot").
	fingerprint.Classification
}

// Model records one classifier's assertion about the hardware
// manufacturer and model designation behind a given modem.
type Model struct {
	IMEI         string // Key. Modem identity being classified.
	Manufacturer string // Asserted hardware manufacturer (e.g. "Segway").
	Designation  string // Asserted model designation (e.g. "Max G30").
	fingerprint.Classification
}
