# RTK2go Test Client Todo List

## Current Status
The RTK2go test client has been modified to use actual GNSS data from the connected receiver on COM3 instead of hardcoded London coordinates. The client now correctly parses GGA sentences from the NMEA data stream and displays the real-time position with proper RTK status (NONE, SINGLE, DGPS, FLOAT, FIX).

## Remaining Tasks

### High Priority
1. **Fix Position Reset Issue**: Sometimes the position is reset to (0,0) when a complete GGA sentence is not found in a single read. Implement a buffer to accumulate NMEA data across multiple reads to ensure we don't miss any GGA sentences.

2. **Improve Error Handling**: Add better error handling for cases where the GNSS receiver is not connected or not sending valid data.

3. **Maintain Last Known Position**: When no valid position is available, maintain the last known position instead of resetting to (0,0) or falling back to simulated data.

### Medium Priority
1. **Add Position Filtering**: Implement a simple filter to smooth out position jumps and provide a more stable display.

2. **Improve RTK Status Display**: Enhance the status display to show more information about the RTK solution, such as the number of satellites used in the solution, the age of differential corrections, etc.

3. **Add Position Logging**: Add the ability to log position data to a file for later analysis.

### Low Priority
1. **Add Map Display**: Integrate with a mapping library to display the position on a map.

2. **Add Configuration Options**: Add more configuration options for the GNSS receiver, such as the ability to set the update rate, enable/disable specific GNSS constellations, etc.

3. **Add Support for Other NMEA Sentences**: Add support for parsing other NMEA sentences, such as GSA (satellite status), GSV (satellites in view), etc.

## Implementation Notes
- The current implementation uses the TOP708Device directly to read NMEA data from the GNSS receiver.
- The GGA sentence parsing is working correctly when complete sentences are received.
- The RTK status is determined based on the fix quality field in the GGA sentence.
- The position is calculated by converting the NMEA format (DDMM.MMMM) to decimal degrees.
