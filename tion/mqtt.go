package tion

// RestStatus structure
type RestStatus struct {
	Gate          string `json:"gate"`
	On            *bool  `json:"on"`
	Heater        bool   `json:"heater"`
	Sound         bool   `json:"sound"`
	Out           int8   `json:"temp_out"`
	In            int8   `json:"temp_in"`
	Target        int8   `json:"temp_target"`
	Speed         *int8  `json:"speed"`
	FilterRemains int    `json:"filters"`
	Firmware      int    `json:"firmware"`
}

// RestFromStatus to RestStatus
func RestFromStatus(s *Status) *RestStatus {
	return &RestStatus{
		Gate:          s.GateStatus(),
		On:            &s.Enabled,
		Heater:        s.HeaterEnabled,
		Out:           s.TempOut,
		In:            s.TempIn,
		Target:        s.TempTarget,
		Speed:         &s.Speed,
		Sound:         s.SoundEnabled,
		FilterRemains: s.FiltersRemains,
		Firmware:      s.FirmwareVersion,
	}
}

// StatusFromRest to Status
func StatusFromRest(rs *RestStatus) *Status {
	s := Status{}
	s.SetGateStatus(rs.Gate)
	if rs.Speed != nil {
		s.Speed = *rs.Speed
	}
	if rs.On != nil {
		s.Enabled = *rs.On
	}
	s.HeaterEnabled = rs.Heater
	s.TempOut = rs.Out
	s.TempIn = rs.In
	s.TempTarget = rs.Target
	s.SoundEnabled = rs.Sound
	s.FiltersRemains = rs.FilterRemains
	s.FirmwareVersion = rs.Firmware
	return &s
}
