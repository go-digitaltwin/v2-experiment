// Package classify holds assertions that automated classifiers make about
// modem identity: which operator runs a scooter and what hardware model it is.
//
// The observation platform runs multiple independent classifiers (DNS
// fingerprinting, IMSI-range analysis, payload-structure recognition) that
// may disagree. Both an Operator and a Model assertion carry a Source, a
// Confidence score, and a classifier Version so that downstream consumers
// can reconcile conflicts without forcing premature resolution.
//
// These types are placeholders; naming and structure will be refined once
// the classification story is fleshed out.
package classify

// Operator records one classifier's assertion about which fleet operator
// runs the scooter behind a given modem.
//
// Multiple classifiers may assert different operators for the same IMEI
// (e.g., DNS sees lime.com while the IMSI-range maps to Bird). All
// assertions coexist; the composite key (IMEI + Source) keeps them
// separate.
type Operator struct {
	IMEI       string  // Key. Modem identity being classified.
	Source     string  // Key. Classifier identity (e.g. "dns-fingerprint", "imsi-range").
	Operator   string  // Asserted operator name (e.g. "Lime", "Bird", "Dot").
	Confidence float64 // Classifier confidence, 0.0 to 1.0.
	Version    string  // Classifier deployment version (e.g. "v2.3").
}

// Model records one classifier's assertion about the hardware
// manufacturer and model designation behind a given modem.
type Model struct {
	IMEI         string  // Key. Modem identity being classified.
	Source       string  // Key. Classifier identity.
	Manufacturer string  // Asserted hardware manufacturer (e.g. "Segway").
	Model        string  // Asserted model designation (e.g. "Max G30").
	Confidence   float64 // Classifier confidence, 0.0 to 1.0.
	Version      string  // Classifier deployment version.
}
