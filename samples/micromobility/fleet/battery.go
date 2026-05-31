package fleet

// Battery is a swappable lithium-ion pack identified by the serial number
// on its BMS chip.
//
// Batteries circulate independently of vehicles: field technicians swap
// depleted packs in the street, so a single battery passes through many
// scooters over its lifetime. The platform learns about batteries from the
// BMS readings in each scooter's telemetry payload. A battery-swap event
// manifests as a sudden change in the reported serial number.
//
// This is the digested twin state: a single normalized state of charge plus
// the readings the twin needs. The raw BMS sub-object (the external
// telemetry.Battery) is wider, carrying the redundant design, full, and
// instantaneous energies that downstream battery-health analysis consumes.
//
// The Vehicle→Battery relationship is maintained on [Vehicle] (BatterySerial),
// not here. If a reverse link (Battery→Vehicle) is ever needed, it will be
// added as a field with supporting causality mechanisms to keep both sides
// eventually consistent.
type Battery struct {
	Serial  string  // Primary key. BMS serial number printed on the casing.
	SoC     float64 // State of charge, 0.0 (empty) to 1.0 (full).
	Voltage float64 // Pack voltage in volts.
	Temp    int     // Cell temperature in degrees Celsius.
	Cycles  uint32  // Lifetime charge-cycle count reported by the BMS.
}
