package fleet

// Battery is a swappable lithium-ion pack identified by the serial number
// on its BMS chip.
//
// Batteries circulate independently of vehicles: field technicians swap
// depleted packs in the street, so a single battery passes through many
// scooters over its lifetime. The platform learns about batteries from the
// BMS readings in each scooter's telemetry payload. A battery-swap event
// manifests as a sudden change in the reported serial number.
type Battery struct {
	Serial     string  // Primary key. BMS serial number printed on the casing.
	VehicleVIN string  // Foreign key → [Vehicle]. Which scooter this pack is installed in.
	SoC        float64 // State of charge, 0.0 (empty) to 1.0 (full).
	Voltage    float64 // Pack voltage in volts.
	Temp       int     // Cell temperature in degrees Celsius.
	Cycles     uint32  // Lifetime charge-cycle count reported by the BMS.
}
