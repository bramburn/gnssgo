<img src="https://r2cdn.perplexity.ai/pplx-full-logo-primary-dark%402x.png" class="logo" width="120"/>

# give me a detailed guide of the nmea codes used for rtk, ntrip, rover, caster, server

RTK (Real-Time Kinematic) systems rely on NMEA 0183 messages to communicate positioning data between rovers, casters, and servers in NTRIP (Networked Transport of RTCM via Internet Protocol) networks. Below is a detailed breakdown of key NMEA codes and their roles:

---

## Core NMEA Messages for RTK/NTRIP

**\$GPGGA (Global Positioning System Fix Data)**

- Provides essential RTK positioning data:
    - UTC time (hhmmss.ss format)
    - Latitude/longitude (ddmm.mmmmm format, ±1.85cm precision per 5th decimal digit)[^1_1]
    - Fix quality (1=RTK Float, 4=RTK Fixed)[^1_5]
    - HDOP and altitude
- **Critical for NTRIP**: Clients send GGA messages to casters to request localized correction streams[^1_2][^1_4]. Example format:

```bash
$GPGGA,181908.00,3404.7041778,N,07044.3966270,W,4,13,1.00,495.144,M,29.200,M,0.10,0000*40
```


**\$GPGSA (GNSS DOP and Active Satellites)**

- Lists satellites used in solution and accuracy metrics:
    - PDOP/HDOP/VDOP
    - 3D fix type (2=2D, 3=3D)[^1_6]

**\$GPGSV (GNSS Satellites in View)**

- Reports visible satellites:
    - Satellite ID, elevation, azimuth, and SNR
    - Essential for diagnosing signal quality[^1_3][^1_7]

**\$GPRMC (Recommended Minimum Specific GNSS Data)**

- Compact position/time/velocity summary:
    - Ground speed (knots)
    - Magnetic variation[^1_6]

---

## NTRIP-Specific NMEA Implementations

### Client-to-Caster Communication

- **NMEA-GGA Requirement**: Many casters require clients to send GGA messages with approximate positions to select appropriate correction streams (`<nmea>=1` in NTRIP parameters)[^1_2][^1_4].
- **\$GNSNC (NTRIP Client Status)**:
    - Transmits connection status between rover and caster
    - Includes mountpoint, IP, and data latency metrics[^1_7]


### Caster/Server Parameters

NTRIP servers use metadata fields linked to NMEA data:


| Parameter | Purpose | Example Values |
| :-- | :-- | :-- |
| `<nmea>` | Enforce GGA transmission requirement | 0 (disabled), 1 (enabled)[^1_2] |
| `<latitude>` | Approx. position (from GGA) | 40.12 (decimal deg) |
| `<longitude>` | Approx. position (from GGA) | -104.77 (decimal deg) |


---

## RTK Workflow Integration

1. **Rover Initialization**:
    - Outputs GGA, GSA, GSV messages to assess satellite geometry[^1_3][^1_6]
2. **NTRIP Connection**:
    - Rover sends GGA via HTTP to caster (e.g., `GET /MOUNTPOINT?GGA=... HTTP/1.1`)[^1_4]
3. **Correction Streaming**:
    - Caster replies with RTCM3.x messages matched to rover's GGA position[^1_2]
4. **RTK Fix Validation**:
    - GST messages report pseudorange errors (<1cm in RTK Fixed)[^1_7]

---

## Protocol Limitations and Alternatives

While NMEA supports RTK workflows, its 1.85cm precision per 5th decimal digit[^1_1] contrasts with RTK’s 1cm accuracy. For millimeter precision, most systems pair NMEA with:

- **UBX Protocol**: Outputs latitude/longitude in dd.dddddddd format (1.11mm precision)[^1_1]
- **RTCM3.x Messages**: Carrier-phase corrections (1004-1012, 1033, 1074-1084)[^1_7]

Use `NMEA-GGA` + `RTCM3` as a standard combination for RTK via NTRIP, reserving UBX for survey-grade applications requiring sub-centimeter precision[^1_1][^1_7].

<div style="text-align: center">⁂</div>

[^1_1]: https://learn.sparkfun.com/tutorials/gps-rtk-hookup-guide/nmea-and-rtk

[^1_2]: https://gssc.esa.int/wp-content/uploads/2018/07/NtripDocumentation.pdf

[^1_3]: https://www.comnavtech.com/about/blogs/584.html

[^1_4]: http://www.wsrn3.org/content/reference/cheatsheet_leica1200-to-network-via-ntrip.pdf

[^1_5]: https://receiverhelp.trimble.com/alloy-gnss/en-us/NMEA-0183messages_GGA.html

[^1_6]: https://www.gpsworld.com/what-exactly-is-gps-nmea-data/

[^1_7]: https://docs.sparkfun.com/SparkFun_RTK_Everywhere_Firmware/menu_messages/

[^1_8]: https://portal.u-blox.com/s/question/0D52p0000E0o9asCQA/nmea-messages-to-monitor-rtk-position

