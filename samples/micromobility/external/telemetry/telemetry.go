// Package telemetry is the as-delivered representation of the application-layer
// telemetry a scooter's firmware emits: the raw report payload as it crosses
// into the platform from the outside, before any digestion.
//
// A [Report] mirrors the wire format the feed exports. It is deliberately
// faithful to the source, not digested: it keeps the full envelope (sequence,
// timestamp, trigger), the redundant battery energies, and the self-reported
// network identity in [Access]. Nothing is discarded. The digested twin
// (fleet.Vehicle, fleet.Battery, and the network-tap-correlated radio.Modem)
// is derived from these reports; the raw fields are retained here for
// downstream battery-health analysis, corroboration against the network tap,
// and replay.
//
// Reports are keyed on [Report.DeviceID] — the firmware's self-assigned
// identity, a VIN or an operator-assigned fleet id — not the resolved twin key.
// Mapping a DeviceID to a fleet.Vehicle VIN is an identity-resolution step
// performed during digestion, using the vendor instrument.
package telemetry

import (
	"net/netip"
	"time"
)

// Report is one telemetry payload as delivered by the feed.
//
// The sensor sub-objects are pointers because the firmware omits what it cannot
// populate: there is no GNSS on a report sent before the receiver has a fix,
// and a battery-swap report carries only the new pack's reading. A non-nil
// pointer means the firmware reported that sensor in this payload.
type Report struct {
	DeviceID string    `json:"device_id"` // Firmware identity: a VIN or operator-assigned fleet id.
	Time     time.Time `json:"ts"`        // Timestamp at the scooter's clock (ISO 8601).
	Seq      uint32    `json:"seq"`       // Monotonic sequence number; gaps reveal lost reports.
	Trigger  string    `json:"trigger"`   // Why this report was sent. See the Trigger constants.

	GNSS     *GNSS    `json:"gnss,omitempty"`
	Speed    float64  `json:"speed"`              // Ground speed in km/h (GNSS-derived).
	Heading  float64  `json:"heading"`            // Course over ground in degrees, 0-360.
	Odometer *float64 `json:"odometer,omitempty"` // Cumulative trip distance in km; ride reports only.
	Battery  *Battery `json:"battery,omitempty"`
	Access   *Access  `json:"access,omitempty"`
	Accel    *Accel   `json:"accel,omitempty"`
}

// GNSS is the satellite-fix sub-object. An absent Report.GNSS pointer or a Fix
// of [FixNone] means the receiver had no positional fix.
type GNSS struct {
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
	Alt  float64 `json:"alt,omitempty"`  // Altitude in meters above sea level.
	HDOP float64 `json:"hdop,omitempty"` // Horizontal dilution of precision (lower is better).
	Sats uint8   `json:"sats,omitempty"` // Satellites used in the fix.
	Fix  string  `json:"fix"`            // Fix quality. See the Fix constants.
}

// Fix quality values for [GNSS.Fix].
const (
	FixNone = "none"
	Fix2D   = "2d"
	Fix3D   = "3d"
)

// Battery is the raw BMS sub-object. It is intentionally wide and redundant:
// Energy divided by FullEnergy is the state of charge, and FullEnergy trends
// down from DesignEnergy as cycles accumulate. Real BMSs expose different
// subsets of these readings. The platform digests whichever measurements arrive
// into a single normalized state of charge on fleet.Battery, keeping the raw
// energies here for downstream battery-health analysis.
type Battery struct {
	ID           string  `json:"id"`                      // BMS-reported serial of the installed pack.
	SoC          float64 `json:"soc,omitempty"`           // State of charge, 0.0-1.0, when reported directly.
	Voltage      float64 `json:"voltage,omitempty"`       // Pack voltage in volts.
	DesignEnergy uint32  `json:"design_energy,omitempty"` // Factory full-charge energy when new, in mWh.
	FullEnergy   uint32  `json:"full_energy,omitempty"`   // Latest measured full-charge capacity, in mWh.
	Energy       uint32  `json:"energy,omitempty"`        // Current stored energy, in mWh.
	Power        int     `json:"power,omitempty"`         // Instantaneous power in mW (+ charging, - load).
	Temp         int     `json:"temp,omitempty"`          // Cell temperature in degrees Celsius.
	Cycles       uint32  `json:"cycles,omitempty"`        // Lifetime charge-cycle count.
}

// Access is the self-reported point-of-attachment sub-object: the interface
// through which the firmware believes its telemetry reaches the backend.
//
// These values are the firmware's own view: the IP is read from the OS network
// stack, the modem identifiers via the AT command interface when the firmware
// supports it. They are confirmatory, not authoritative — the network tap's
// view of the same modem (radio.Modem) is the source of truth for network-layer
// identity, and the two may disagree temporarily (e.g. after an IP change the
// tap sees before the next beacon). The digest keeps only the attachment mode
// and the application-unique anchors (BLE name, dock id) on fleet.Vehicle; the
// raw network identity here is retained for corroboration against the network
// tap.
//
// Type selects the variant; fields of other variants are at their zero values.
type Access struct {
	Type  string     `json:"type"`            // Connectivity variant. See the Access constants.
	IP    netip.Addr `json:"ip"`              // Cellular: data-session IP from the OS network stack.
	IMEI  string     `json:"imei,omitempty"`  // Cellular: modem IMEI via AT, if the firmware queries it.
	IMSI  string     `json:"imsi,omitempty"`  // Cellular: subscriber identity via AT, if available.
	BLE   string     `json:"ble,omitempty"`   // Bluetooth: rider's phone BLE device name (Local Name).
	BSSID string     `json:"bssid,omitempty"` // Station: WiFi BSSID or operator-assigned dock id.
}

// Access variants for [Access.Type].
const (
	AccessCellular  = "cellular"
	AccessBluetooth = "bluetooth"
	AccessStation   = "station"
)

// Accel is the accelerometer sub-object.
type Accel struct {
	Motion bool    `json:"motion"`         // Motion-detection flag.
	Tilt   float64 `json:"tilt,omitempty"` // Tilt angle in degrees (0 = upright, 90 = on its side).
}

// Trigger values for [Report.Trigger]. A report is either periodic (fired on a
// timer) or edge-triggered (fired by a state change the firmware detects).
const (
	TriggerPeriodic     = "periodic"      // Idle or in-ride heartbeat, on a timer.
	TriggerMotionStart  = "motion_start"  // Accelerometer went from stationary to moving.
	TriggerMotionStop   = "motion_stop"   // Accelerometer went from moving to stationary.
	TriggerImpact       = "impact"        // Accelerometer spike past the impact threshold.
	TriggerBatterySwap  = "battery_swap"  // BMS reports a different serial after a power cycle.
	TriggerLowBattery   = "low_battery"   // State of charge dropped below the configured threshold.
	TriggerAccessUpdate = "access_update" // Point of attachment changed (new IP, phone, or dock).
	TriggerGNSSFix      = "gnss_fix"      // Receiver acquired a fix after a period with none.
	TriggerPowerOn      = "power_on"      // Booted after battery insertion or deep-sleep wake.
)
