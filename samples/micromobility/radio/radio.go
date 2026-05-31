// Package radio defines the cellular infrastructure observed through the
// network tap: modems, eSIM profiles, and cell attachments.
//
// The observation platform taps into the cellular network stack to see
// signaling events (attach, detach, handover), DNS queries, and IP session
// state. This is a real-time, high-frequency data source anchored on the
// modem's IMEI. It provides network-layer identity and the raw material for
// operator and model classification.
//
// Modem and [fleet.Vehicle] are separate entities observed through independent
// channels (network layer vs. application layer). For cellular-connected
// scooters both are present and correlated; for phone-bridged scooters the
// Vehicle may exist without a stable Modem anchor.
package radio

import "net/netip"

// Modem is a cellular modem identified by its 15-digit IMEI.
//
// Scooters with built-in cellular connectivity carry two permanent hardware
// identifiers: an IMEI (the modem) and an EID (the eUICC chip). The active
// eSIM profile adds an ICCID and IMSI; these change when the operator
// remotely provisions a new profile (e.g. to switch carriers). The IP
// address is ephemeral, reassigned on every new data session.
type Modem struct {
	IMEI        string      // Primary key. Hardware modem identity, etched at manufacture.
	EID         string      // eUICC chip identity. Permanent; changes only if hardware is replaced.
	ICCID       string      // Active eSIM profile identifier. Changes on reprovisioning.
	IMSI        string      // Subscriber identity used by the network for authentication and routing.
	IP          netip.Addr  // Current data-session IP address. Retracted when the session drops.
	ServingCell ServingCell // Current cell attachment as seen by the network tap.
	Firmware    string      // Modem firmware version.
}

// ServingCell holds the network tap's observation of a modem's cell attachment.
//
// This data is exclusively network-side: the scooter's own software has no
// access to radio-level information (serving cell, signal quality, neighbor
// cells). The platform observes it through signaling events on the IMSI.
type ServingCell struct {
	PLMN           string // Public Land Mobile Network identifier (MCC-MNC).
	TrackingArea   uint16 // Tracking area code.
	CellID         uint32 // E-UTRAN cell identity.
	SignalStrength int    // Reference signal received power (RSRP) in dBm.
}