[^1_9]: https://github.com/dotMorten/NmeaParser/blob/main/docs/concepts/ntrip.md/

[^1_10]: https://www.ntp.org/documentation/drivers/driver20/


---

## Main Differences Between NMEA and UBX Protocols

**NMEA** and **UBX** are two communication protocols used by GNSS (GPS) receivers, especially those from u-blox. They serve similar purposes—transmitting positioning and related data—but differ significantly in format, capability, and use cases.

---

### **Format and Structure**

- **NMEA**
    - Text-based, ASCII protocol.
    - Messages are human-readable sentences, starting with a `$` and ending with a checksum (e.g., `$GPGGA,...*hh`).
    - Standardized and widely supported across the GNSS industry.
    - Each message is typically limited in the amount of data it can carry and is verbose, resulting in more bytes transmitted for the same information[^2_5][^2_6].
- **UBX**
    - Binary protocol, proprietary to u-blox.
    - Messages are compact, efficient, and not human-readable.
    - Lower bandwidth usage due to binary encoding.
    - Allows for more comprehensive and detailed data transmission per message[^2_5][^2_6].

---

### **Functionality and Features**

- **NMEA**
    - Primarily designed for marine navigation, but widely adopted for GNSS.
    - Provides basic positioning, satellite, and status information (e.g., position, time, fix type).
    - Standard NMEA messages generally include only 2D velocity data[^2_5].
    - Offers limited configuration and control over the receiver.
    - Satellite numbering and GNSS system support are limited and sometimes require extensions for multi-constellation use[^2_2].
- **UBX**
    - Offers access to advanced receiver features and settings.
    - Provides richer data, including 3D velocity (useful for inertial navigation and attitude correction)[^2_5].
    - Supports detailed configuration, including enabling/disabling GNSS systems, changing update rates, and querying receiver status.
    - Can deliver higher precision and more specialized information (e.g., raw measurements, detailed satellite data).
    - Allows firmware and operational updates, integrity checks, and advanced features like geofencing and anti-spoofing[^2_2].

---

### **Performance and Efficiency**

- **NMEA**
    - ASCII format is easier for humans to read but less efficient for computers to parse and transmit.
    - Parsing NMEA sentences requires more CPU cycles due to string manipulation[^2_6].
    - Larger message size for the same data compared to binary protocols[^2_5][^2_6].
- **UBX**
    - Binary format is more efficient for both transmission and parsing.
    - Lower CPU and bandwidth usage, especially at higher update rates or with large data sets[^2_5][^2_6].
    - Preferred for embedded applications and high-performance systems.

---

### **Compatibility and Use Cases**

| Aspect | NMEA | UBX |
| :-- | :-- | :-- |
| Readability | Human-readable | Machine-readable (binary) |
| Industry Support | Very broad (standardized) | u-blox devices only |
| Configuration | Limited | Extensive |
| Data Richness | Basic (position, time, satellites) | Advanced (raw data, 3D velocity, config) |
| Efficiency | Less efficient (more bytes, CPU load) | Highly efficient |
| Application | General GNSS, marine, legacy systems | Advanced, high-precision, embedded |


---

### **Summary Table**

| Feature | NMEA Protocol | UBX Protocol |
| :-- | :-- | :-- |
| Format | ASCII (text) | Binary |
| Standardization | Industry standard | Proprietary (u-blox only) |
| Data Content | Basic GNSS info | Detailed, configurable GNSS info |
| Efficiency | Less efficient (larger, slower) | More efficient (smaller, faster) |
| Configurability | Minimal | Extensive |
| Velocity Data | 2D only | 3D (full vector) |
| Use Case | General, compatibility | High-performance, u-blox-centric |


---

**In summary:**
NMEA is a widely compatible, human-readable protocol best for simple, general-purpose GNSS applications, while UBX is a compact, binary, and highly configurable protocol optimized for advanced features and efficiency on u-blox devices[^2_2][^2_5][^2_6].

<div style="text-align: center">⁂</div>

[^2_1]: https://portal.u-blox.com/s/question/0D52p0000CDJrwICQT/ubx-vs-nmea

[^2_2]: https://junipersys.com/data/support/u-blox7-to-u-blox8-m8-fw301-swmigrationguide_an_(ubx-15031124).pdf

[^2_3]: https://forum.arduino.cc/t/ubx-or-nmea-which-one-to-choose-for-ground-station/1024816

[^2_4]: https://portal.u-blox.com/s/question/0D52p0000AXd6LRCQZ/m8-nmea-reports-different-position-than-ubx

[^2_5]: http://www.multiwii.com/forum/viewtopic.php?t=5994

[^2_6]: https://lists.gnu.org/archive/html/paparazzi-devel/2012-03/msg00004.html

[^2_7]: https://paparazzi-devel.nongnu.narkive.com/wncwWCy0/ubx-vs-nmea-protocol-inquiry

[^2_8]: https://portal.u-blox.com/s/question/0D52p00008HKCphCAH/regarding-m8n-protocol-change-from-nmea-to-ubx

