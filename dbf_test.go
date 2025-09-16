package dbfmini

import (
	"encoding/binary"
	"testing"
	"time"
)

func TestParseRecordTimeFieldAtEndOfDay(t *testing.T) {
	d := &DBF{
		Fields: []Field{
			{
				Name: "DT",
				Type: 'T',
				Size: 8,
			},
		},
	}

	record := make([]byte, 1+8)
	record[0] = ' '

	julian := julianDay(2023, time.August, 15)
	binary.LittleEndian.PutUint32(record[1:], uint32(julian))

	ms := uint32(((23*60+59)*60+59)*1000 + 999)
	binary.LittleEndian.PutUint32(record[5:], ms)

	rec, skip, err := d.parseRecord(record)
	if err != nil {
		t.Fatalf("parseRecord returned error: %v", err)
	}
	if skip {
		t.Fatalf("parseRecord skipped record unexpectedly")
	}

	raw, ok := rec["DT"]
	if !ok {
		t.Fatalf("record missing DT field")
	}

	dt, ok := raw.(time.Time)
	if !ok {
		t.Fatalf("DT field type = %T, want time.Time", raw)
	}

	expected := time.Date(2023, time.August, 15, 23, 59, 59, 0, time.UTC)
	if !dt.Equal(expected) {
		t.Fatalf("unexpected datetime: got %v, want %v", dt, expected)
	}
}

func julianDay(year int, month time.Month, day int) int {
	a := (14 - int(month)) / 12
	y := year + 4800 - a
	m := int(month) + 12*a - 3
	return day + (153*m+2)/5 + 365*y + y/4 - y/100 + y/400 - 32045
}
