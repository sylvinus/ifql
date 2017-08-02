package query

import "time"

// Time represents either a relavite or absolute time.
// If Time.Absolute is time.Zero then the time can is relative
type Time struct {
	Relative time.Duration
	Absolute time.Time
}

// Time returns the time specified relative to now.
func (t Time) Time(now time.Time) time.Time {
	if t.Absolute.IsZero() {
		return now.Add(t.Relative)
	}
	return t.Absolute
}

func (t *Time) UnmarshalText(data []byte) error {
	str := string(data)
	if str == "now" {
		t.Relative = 0
		t.Absolute = time.Time{}
		return nil
	}
	d, err := time.ParseDuration(str)
	if err == nil {
		t.Relative = d
		t.Absolute = time.Time{}
		return nil
	}
	t.Relative = 0
	t.Absolute, err = time.Parse(time.RFC3339Nano, str)
	if err != nil {
		return err
	}
	t.Absolute = t.Absolute.UTC()
	return nil
}

func (t Time) MarshalText() ([]byte, error) {
	if t.Absolute.IsZero() {
		if t.Relative == 0 {
			return []byte("now"), nil
		}
		return []byte(t.Relative.String()), nil
	}
	return []byte(t.Absolute.Format(time.RFC3339Nano)), nil
}

// Duration is a marshalable duration type.
//TODO make this the real duration parsing not just time.ParseDuration
type Duration time.Duration

func (d *Duration) UnmarshalText(data []byte) error {
	dur, err := time.ParseDuration(string(data))
	if err != nil {
		return err
	}
	*d = Duration(dur)
	return nil
}

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}
