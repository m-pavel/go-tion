package tion

// RestStatus structure
type RestStatus struct {
	Gate          string `json:"gate"`
	On            *bool  `json:"on"`
	Heater        *bool  `json:"heater"`
	Sound         bool   `json:"sound"`
	Out           int8   `json:"temp_out"`
	In            int8   `json:"temp_in"`
	Target        int8   `json:"temp_target"`
	Speed         *int8  `json:"speed"`
	FilterRemains int    `json:"filters"`
	Firmware      int    `json:"firmware"`

	Hours        int8 `json:"run_hours"`
	Minutes      int8 `json:"run_minutes"`
	ErrorCode    int8 `json:"error"`
	Productivity int8 `json:"productivity"`
	RunDays      int  `json:"run_days"`
}

// RestFromStatus to RestStatus
func RestFromStatus(s *Status) *RestStatus {
	speed := s.Speed
	if !s.Enabled {
		speed = 0
	}
	return &RestStatus{
		Gate:          s.GateStatus(),
		On:            &s.Enabled,
		Heater:        &s.HeaterEnabled,
		Out:           s.TempOut,
		In:            s.TempIn,
		Target:        s.TempTarget,
		Speed:         &speed,
		Sound:         s.SoundEnabled,
		FilterRemains: s.FiltersRemains,
		Firmware:      s.FirmwareVersion,
		Hours:         s.Hours,
		Minutes:       s.Minutes,
		RunDays:       s.RunDays,
		Productivity:  s.Productivity,
		ErrorCode:     s.ErrorCode,
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
	if rs.Heater != nil {
		s.HeaterEnabled = *rs.Heater
	}
	s.TempOut = rs.Out
	s.TempIn = rs.In
	s.TempTarget = rs.Target
	s.SoundEnabled = rs.Sound
	s.FiltersRemains = rs.FilterRemains
	s.FirmwareVersion = rs.Firmware
	s.ErrorCode = rs.ErrorCode
	s.RunDays = rs.RunDays
	s.Hours = rs.Hours
	s.Minutes = rs.Minutes
	s.Productivity = rs.Productivity
	return &s
}
