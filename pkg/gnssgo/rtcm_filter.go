package gnssgo

// RTCMMessageType represents an RTCM message type
type RTCMMessageType int

// RTCM message types
const (
	// GPS L1 observation
	RTCM_MSG_1001 RTCMMessageType = 1001 // GPS L1-only RTK observables
	RTCM_MSG_1002 RTCMMessageType = 1002 // GPS Extended L1-only RTK observables
	RTCM_MSG_1003 RTCMMessageType = 1003 // GPS L1/L2 RTK observables
	RTCM_MSG_1004 RTCMMessageType = 1004 // GPS Extended L1/L2 RTK observables

	// GLONASS L1 observation
	RTCM_MSG_1009 RTCMMessageType = 1009 // GLONASS L1-only RTK observables
	RTCM_MSG_1010 RTCMMessageType = 1010 // GLONASS Extended L1-only RTK observables
	RTCM_MSG_1011 RTCMMessageType = 1011 // GLONASS L1/L2 RTK observables
	RTCM_MSG_1012 RTCMMessageType = 1012 // GLONASS Extended L1/L2 RTK observables

	// System parameters
	RTCM_MSG_1013 RTCMMessageType = 1013 // System parameters

	// Reference station information
	RTCM_MSG_1005 RTCMMessageType = 1005 // Stationary RTK reference station ARP
	RTCM_MSG_1006 RTCMMessageType = 1006 // Stationary RTK reference station ARP with antenna height

	// Antenna description
	RTCM_MSG_1007 RTCMMessageType = 1007 // Antenna descriptor
	RTCM_MSG_1008 RTCMMessageType = 1008 // Antenna descriptor and serial number

	// GPS ephemeris
	RTCM_MSG_1019 RTCMMessageType = 1019 // GPS ephemerides

	// GLONASS ephemeris
	RTCM_MSG_1020 RTCMMessageType = 1020 // GLONASS ephemerides

	// MSM messages
	RTCM_MSG_1071 RTCMMessageType = 1071 // GPS MSM1
	RTCM_MSG_1072 RTCMMessageType = 1072 // GPS MSM2
	RTCM_MSG_1073 RTCMMessageType = 1073 // GPS MSM3
	RTCM_MSG_1074 RTCMMessageType = 1074 // GPS MSM4
	RTCM_MSG_1075 RTCMMessageType = 1075 // GPS MSM5
	RTCM_MSG_1076 RTCMMessageType = 1076 // GPS MSM6
	RTCM_MSG_1077 RTCMMessageType = 1077 // GPS MSM7

	RTCM_MSG_1081 RTCMMessageType = 1081 // GLONASS MSM1
	RTCM_MSG_1082 RTCMMessageType = 1082 // GLONASS MSM2
	RTCM_MSG_1083 RTCMMessageType = 1083 // GLONASS MSM3
	RTCM_MSG_1084 RTCMMessageType = 1084 // GLONASS MSM4
	RTCM_MSG_1085 RTCMMessageType = 1085 // GLONASS MSM5
	RTCM_MSG_1086 RTCMMessageType = 1086 // GLONASS MSM6
	RTCM_MSG_1087 RTCMMessageType = 1087 // GLONASS MSM7

	RTCM_MSG_1091 RTCMMessageType = 1091 // Galileo MSM1
	RTCM_MSG_1092 RTCMMessageType = 1092 // Galileo MSM2
	RTCM_MSG_1093 RTCMMessageType = 1093 // Galileo MSM3
	RTCM_MSG_1094 RTCMMessageType = 1094 // Galileo MSM4
	RTCM_MSG_1095 RTCMMessageType = 1095 // Galileo MSM5
	RTCM_MSG_1096 RTCMMessageType = 1096 // Galileo MSM6
	RTCM_MSG_1097 RTCMMessageType = 1097 // Galileo MSM7

	RTCM_MSG_1101 RTCMMessageType = 1101 // SBAS MSM1
	RTCM_MSG_1102 RTCMMessageType = 1102 // SBAS MSM2
	RTCM_MSG_1103 RTCMMessageType = 1103 // SBAS MSM3
	RTCM_MSG_1104 RTCMMessageType = 1104 // SBAS MSM4
	RTCM_MSG_1105 RTCMMessageType = 1105 // SBAS MSM5
	RTCM_MSG_1106 RTCMMessageType = 1106 // SBAS MSM6
	RTCM_MSG_1107 RTCMMessageType = 1107 // SBAS MSM7

	RTCM_MSG_1111 RTCMMessageType = 1111 // QZSS MSM1
	RTCM_MSG_1112 RTCMMessageType = 1112 // QZSS MSM2
	RTCM_MSG_1113 RTCMMessageType = 1113 // QZSS MSM3
	RTCM_MSG_1114 RTCMMessageType = 1114 // QZSS MSM4
	RTCM_MSG_1115 RTCMMessageType = 1115 // QZSS MSM5
	RTCM_MSG_1116 RTCMMessageType = 1116 // QZSS MSM6
	RTCM_MSG_1117 RTCMMessageType = 1117 // QZSS MSM7

	RTCM_MSG_1121 RTCMMessageType = 1121 // BeiDou MSM1
	RTCM_MSG_1122 RTCMMessageType = 1122 // BeiDou MSM2
	RTCM_MSG_1123 RTCMMessageType = 1123 // BeiDou MSM3
	RTCM_MSG_1124 RTCMMessageType = 1124 // BeiDou MSM4
	RTCM_MSG_1125 RTCMMessageType = 1125 // BeiDou MSM5
	RTCM_MSG_1126 RTCMMessageType = 1126 // BeiDou MSM6
	RTCM_MSG_1127 RTCMMessageType = 1127 // BeiDou MSM7

	// Proprietary messages
	RTCM_MSG_4094 RTCMMessageType = 4094 // Trimble proprietary message
)

