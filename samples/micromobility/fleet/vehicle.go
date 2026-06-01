// Package fleet defines the physical assets observed through application-layer
// telemetry: vehicles (scooters), batteries, and their sensors.
//
// A shared electric scooter is a small vehicle built around a control board
// that reads sensors (GPS, accelerometer, BMS) and reports telemetry through
// whatever network interface is available. The observation platform learns
// about vehicles from the telemetry payloads they emit, not from operator
// fleet management systems.
//
// The types here are the digested twin state. The raw payloads they are
// derived from arrive from outside the platform and live in the external
// telemetry package as telemetry.Report; the digest normalizes those reports
// into the state the twin needs and drops the envelope and redundant readings
// the twin does not.
package fleet

// Vehicle is an electric scooter identified by its frame-stamped VIN.
//
// The platform observes vehicles through application-layer telemetry: periodic
// heartbeats when idle (60-120 s) and high-frequency beacons during rides
// (2-5 s). Ephemeral properties (Speed, GNSS, InRide) are retracted when
// telemetry goes silent beyond a configured TTL; stable associations persist
// as historical record.
type Vehicle struct {
	VIN           string     // Primary key. Frame identity stamped at manufacture.
	ModemIMEI     string     // Foreign key → [radio.Modem]. Which cellular modem is installed.
	BatterySerial string     // Foreign key → [Battery]. Which battery pack is installed.
	Speed         float64    // Ground speed in km/h, derived from GNSS.
	Heading       float64    // Course over ground in degrees, 0-360.
	GNSS          GNSS       // Satellite fix from the scooter's GPS receiver.
	InRide        bool       // Whether a ride is in progress (platform-inferred from telemetry pattern).
	Odometer      float64    // Cumulative trip distance in km.
	AccelMotion   bool       // Accelerometer motion-detection flag.
	AccelTilt     float64    // Tilt angle in degrees (0 = upright, 90 = on its side).
	Attachment    Attachment // Current point of attachment, digested from telemetry.
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

// Attachment is the digested point of attachment: how the scooter is currently
// reaching the backend. It is derived from the telemetry feed's raw access
// sub-object (telemetry.Access), reduced to twin state.
//
// Only the mode and the application-unique anchors live here. For [AttachCellular],
// the network identity (IP, IMEI, IMSI) is not duplicated onto the vehicle: it
// belongs to the modem and is reached through [Vehicle.ModemIMEI] → [radio.Modem],
// whose network-tap view is authoritative. The telemetry's self-reported copy of
// that identity stays in the external telemetry layer for corroboration against the tap.
// [AttachBluetooth] and [AttachStation] carry anchors the network tap never sees
// (the phone's BLE name, the dock id), so the telemetry feed is their only source.
//
// The connectivity model varies by operator and hardware generation. A scooter
// with a built-in cellular modem maintains a persistent data session even when
// idle; a phone-bridged scooter communicates only during rides through the
// rider's phone; a docked scooter uploads buffered telemetry over station WiFi.
//
// Mode selects the variant; fields belonging to other variants are at their
// zero values.
type Attachment struct {
	Mode   string // Connectivity variant: "cellular", "bluetooth", or "station".
	BLE    string // Bluetooth: rider's phone BLE device name (Local Name).
	DockID string // Station: WiFi BSSID or operator-assigned dock ID.
}

// Attachment mode constants for [Attachment.Mode].
const (
	AttachCellular  = "cellular"
	AttachBluetooth = "bluetooth"
	AttachStation   = "station"
)