[^2_9]: https://www.embien.com/automotive-insights/exploring-the-nmea-2000-protocol

[^2_10]: https://gpsd.io/ubxtool-examples.html


---

## Specific NMEA Messages Used for RTK Applications

RTK (Real-Time Kinematic) GNSS systems rely on a set of NMEA 0183 messages to communicate precise positioning, satellite status, and system integrity information. The following NMEA messages are most commonly used in RTK workflows:

---

**Essential NMEA Messages for RTK:**

- **\$GGA (Global Positioning System Fix Data):**
    - Provides the main position, fix quality (including RTK fix status), number of satellites used, horizontal dilution of precision (HDOP), altitude, and geoid separation.
    - *Critical for RTK and NTRIP workflows; often required by NTRIP casters to determine the rover’s approximate location and deliver the correct differential corrections.*[^3_1][^3_3][^3_5][^3_9]
- **\$GSA (GNSS DOP and Active Satellites):**
    - Indicates which satellites are used in the solution and provides dilution of precision (PDOP, HDOP, VDOP) values.
    - Useful for assessing the quality and reliability of the RTK solution.[^3_3][^3_5]
- **\$GSV (GNSS Satellites in View):**
    - Lists all satellites currently visible, with their PRN numbers, elevation, azimuth, and signal-to-noise ratio (SNR).
    - Helps diagnose signal environment and satellite availability.[^3_1][^3_3][^3_5]
- **\$RMC (Recommended Minimum Specific GNSS Data):**
    - Contains position, time, date, speed, and course over ground.
    - Useful for logging and navigation applications alongside RTK.[^3_1][^3_3][^3_5]
- **\$GST (GNSS Pseudorange Error Statistics):**
    - Provides estimated position error statistics (standard deviations for latitude, longitude, and altitude).
    - Important for monitoring RTK solution accuracy.[^3_3][^3_5]
- **\$VTG (Course Over Ground and Ground Speed):**
    - Reports the actual track made good and speed over ground.
    - Useful for dynamic RTK applications (e.g., vehicle guidance).[^3_3][^3_5]
- **\$ZDA (Time and Date):**
    - Outputs UTC day, month, year, and local time zone offset.
    - Used for precise time-stamping of RTK data.[^3_1][^3_3][^3_5]
- **\$GLL (Geographic Position – Latitude/Longitude):**
    - Gives latitude, longitude, and time.
    - Sometimes used for additional position reporting.[^3_1][^3_5]

---

**Proprietary and Extended NMEA Messages (RTK-Specific):**

- **\$SNC (NTRIP Client Status):**
    - Used in some systems to communicate NTRIP connection status and related metrics.[^3_5]
- **Custom Messages (e.g., EBP, ETC):**
    - Some receivers output custom NMEA-like messages for advanced RTK features such as base position or tilt compensation.[^3_3]

---

**Talker ID Note:**
The two-letter prefix (e.g., GP, GL, GA, GB, GN) in the message header indicates the satellite system (GPS, GLONASS, Galileo, BeiDou, or multi-GNSS), which is important in multi-constellation RTK solutions.[^3_1][^3_5]

---

**Summary Table**


| NMEA Message | Purpose in RTK Applications |
| :-- | :-- |
| GGA | Main position/fix/RTK status |
| GSA | Satellites used, DOP values |
| GSV | Satellites in view, SNR |
| RMC | Position, time, velocity |
| GST | Position error estimates |
| VTG | Course and speed over ground |
| ZDA | UTC date and time |
| GLL | Latitude, longitude, and time |
| SNC/Custom | NTRIP/RTK-specific status or advanced features |


---

**In practice, the most critical NMEA message for RTK is \$GGA, but a complete RTK solution typically outputs GGA, GSA, GSV, RMC, GST, VTG, and ZDA for comprehensive status, diagnostics, and integration with mapping or control systems**[^3_1][^3_3][^3_5][^3_9].

<div style="text-align: center">⁂</div>

[^3_1]: https://www.singularxyz.com/3485.html

[^3_2]: https://www.gpsworld.com/what-exactly-is-gps-nmea-data/

[^3_3]: https://blog.emlid.com/what-is-nmea-and-how-to-feed-data-from-reach-to-a-third-party-device/

[^3_4]: https://drfasching.com/wp-content/uploads/2021/02/NV08C_RTK_NMEA_Protocol_Specification_V14_ENG.pdf

[^3_5]: https://docs.sparkfun.com/SparkFun_RTK_Everywhere_Firmware/menu_messages/

[^3_6]: https://customersupport.septentrio.com/s/article/NMEA-sentences

[^3_7]: https://gpsd.gitlab.io/gpsd/NMEA.html

[^3_8]: https://www.raveon.com/ApplicationNotes/AN246(NMEA_GNSS).pdf

[^3_9]: https://portal.u-blox.com/s/question/0D52p0000E0o9asCQA/nmea-messages-to-monitor-rtk-position

[^3_10]: https://developer.sbg-systems.com/sbgECom/5.1/nmea_msg.html