// RTCMFilter is a function that filters RTCM messages
type RTCMFilter func(msgType RTCMMessageType) bool

// DefaultRTCMFilter returns a default RTCM filter that filters out unwanted message types
func DefaultRTCMFilter() RTCMFilter {
	return func(msgType RTCMMessageType) bool {
		// Filter out unwanted message types
		switch msgType {
		case RTCM_MSG_4094: // Trimble proprietary message
			return false
		case RTCM_MSG_1013: // System parameters (not needed for RTK)
			return false
		default:
			return true
		}
	}
}

// CriticalRTCMFilter returns a filter that only allows critical message types for RTK
func CriticalRTCMFilter() RTCMFilter {
	return func(msgType RTCMMessageType) bool {
		// Only allow critical message types
		switch msgType {
		// Reference station information (needed for RTK)
		case RTCM_MSG_1005, RTCM_MSG_1006:
			return true
		
		// GPS observations
		case RTCM_MSG_1001, RTCM_MSG_1002, RTCM_MSG_1003, RTCM_MSG_1004:
			return true
			
		// GLONASS observations
		case RTCM_MSG_1009, RTCM_MSG_1010, RTCM_MSG_1011, RTCM_MSG_1012:
			return true
			
		// GPS ephemeris
		case RTCM_MSG_1019:
			return true
			
		// GLONASS ephemeris
		case RTCM_MSG_1020:
			return true
			
		// MSM messages (high priority)
		case RTCM_MSG_1074, RTCM_MSG_1084, RTCM_MSG_1094, RTCM_MSG_1124:
			return true
			
		default:
			return false
		}
	}
}

// FilterRTCMMessages filters RTCM messages based on the provided filter
func FilterRTCMMessages(msgTypes []RTCMMessageType, filter RTCMFilter) []RTCMMessageType {
	var filtered []RTCMMessageType
	for _, msgType := range msgTypes {
		if filter(msgType) {
			filtered = append(filtered, msgType)
		}
	}
	return filtered
}
