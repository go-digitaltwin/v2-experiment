// Package fingerprint defines reusable types for classification assertions in
// digital twin domain models.
//
// Classification entities represent competing inferences: multiple classifiers
// may simultaneously assert different values for the same entity. Domain models
// embed [Classification] to standardize the fingerprint metadata pattern.
//
// [Classification] is designed for embedding (anonymous field). When embedded,
// its fields are promoted to the parent entity:
//
//   - [Classification.Source] becomes part of the composite primary key
//     (alongside the entity's own foreign key). A [Signature] identifies a
//     specific classifier by type and deployment: two classifiers with
//     different sources can assert different values for the same entity.
//   - [Classification.Active], [Classification.Confidence], and
//     [Classification.TTL] are regular properties with individual delta
//     operations.
//
// [Signature] is a named struct field: it is provided as a whole value when
// used as a key component in a delta constructor, and replaced atomically
// when used as a property.
package fingerprint

import "time"

// Classification holds metadata common to all classification assertions.
//
// Domain model types embed this struct to gain the standard fingerprint fields.
// Because it is an anonymous (embedded) field, [Source] is promoted to the
// parent entity and serves as part of the composite primary key alongside the
// entity's own foreign key.
//
// Scalar fields (Active, Confidence, TTL) each get their own delta operations.
type Classification struct {
	// Source identifies which classifier produced this assertion. The full
	// signature (Type + Instance) forms part of the composite primary key:
	// different classifier types, and different deployments of the same type,
	// each maintain their own row for the same entity.
	Source Signature

	// Active distinguishes production classifiers from those under evaluation.
	// Downstream actions (alerts, policy enforcement) typically filter on
	// Active == true. Keeping evaluation classifiers visible in the same table
	// enables side-by-side comparison without a separate pipeline.
	Active bool

	// Confidence is the classifier's self-reported certainty in the range
	// [0.0, 1.0]. Downstream consumers can threshold or rank by confidence
	// when multiple classifiers disagree.
	Confidence float64

	// TTL is the validity window for this assertion. If the classifier does
	// not re-assert within this duration, the classification should be
	// considered stale. How staleness is measured and enforced is up to the
	// application. A zero value means the assertion does not expire.
	TTL time.Duration
}

// Signature identifies a specific classifier by type and deployment. The full
// value (Type + Instance) is the classifier's identity within the domain
// model: each unique signature produces its own classification row.
//
// Type is the stable classifier name; Instance is a catch-all string for
// deployment-specific taxonomy that may become more structured over time
// (version tags, region identifiers, build hashes, experiment labels).
type Signature struct {
	// Type identifies the kind of classifier: a stable name known to the
	// system that groups all deployments of the same classification logic.
	//
	// Examples: "dns-fingerprint", "imsi-range", "payload-analysis".
	Type string

	// Instance identifies a specific deployment, version, or configuration of
	// the classifier named by Type. This is intentionally a free-form string
	// to accommodate the various ways classifier deployments are versioned.
	//
	// Examples: "v2.3", "prod-us-east-1", "experimental-2026Q1".
	Instance string
}
