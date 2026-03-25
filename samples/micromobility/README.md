# Shared Electric Scooters

This sample domain models shared electric scooters (Lime, Bird, Dot and similar
operators) as seen from a cellular network observation platform. The platform does
not have access to any operator's fleet management system. Everything it knows
about the scooters, it learns from the network.

## The World

A city has thousands of electric scooters deployed across sidewalks, docking
stations, and street corners. Several operators compete in the same geography:
Lime, Bird, Dot, and others. Each operator manages its own fleet independently,
but all their scooters share the same cellular infrastructure to reach their
backends.

### Scooters

A scooter is a small electric vehicle built around a control board running the
operator's firmware. The board reads sensors (GPS receiver, accelerometer, BMS
interface) and runs the application logic that decides what to report, when to
report it, and how to encode the telemetry payload. Think of it as a small
embedded computer with peripherals.

Connectivity is one of those peripherals. The scooter's software sends its
telemetry through whatever network interface is available: a built-in cellular
modem, a Bluetooth bridge to the rider's phone, WiFi at a docking station, or
some combination. The modem, when present, is a communication channel, not the
brain. It provides an IP interface to the scooter's software much like an
Ethernet adapter on a server.

Scooters carry a VIN stamped on the frame and a QR code for riders to scan. From
the operator's perspective, these are the primary identifiers. From the network's
perspective, the scooter itself is invisible; only the network interface's
identifiers (IMEI, IMSI, IP for cellular; the phone's identifiers for
Bluetooth-bridged scooters) are visible.

### Modems and eSIMs

Scooters that have a built-in cellular modem carry two permanent hardware
identifiers: an IMEI (the modem) and an EID (the eUICC chip that stores eSIM
profiles). The active eSIM profile adds two more: an ICCID (the profile itself)
and an IMSI (the subscriber identity the network uses for authentication and
routing). The network assigns a dynamic IP address when the modem establishes a
data session.

These identifiers have different lifetimes. The IMEI and EID are etched at
manufacture and never change. The ICCID and IMSI are tied to the active eSIM
profile; they change when the operator remotely provisions a new profile, for
example to switch carriers for better regional coverage. The IP address is
ephemeral, reassigned on every new data session.

### Batteries

Scooters run on swappable lithium-ion battery packs. Each pack has a serial
number printed on its casing and a small BMS (battery management system) chip
that reports charge level and cycle count through the scooter's telemetry payload.
Field technicians swap batteries in the street: they pull the depleted pack,
insert a fresh one, and move on. A single battery may pass through dozens of
scooters over its lifetime.

### Riders and Rides

A rider unlocks a scooter through the operator's app, rides to a destination, and
parks it. From the network side, a ride looks like a burst of frequent telemetry
beacons with non-zero speed and changing GPS coordinates. The ride itself is a
business concept managed by the operator's backend; the network never sees a
"ride started" or "ride ended" event directly. It can only infer ride activity
from the telemetry pattern.

### Field Technicians

Operators employ field technicians (sometimes called "juicers" or "chargers") who
drive vans through the city collecting scooters with low batteries, swapping
packs, relocating scooters to high-demand areas, and performing maintenance. From
the network's perspective, a battery swap manifests as a sudden change in the
battery serial number reported in the telemetry payload. A relocation looks like a
GPS jump while the scooter is not in a ride.

## How Scooters Connect

Not all scooters connect the same way. The connectivity model varies by operator,
hardware generation, and cost trade-offs. This variance directly affects what the
observation platform can see and how reliably it can track each scooter.

### Cellular (always-on modem)

The richest observability scenario. The scooter has a built-in cellular modem with
an eSIM. It maintains a persistent data session even when idle: the modem is
always attached to the network, the IP address is allocated, and the scooter's
software sends periodic heartbeats regardless of ride activity. The platform sees
the modem's IMEI and IMSI at the network layer and correlates them with the
application-layer telemetry flowing through the session.

This is the primary scenario explored in this domain model.

### Phone-bridged (Bluetooth relay)

Some operators (or earlier hardware generations) save the cost of a cellular modem
by having the scooter communicate through the rider's phone. The scooter connects
to the phone via Bluetooth Low Energy; the phone relays telemetry to the
operator's backend over its own cellular connection.

