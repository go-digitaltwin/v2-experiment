// Package fleet defines the physical assets observed through application-layer
// telemetry: vehicles (scooters), batteries, and their sensors.
//
// A shared electric scooter is a small vehicle built around a control board
// that reads sensors (GPS, accelerometer, BMS) and reports telemetry through
// whatever network interface is available. The observation platform learns
// about vehicles from the telemetry payloads they emit, not from operator
// fleet management systems.
package fleet

import "net/netip"

// Vehicle is an electric scooter identified by its frame-stamped VIN.
//
// The platform observes vehicles through application-layer telemetry: periodic
// heartbeats when idle (60-120 s) and high-frequency beacons during rides
// (2-5 s). Ephemeral properties (Speed, GNSS, InRide) are retracted when
// telemetry goes silent beyond a configured TTL; stable associations persist
// as historical record.
type Vehicle struct {
	VIN           string  // Primary key. Frame identity stamped at manufacture.
	ModemIMEI     string  // Foreign key → [radio.Modem]. Which cellular modem is installed.
	BatterySerial string  // Foreign key → [Battery]. Which battery pack is installed.
	Speed         float64 // Ground speed in km/h, derived from GNSS.
	Heading       float64 // Course over ground in degrees, 0-360.
	GNSS          GNSS    // Satellite fix from the scooter's GPS receiver.
	InRide        bool    // Whether a ride is in progress (platform-inferred from telemetry pattern).
	Odometer      float64 // Cumulative trip distance in km.
	AccelMotion   bool    // Accelerometer motion-detection flag.
	AccelTilt     float64 // Tilt angle in degrees (0 = upright, 90 = on its side).
	Access        Access  // Current point of attachment to the wider network.
}

// GNSS holds a satellite fix from the scooter's GPS receiver.
//
// This is a wide struct: "last reported GNSS fix" means the entire value
// including precision metadata, not just coordinates. A zero-value GNSS
// represents no fix (the receiver has not acquired satellites).
type GNSS struct {
	Lat  float64 // WGS 84 latitude.
	Lng  float64 // WGS 84 longitude.
	Alt  float64 // Altitude in meters above sea level.
	HDOP float64 // Horizontal dilution of precision (lower is better).
	Sats uint8   // Number of satellites used in the fix.
	Fix  string  // Fix quality: "none", "2d", or "3d".
}

// Fix quality constants for [GNSS.Fix].
const (
	FixNone = "none"
	Fix2D   = "2d"
	Fix3D   = "3d"
)

// Access describes a scooter's current point of attachment to the wider
// network: the interface through which telemetry reaches the operator's
// backend.
//
// These fields are self-reported by the scooter's application software:
// the IP is read from the OS network stack, and the modem identifiers are
// queried via the AT command interface when the firmware supports it.
// Access values are confirmatory, not authoritative: the network tap's
// view of the same modem ([radio.Modem]) is the source of truth for
// network-layer identity. The two observations arrive through independent
// channels and may disagree temporarily (e.g., after an IP change that the
// network tap sees before the next telemetry beacon).
//
// The connectivity model varies by operator and hardware generation. A
// scooter with a built-in cellular modem maintains a persistent data session
// even when idle; a phone-bridged scooter communicates only during rides
// through the rider's phone; a docked scooter uploads buffered telemetry
// over station WiFi.
//
// The Type field selects the variant; fields belonging to other variants are
// at their zero values.
//
// IP is a deliberate simplification. A real cellular session may carry
// multiple addresses at once: an IPv4 and an IPv6 from a dual-stack PDP
// context, or several IPs across multiple PDP contexts (e.g. a fleet-side
// management APN alongside a payload APN). Modeling all of them would
// require a slice and a way to attribute each address to its PDP context.
// This sample keeps a single primary IP because the narrative does not
// exercise multi-context scenarios; a production model would lift this
// restriction.
type Access struct {
	Type  string     // Connectivity variant: "cellular", "bluetooth", or "station".
	IP    netip.Addr // Cellular: current data-session IP from the OS network stack.
	IMEI  string     // Cellular: modem IMEI, if the firmware queries the AT interface.
	IMSI  string     // Cellular: subscriber identity, if available via AT+CIMI or equivalent.
	BLE   string     // Bluetooth: rider's phone BLE device name (Local Name).
	BSSID string     // Station: WiFi BSSID or operator-assigned dock ID.
}

// Access type constants for [Access.Type].
const (
	AccessCellular  = "cellular"
	AccessBluetooth = "bluetooth"
	AccessStation   = "station"
)