From the platform's perspective, the scooter is invisible between rides. During a
ride, the platform sees the *phone's* IMEI, IMSI, and IP, not the scooter's. The
application payload still contains the scooter's device ID and sensor readings,
but the network-layer anchor belongs to the phone and changes with every rider.
The platform can extract telemetry but cannot maintain a persistent identity for
the scooter across rides.

### Hybrid

Some operators use both: a cellular modem for always-on fleet management and a
Bluetooth connection to the rider's phone for real-time ride features (unlock,
lock, in-ride UI). The platform sees both the modem's session and the phone's
session. Correlating the two (same scooter, two network identities) is a
non-trivial classification problem.

### WiFi at dock

A few operators equip docking stations with WiFi. Scooters connect when docked
and upload buffered telemetry in bulk. The platform sees intermittent bursts of
data with long gaps between them. The WiFi MAC address is a per-station
identifier, not a per-scooter one.

### Why This Variance Matters

The observation platform must handle all of these modes. A scooter with an
always-on cellular modem produces a continuous, rich data stream. A
phone-bridged scooter produces ride-time-only fragments with a shifting network
anchor. A WiFi-docked scooter produces delayed batches. The domain model
accommodates this through two mechanisms:

1. **Layered identity.** The Modem and Vehicle are separate entities observed
   through different channels. For cellular-connected scooters, both entities
   exist and are correlated. For phone-bridged scooters, the platform may only
   have the Vehicle (from extracted application payloads) without a stable Modem
   anchor. The model does not require both.

2. **Partial data.** The assert/retract/ignore delta semantics handle gaps
   naturally. The platform asserts what it knows, ignores what it does not, and
   retracts what has gone stale. A phone-bridged scooter with no telemetry
   between rides simply has its ephemeral properties retracted by timeout.

## Instrumentation

The observation platform does not receive data from a single source. Different
instrumentation points provide different slices of information, each with its own
data characteristics, refresh rate, and reliability.

### Vendor Instrument

Operators maintain registries that map physical scooter attributes to fleet
identifiers. A vendor-specific instrument (an integration with the operator's
provisioning system, or a tap into their asset management API) provides:

- **Frame identity**: VIN, hardware generation, manufacture date.
- **Fleet correlation**: the operator-assigned fleet ID (the one encoded in the QR
  code and used for manual unlock via the rider's app).
- **Modem provisioning**: which IMEI is installed in which vehicle, eSIM profile
  assignments.

This data is stable. It changes on provisioning events (new scooter registered),
modem replacements, or fleet reassignment (scooter sold between operators). It is
the source of truth for correlating the physical vehicle (VIN) with the fleet
identity and the modem (IMEI).

### Network Tap

A tap into the cellular network stack provides real-time access to signaling and
DNS traffic flowing between scooters and their operator backends. Depending on the
deployment, this may be a deep packet inspection appliance, a lawful intercept
interface, or an API exposed by a network function. It provides:

- **Signaling events**: modem attach/detach, IP session establishment, handovers,
  serving cell identity.
- **DNS queries**: domain name resolutions that reveal which operator's API the
  scooter contacts (e.g., `api.lime.com`, `device.bird.co`).

This is a real-time, high-frequency data source anchored on the IMEI. It provides
the raw material for network-layer identity (IMEI, IMSI, IP) and for
classification (DNS patterns, IMSI ranges).

### Telemetry Feed

Application-layer telemetry (the structured payloads described in
[Telemetry Events](#telemetry-events)) does not necessarily come from the same
network tap. While packet inspection can intercept telemetry in transit, a more
common arrangement is for the platform to receive telemetry in batch from each
fleet operator's backend: the operator collects reports from its scooters and
periodically exports them.

This means telemetry may arrive with different latency and completeness depending
on the operator's export cadence and data-sharing agreement. Some operators stream
near-real-time; others provide hourly or daily batches. The platform treats all
telemetry the same way (as delta assertions keyed on device_id), regardless of how
it was delivered.

### What Each Instrument Contributes

| Instrument | Data | Refresh | Identity anchor |
|------------|------|---------|-----------------|
| **Vendor instrument** | VIN, fleet ID, hardware generation, modem assignment | On provisioning events | VIN (vehicle frame) |
| **Network tap** | IMEI, IMSI, IP, serving cell, DNS queries | On network events (real-time) | IMEI (modem) |
| **Telemetry feed** | Device ID, GNSS, speed, battery, accelerometer, access point | Per report (real-time or batched) | device_id (fleet ID or VIN) |

These instruments observe through independent channels. For cellular-connected
scooters, the platform correlates them: the vendor instrument says "VIN
`WMX00042` has IMEI `353456789012345`," the network tap sees that IMEI attach to
the network, and the telemetry feed delivers application payloads keyed on the
same device. For phone-bridged scooters, only the telemetry feed (relayed through
the operator's backend) may be available; the vendor instrument provides the
stable identity that the network layer cannot.

## Telemetry Events

The scooter's software (not the modem) controls telemetry: what to report, when
to report it, and how to encode it. The specifics vary by operator and firmware
generation, but the general patterns are consistent across the industry.

### Payload Structure

Each telemetry report carries a common envelope and a set of sensor readings. The
fields below use standard abbreviations from IoT and vehicle telematics:

| Field | Type | Description |
|-------|------|-------------|
| `device_id` | string | Scooter identity (VIN or operator-assigned fleet ID) |
| `ts` | ISO 8601 | Timestamp at the scooter's clock |
| `seq` | uint32 | Monotonic sequence number; gaps reveal lost reports |
| `trigger` | string | Why this report was sent (see [Triggers](#triggers)) |
| `gnss.lat`, `gnss.lng` | float64 | WGS 84 coordinates |
| `gnss.alt` | float64 | Altitude in meters above sea level |
| `gnss.hdop` | float64 | Horizontal dilution of precision (lower is better) |
| `gnss.sats` | uint8 | Satellites used in the fix |
| `gnss.fix` | string | Fix quality: `"none"`, `"2d"`, `"3d"` |
| `speed` | float64 | Ground speed in km/h (GNSS-derived) |
| `heading` | float64 | Course over ground in degrees, 0–360 |
| `battery.id` | string | BMS-reported serial number of the installed pack |
| `battery.soc` | float64 | State of charge, 0.0–1.0 |
| `battery.voltage` | float64 | Pack voltage in volts |
| `battery.temp` | int | Cell temperature in °C |
| `battery.cycles` | uint32 | Lifetime charge cycle count |
| `odometer` | float64 | Cumulative trip distance in km |
| `access` | object | Current point of attachment (see below) |
| `accel.motion` | bool | Accelerometer motion detection flag |
| `accel.tilt` | float64 | Tilt angle in degrees (0 = upright, 90 = on its side) |

The `access` field describes the scooter's current **point of attachment** to the
wider network: the intermediate node through which data reaches the operator's
backend. Its internal structure varies by connectivity type:

- **Cellular** (`"cellular"`): the *serving cell*, the cell tower sector the modem
  is camped on. Identified by PLMN, tracking area, and cell ID. Includes signal
  quality (RSRP).
- **Bluetooth** (`"bluetooth"`): the rider's phone acting as a relay. Identified
  by a peer identifier. Includes signal quality (RSSI).
- **Station** (`"station"`): a fixed infrastructure access point, typically WiFi
  at a docking station. Identified by the AP's address. Includes signal quality
  (RSSI).

The scooter's software includes the current access point in each telemetry report
when available. When the point of attachment changes (cell handover, new phone
pairs, docking station connects) or its properties update meaningfully, a
dedicated trigger fires (see [Triggers](#triggers)).

Not every field is present in every report. The scooter's software omits fields it
cannot populate (e.g., `gnss.*` when there is no satellite fix, `access` on
scooters with no active connectivity metadata).

The `device_id` is the application-layer identity of the scooter. It is
independent of the IMEI: the IMEI belongs to the modem at the network layer,
while the `device_id` belongs to the scooter's software at the application layer.
The platform correlates the two by observing which `device_id` payloads flow
through which IMEI's IP session.

### Triggers

The `trigger` field records why this report was emitted. Reports are either
periodic (fired on a timer) or edge-triggered (fired by a state change detected
by the scooter's software):

**Periodic:**

| Trigger | Cadence | Condition |
|---------|---------|-----------|
| `periodic` | 60–120 s | Scooter is idle (no motion detected) |
| `periodic` | 2–5 s | Scooter is in motion (ride in progress) |

The cadence is governed by the scooter's firmware. The same trigger value
covers both idle and ride modes; the platform infers the scooter's state from
the reporting frequency and the `accel.motion` flag.

**Edge-triggered:**

| Trigger | Fired when |
|---------|------------|
| `motion_start` | Accelerometer transitions from stationary to moving |
| `motion_stop` | Accelerometer transitions from moving to stationary |
| `impact` | Accelerometer spike exceeds impact threshold (fall or collision) |
| `battery_swap` | BMS reports a different battery serial after power cycle |
| `low_battery` | State of charge drops below configured threshold |
| `access_change` | Point of attachment changes (cell handover, new phone, dock connect) |
| `access_update` | Same point of attachment, signal quality or properties refreshed |
| `gnss_fix` | GNSS acquires a fix after a period with no fix |
| `power_on` | Scooter boots after battery insertion or deep sleep wake |

Edge-triggered reports are sent immediately, independently of the periodic timer.
A burst of events can occur in quick succession (e.g., `power_on` followed by
`battery_swap` followed by `gnss_fix` within seconds of a battery replacement).

### Reporting Cadence Over Time

```
          power_on  motion_start       periodic (ride)         motion_stop
              │         │          ┌──────────┬──────────┐         │
  ──idle──────┼─────────┼──ride────┼──ride────┼──ride────┼─────────┼──idle─────
  60s    60s  ▼    60s  ▼  3s  3s  ▼  3s  3s  ▼  3s  3s  ▼   3s   ▼  60s   60s
  ·      ·    ·    ·    ·  · · · · ·  · · · · ·  · · · · ·   · ·  ·  ·     ·
```

## A Scooter's Day

The following traces one cellular-connected scooter through a full day. This is
the richest observability scenario; phone-bridged scooters would produce a sparser
sequence (telemetry only during rides, no persistent modem identity, longer gaps).

### Early Morning

Scooter `WMX00042` sits idle near a docking station at Rothschild Boulevard. Its
modem (IMEI `353456789012345`) has been connected since last night with IP
`100.72.14.201`. Battery `BAT-1042` is at 12% after a long evening of rides.

The scooter's software sends idle heartbeats every 90 seconds:

```
{device_id: "WMX00042", ts: "06:12:04Z", seq: 81040, trigger: "periodic",
 gnss: {lat: 32.0636, lng: 34.7748, fix: "3d", hdop: 1.1, sats: 9},
 speed: 0, heading: 0,
 battery: {id: "BAT-1042", soc: 0.12, voltage: 36.1, temp: 18, cycles: 487},
 accel: {motion: false, tilt: 2.1}}
```

A field technician pulls up in a van, pops the battery latch, and swaps
`BAT-1042` for a fresh pack `BAT-2187`. The scooter powers down momentarily and
comes back up. Three edge-triggered events fire in rapid succession:

```
{device_id: "WMX00042", ts: "06:14:31Z", seq: 81041, trigger: "power_on",
 battery: {id: "BAT-2187", soc: 0.97, voltage: 42.0, temp: 22, cycles: 31},
 accel: {motion: false, tilt: 3.0}}

{device_id: "WMX00042", ts: "06:14:31Z", seq: 81042, trigger: "battery_swap",
 battery: {id: "BAT-2187", soc: 0.97, voltage: 42.0, temp: 22, cycles: 31}}

{device_id: "WMX00042", ts: "06:14:38Z", seq: 81043, trigger: "gnss_fix",
 gnss: {lat: 32.0636, lng: 34.7748, fix: "3d", hdop: 1.3, sats: 7},
 battery: {id: "BAT-2187", soc: 0.97, voltage: 42.0, temp: 22, cycles: 31},
 accel: {motion: false, tilt: 2.4}}
```

Note: the `power_on` event has no GNSS data because the receiver has not acquired
a fix yet. The `battery_swap` event fires once the scooter's software detects the
new battery serial. The `gnss_fix` event follows a few seconds later when
satellites are reacquired.

In the platform's model, `BAT-2187` is now associated with `WMX00042`, and
`BAT-1042` has no vehicle association until it appears in another scooter.

### Morning Commute

At 08:14, a rider unlocks the scooter through the Lime app. The scooter's
software detects motion:

```
{device_id: "WMX00042", ts: "08:14:02Z", seq: 81099, trigger: "motion_start",
 gnss: {lat: 32.0636, lng: 34.7748, fix: "3d", hdop: 0.9, sats: 11},
 speed: 0.3, heading: 355,
 battery: {id: "BAT-2187", soc: 0.94, voltage: 41.8, temp: 26, cycles: 31},
 odometer: 0.0,
 accel: {motion: true, tilt: 1.8}}
```

The reporting cadence jumps to every 3 seconds. As the rider heads north along
Dizengoff Street, periodic ride telemetry flows:

```
{.., ts: "08:14:05Z", seq: 81100, trigger: "periodic",
 gnss: {lat: 32.0638, lng: 34.7747, ..}, speed: 5.2, heading: 358,
 battery: {id: "BAT-2187", soc: 0.94, ..}, odometer: 0.01, ..}

{.., ts: "08:14:08Z", seq: 81101, trigger: "periodic",
 gnss: {lat: 32.0643, lng: 34.7746, ..}, speed: 12.7, heading: 2, ..}

{.., ts: "08:14:11Z", seq: 81102, trigger: "periodic",
 gnss: {lat: 32.0651, lng: 34.7745, ..}, speed: 17.8, heading: 1, ..}
```

The platform asserts Speed and Location on each beacon and detects the sustained
movement pattern, setting InRide to true.

Meanwhile, at the network layer, the scooter's modem resolves `api.lime.com`. A
DNS-based classifier asserts that IMEI `353456789012345` is operated by Lime with
0.92 confidence. An IMSI-range analyzer recognizes the IMSI prefix `42501...` as
a Lime-contracted SIM plan and asserts the same operator at 0.71 confidence. A
payload-structure classifier recognizes the telemetry encoding as Segway Max G30
firmware and asserts a model classification. Three classifiers, three independent
assertions, all keyed on the same IMEI.

### Mid-Ride: Cell Handover

The rider crosses into a different cell sector. The network hands over the
session. The scooter's software detects the change in point of attachment:

```
{device_id: "WMX00042", ts: "08:22:17Z", seq: 81244, trigger: "access_change",
 gnss: {lat: 32.0812, lng: 34.7801, fix: "3d", hdop: 0.8, sats: 12},
 speed: 16.4, heading: 12,
 battery: {id: "BAT-2187", soc: 0.91, ..},
 access: {type: "cellular", mcc: 425, mnc: 01, tac: 12401, cid: 52417, rsrp: -78},
 accel: {motion: true, ..}}
```

At the network layer, the IP changes from `100.72.14.201` to `100.72.31.88`.
The IMEI and IMSI remain the same. The platform asserts the new IP on the Modem
entity; the old IP is overwritten by the new assertion.

### Ride Ends

At 08:31, the rider parks near Rabin Square and locks the scooter:

```
{device_id: "WMX00042", ts: "08:31:44Z", seq: 81420, trigger: "motion_stop",
 gnss: {lat: 32.0873, lng: 34.7811, fix: "3d", hdop: 0.9, sats: 10},
 speed: 0, heading: 15,
 battery: {id: "BAT-2187", soc: 0.88, voltage: 41.2, temp: 30, cycles: 31},
 odometer: 3.7,
 accel: {motion: false, tilt: 2.0}}
```

The scooter's software drops back to idle cadence: one heartbeat every 90
seconds. The platform's ride detector sees the transition from high-frequency
motion telemetry to low-frequency idle heartbeats and sets InRide to false.

### Afternoon: Silence and Staleness

By 14:00, six hours have passed without a ride. The scooter is still sending
idle heartbeats (speed 0, same location), keeping its properties fresh:

```
{device_id: "WMX00042", ts: "14:00:12Z", seq: 81783, trigger: "periodic",
 gnss: {lat: 32.0873, lng: 34.7811, fix: "3d", ..},
 speed: 0,
 battery: {id: "BAT-2187", soc: 0.85, ..},
 accel: {motion: false, tilt: 2.0}}
```

Now imagine the scooter gets moved into a parking garage by an overzealous
building manager. The GNSS fix degrades, then disappears. The cellular signal
weakens. Eventually, the scooter's software cannot send at all:

```
{device_id: "WMX00042", ts: "14:15:42Z", seq: 81793, trigger: "periodic",
 gnss: {fix: "none"},
 battery: {id: "BAT-2187", soc: 0.84, ..},
 access: {type: "cellular", mcc: 425, mnc: 01, tac: 12401, cid: 52418, rsrp: -112},
 accel: {motion: false, tilt: 2.1}}

... then silence.
```

After the configured telemetry TTL expires (say, 15 minutes with no beacon), the
platform retracts Speed, Location, and InRide. The scooter still exists in the
model (VIN, IMEI association, battery serial are stable facts), but its ephemeral
properties are gone. The platform knows the scooter exists; it does not know where
it is or what it is doing.

Note the last report: the GNSS fix was `"none"` but the battery and accelerometer
data were still valid. The scooter's software sent what it could. The platform
asserted the battery level and retracted the location (no fix means no
coordinates). Partial data is normal.

If the IP session also times out on the network side, the platform retracts IP
from the Modem entity. The IMEI, EID, ICCID, and IMSI remain: they are facts
about the hardware and its profile, not about the current session.

### Evening: eSIM Reprovisioning

The operator decides to switch carrier for better evening coverage in the city
center. A remote eSIM management platform pushes a new profile to the eUICC. The
ICCID and IMSI change; the old data session is torn down. The platform sees the
new IMSI appear on the same IMEI and asserts the updated profile identifiers.
Because the old IP belonged to the previous profile's session, it is retracted.
The EID stays the same: the eUICC chip did not change, only the profile stored on
it.

This is not a telemetry event. It is a network-layer observation: the platform
sees the IMEI re-attach with a different IMSI. No application payload is involved.

### Decommission

Months later, `WMX00042` is retired from the fleet. The operator stops sending
commands; the scooter stops reporting. Its ephemeral properties are retracted by
timeout. Its stable properties (VIN, IMEI association) remain in the model
indefinitely as historical record, until explicitly retracted by an
administrative event.

## Last Reported Values

When the platform retracts a spot value (speed, location, IP) because its TTL
expired, the value is gone from the live model. But for some properties, the
stale value is still better than nothing.

Consider a fleet dispatcher looking for scooter `WMX00042`, which went silent 40
minutes ago. The live model shows no location (retracted after 15 minutes). But
the *last reported* position, Rabin Square at 14:15, is still the best guess for
where a technician should start looking. Similarly, the last reported battery
level (84%) tells the dispatcher whether the scooter is worth retrieving or
already depleted.

Not every property benefits from this treatment. The table below names properties
after the domain model, then describes what "last reported" means for each:

| Property | Last reported useful? | Notes |
|----------|---------------------|-------|
| **GNSS fix** | Yes | The scooter's GPS position (lat, lng, altitude, precision). This is the primary geolocation source. A 40-minute-old fix is still the best starting point for a field technician searching for a silent scooter. |
| **Serving cell** | Yes | The modem's serving cell (PLMN, tracking area, cell ID, signal strength). Provides a coarse geolocation estimate when GNSS is unavailable. But the last reported serving cell is not the only source of coarse position; see below. |
| **Battery** | Moderate | The full BMS reading (serial, SoC, voltage, temperature, cycles). Helps prioritize which scooters need attention, but a stale level may be outdated by a swap or charging event the platform missed. |
| **Speed** | No | Stale speed is misleading; you care about current or nothing. |
| **InRide** | No | Stale ride status is unreliable. |
| **IP** | No | Stale IP is likely reassigned to another device. |

Note that the high-value properties (GNSS fix, serving cell, battery) are wide:
they are structured objects with multiple fields, not scalars. "Last reported GNSS
fix" means the entire `gnss` struct from the most recent telemetry report that
had a valid fix, including its precision metadata (`hdop`, `sats`, `fix`).

### Geolocation as a Derived View

The properties above live in the domain model as-is (GNSS fix on the vehicle,
serving cell on the modem). But fleet operations often want a single answer:
"where is this scooter?" That answer, a *geolocation*, is derived from whichever
source is available, in rough priority order:

1. **GNSS fix** — highest precision (meters), from the scooter's GPS receiver.
2. **Serving cell** — coarse (hundreds of meters to kilometers), from the modem's
   last reported cell attachment or from the rider's phone during a BLE-bridged
   ride.
3. **Signaling triangulation** — moderate precision, derived from network
   signaling associated with the IMSI (e.g., timing advance, neighboring cell
   measurements).

Each of these sources has different precision, freshness, and availability. The
platform could maintain a composite "best available geolocation" that selects the
best source, or leave the fusion to downstream consumers querying the individual
properties. This is a design question for implementation.

### How to Provide Last Reported Values

Three approaches, each with different trade-offs:

**Separate entity.** A dedicated entity type (e.g., a "last reported position"
keyed on VIN) that is asserted alongside the spot value but is not subject to the
same TTL. It carries its own timestamp ("as of when?") and is only overwritten by
the next valid observation, never retracted by timeout. Clean separation, but
doubles the number of entities for each property that needs it.

**Additional fields on the same entity.** The Vehicle carries both `Location`
(current, retractable) and `LastReportedLocation` (persistent until next
observation). Simpler than a separate entity, but the two fields have different
retraction semantics on the same struct, which complicates the delta model.
Asserting a new location must update both fields; retracting by timeout must
clear only the current one.

**Analytics tier only.** The silver Parquet baseline already holds the latest
snapshot per entity. "Last reported location" is a query: find the most recent
snapshot where location was non-null. This requires no additional live entities
and no Go struct changes. The trade-off is latency: the answer is only as fresh
as the last silver rebuild, and consumers must query Parquet rather than subscribe
to a live NATS stream.

The right approach may vary by property. Geolocation, the highest-value case,
could justify a live entity. Battery level, with moderate value and higher risk of being
outdated, may be fine as an analytics query. The decision is deferred to
implementation.

## Competing Knowledge

Not everything the platform knows is certain. When two classifiers assert
different operators for the same modem (say, the DNS classifier sees `lime.com`
but the IMSI-range analyzer maps the prefix to Bird), both assertions coexist.
Each has its own source identifier, confidence score, and deployment version.
Downstream consumers decide how to reconcile: pick the highest confidence, require
agreement, or flag the conflict.

This is not a bug; it is a feature of the domain. The real world is ambiguous.
Scooters are reflashed and resold between operators. SIM plans are occasionally
shared across brands. Classifiers are imperfect and versioned: today's
`dns-fingerprint v2.3` might be replaced by `v3.0` next quarter with different
accuracy characteristics. The domain model holds all perspectives simultaneously
rather than forcing premature resolution.

## Variance

The scooter domain has many axes of variation, and the platform must handle them
all without requiring a uniform data source.

| Axis | Range | Effect on observability |
|------|-------|------------------------|
| **Operator** | Lime, Bird, Dot, others | Different firmware, payload formats, reporting intervals |
| **Hardware generation** | Gen 1 (BLE-only) through Gen 4 (cellular + BLE) | Gen 1 has no modem; no persistent network identity between rides |
| **Connectivity** | Cellular, phone-bridged, hybrid, WiFi-at-dock | Determines whether a stable Modem entity exists |
| **Carrier** | Multiple MNOs per region | Different IMSI prefixes, APN configurations, IP address pools |
| **Coverage** | Full signal to dead zones | Observation gaps; properties go stale at different rates |
| **Firmware version** | Operator-specific, updated OTA | Payload structure and trigger logic may differ between versions |

The delta model (assert/retract/ignore) absorbs this variance naturally: the
platform asserts what it can observe, ignores what it cannot, and retracts what
has gone stale. A richly-connected Gen 4 scooter and a BLE-only Gen 1 scooter
coexist in the same model; the Gen 1 simply has fewer asserted properties and
longer gaps between updates.

## Characters

The tangible objects that interact in this world:

| Character | Identity | Lifetime | What changes |
|-----------|----------|----------|--------------|
| **Modem** | IMEI (hardware), EID (eUICC) | Permanent — manufactured into the scooter | eSIM profile (ICCID, IMSI), IP session, firmware |
| **Vehicle** | VIN (frame) | Permanent — stamped at manufacture | Which modem is installed, speed, location, ride status |
| **Battery** | Serial number (BMS chip) | Permanent — printed on casing | Which vehicle it is installed in, charge level, cycle count |
| **Operator classification** | IMEI + classifier source | Per classifier assertion | Which operator, confidence, classifier version |
| **Model classification** | IMEI + classifier source | Per classifier assertion | Manufacturer, model designation, confidence |

Note that the Modem and Vehicle are observed through independent channels
(network layer vs application layer). For cellular-connected scooters, both
characters are present and correlated. For phone-bridged scooters, the Vehicle
may exist without a stable Modem anchor.
