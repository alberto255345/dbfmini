package dbfmini

import (
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var allFieldsDBF = []byte{
	0x03, 0x7c, 0x01, 0x01, 0x03, 0x00, 0x00, 0x00, 0x61, 0x01, 0x49, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x4e, 0x41, 0x4d, 0x45,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x43, 0x00, 0x00, 0x00, 0x00,
	0x0a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x41, 0x47, 0x45, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x4e, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x42, 0x41, 0x4c, 0x41, 0x4e, 0x43, 0x45, 0x00, 0x00, 0x00, 0x00, 0x46,
	0x00, 0x00, 0x00, 0x00, 0x0a, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x43, 0x55, 0x52, 0x52,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x59, 0x00, 0x00, 0x00, 0x00,
	0x08, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x41, 0x43, 0x54, 0x49, 0x56, 0x45, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x4c, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x42, 0x49, 0x52, 0x54, 0x48, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x44,
	0x00, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x43, 0x4f, 0x55, 0x4e,
	0x54, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x49, 0x00, 0x00, 0x00, 0x00,
	0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x52, 0x41, 0x54, 0x49, 0x4f, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x42, 0x00, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x53, 0x54, 0x41, 0x4d, 0x50, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x54,
	0x00, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x4d, 0x45, 0x4d, 0x4f,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x4d, 0x00, 0x00, 0x00, 0x00,
	0x0a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x0d, 0x20, 0x4a, 0x6f, 0x73, 0xe9, 0x20, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x34, 0x32, 0x20, 0x20, 0x20,
	0x20, 0x31, 0x32, 0x33, 0x2e, 0x34, 0x35, 0x08, 0xe2, 0x01, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x54, 0x31, 0x39, 0x39, 0x30, 0x30, 0x31, 0x30, 0x31,
	0x40, 0xe2, 0x01, 0x00, 0x6e, 0x86, 0x1b, 0xf0, 0xf9, 0x21, 0x09, 0x40,
	0x0c, 0x8a, 0x25, 0x00, 0x18, 0x58, 0x26, 0x05, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x88, 0xf5, 0x1c, 0xfa,
	0xff, 0xff, 0xff, 0xff, 0x46, 0x32, 0x30, 0x32, 0x31, 0x31, 0x32, 0x33,
	0x31, 0xce, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40, 0x45,
	0xc0, 0x27, 0x85, 0x25, 0x00, 0xd8, 0x25, 0xd3, 0x01, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2a, 0x44, 0x65, 0x6c, 0x65,
	0x74, 0x65, 0x64, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x39, 0x39, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x31, 0x2e, 0x32, 0x33, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x3f, 0x32, 0x30, 0x30, 0x30, 0x30, 0x31,
	0x30, 0x31, 0x07, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x04, 0x40, 0xe1, 0x84, 0x25, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x1a,
}

func writeAllFieldsFixture(t testing.TB) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "all_fields.dbf")
	if err := os.WriteFile(path, allFieldsDBF, 0o600); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}
	return path
}

func TestOpenReadsHeaderAndFields(t *testing.T) {
	path := writeAllFieldsFixture(t)

	db, err := Open(path, nil)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}

	if db.RecordCount != 3 {
		t.Fatalf("RecordCount = %d, want 3", db.RecordCount)
	}

	wantDate := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	if !db.DateOfLastUpd.Equal(wantDate) {
		t.Fatalf("DateOfLastUpd = %v, want %v", db.DateOfLastUpd, wantDate)
	}

	expected := []Field{
		{Name: "NAME", Type: 'C', Size: 10, DecimalPlaces: 0},
		{Name: "AGE", Type: 'N', Size: 5, DecimalPlaces: 0},
		{Name: "BALANCE", Type: 'F', Size: 10, DecimalPlaces: 2},
		{Name: "CURR", Type: 'Y', Size: 8, DecimalPlaces: 4},
		{Name: "ACTIVE", Type: 'L', Size: 1, DecimalPlaces: 0},
		{Name: "BIRTH", Type: 'D', Size: 8, DecimalPlaces: 0},
		{Name: "COUNT", Type: 'I', Size: 4, DecimalPlaces: 0},
		{Name: "RATIO", Type: 'B', Size: 8, DecimalPlaces: 0},
		{Name: "STAMP", Type: 'T', Size: 8, DecimalPlaces: 0},
		{Name: "MEMO", Type: 'M', Size: 10, DecimalPlaces: 0},
	}

	if len(db.Fields) != len(expected) {
		t.Fatalf("len(Fields) = %d, want %d", len(db.Fields), len(expected))
	}

	for i, want := range expected {
		got := db.Fields[i]
		if got.Name != want.Name || got.Type != want.Type || got.Size != want.Size || got.DecimalPlaces != want.DecimalPlaces {
			t.Fatalf("field[%d] = %+v, want %+v", i, got, want)
		}
	}
}

func TestReadRecordsParsesAllFieldTypes(t *testing.T) {
	path := writeAllFieldsFixture(t)

	db, err := Open(path, &OpenOptions{ReadMode: ReadLoose})
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}

	records, err := db.ReadRecords(0)
	if err != nil {
		t.Fatalf("ReadRecords returned error: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("len(records) = %d, want 2", len(records))
	}

	rec1 := records[0]
	if name, ok := rec1["NAME"].(string); !ok || name != "José" {
		t.Fatalf("rec1 NAME = %#v, want 'José'", rec1["NAME"])
	}
	if age, ok := rec1["AGE"].(float64); !ok || age != 42 {
		t.Fatalf("rec1 AGE = %#v, want 42", rec1["AGE"])
	}
	if bal, ok := rec1["BALANCE"].(float64); !ok || math.Abs(bal-123.45) > 1e-6 {
		t.Fatalf("rec1 BALANCE = %#v, want approx 123.45", rec1["BALANCE"])
	}
	if curr, ok := rec1["CURR"].(float64); !ok || math.Abs(curr-12.34) > 1e-6 {
		t.Fatalf("rec1 CURR = %#v, want approx 12.34", rec1["CURR"])
	}
	if active, ok := rec1["ACTIVE"].(bool); !ok || !active {
		t.Fatalf("rec1 ACTIVE = %#v, want true", rec1["ACTIVE"])
	}
	if birth, ok := rec1["BIRTH"].(time.Time); !ok || !birth.Equal(time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("rec1 BIRTH = %#v, want 1990-01-01", rec1["BIRTH"])
	}
	if count, ok := rec1["COUNT"].(int32); !ok || count != 123456 {
		t.Fatalf("rec1 COUNT = %#v, want 123456", rec1["COUNT"])
	}
	if ratio, ok := rec1["RATIO"].(float64); !ok || math.Abs(ratio-3.14159) > 1e-9 {
		t.Fatalf("rec1 RATIO = %#v, want approx 3.14159", rec1["RATIO"])
	}
	wantStamp1 := time.Date(2023, time.August, 15, 23, 59, 59, 0, time.UTC)
	if stamp, ok := rec1["STAMP"].(time.Time); !ok || !stamp.Equal(wantStamp1) {
		t.Fatalf("rec1 STAMP = %#v, want %v", rec1["STAMP"], wantStamp1)
	}
	if memo, ok := rec1["MEMO"]; !ok || memo != nil {
		t.Fatalf("rec1 MEMO = %#v, want nil", rec1["MEMO"])
	}

	rec2 := records[1]
	if name, ok := rec2["NAME"].(string); !ok || name != "" {
		t.Fatalf("rec2 NAME = %#v, want empty string", rec2["NAME"])
	}
	if age, ok := rec2["AGE"]; !ok || age != nil {
		t.Fatalf("rec2 AGE = %#v, want nil", rec2["AGE"])
	}
	if bal, ok := rec2["BALANCE"]; !ok || bal != nil {
		t.Fatalf("rec2 BALANCE = %#v, want nil", rec2["BALANCE"])
	}
	if curr, ok := rec2["CURR"].(float64); !ok || math.Abs(curr+9876.5432) > 1e-6 {
		t.Fatalf("rec2 CURR = %#v, want approx -9876.5432", rec2["CURR"])
	}
	if active, ok := rec2["ACTIVE"].(bool); !ok || active {
		t.Fatalf("rec2 ACTIVE = %#v, want false", rec2["ACTIVE"])
	}
	wantBirth2 := time.Date(2021, time.December, 31, 0, 0, 0, 0, time.UTC)
	if birth, ok := rec2["BIRTH"].(time.Time); !ok || !birth.Equal(wantBirth2) {
		t.Fatalf("rec2 BIRTH = %#v, want %v", rec2["BIRTH"], wantBirth2)
	}
	if count, ok := rec2["COUNT"].(int32); !ok || count != -50 {
		t.Fatalf("rec2 COUNT = %#v, want -50", rec2["COUNT"])
	}
	if ratio, ok := rec2["RATIO"].(float64); !ok || math.Abs(ratio+42.5) > 1e-9 {
		t.Fatalf("rec2 RATIO = %#v, want approx -42.5", rec2["RATIO"])
	}
	wantStamp2 := time.Date(2020, time.March, 10, 8, 30, 15, 0, time.UTC)
	if stamp, ok := rec2["STAMP"].(time.Time); !ok || !stamp.Equal(wantStamp2) {
		t.Fatalf("rec2 STAMP = %#v, want %v", rec2["STAMP"], wantStamp2)
	}
	if memo, ok := rec2["MEMO"]; !ok || memo != nil {
		t.Fatalf("rec2 MEMO = %#v, want nil", rec2["MEMO"])
	}

	// Reading again after EOF should return empty slice.
	records, err = db.ReadRecords(0)
	if err != nil {
		t.Fatalf("ReadRecords second call returned error: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("second ReadRecords len = %d, want 0", len(records))
	}
}

func TestReadRecordsIncludeDeleted(t *testing.T) {
	path := writeAllFieldsFixture(t)

	db, err := Open(path, &OpenOptions{ReadMode: ReadLoose, IncludeDeleted: true})
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}

	records, err := db.ReadRecords(0)
	if err != nil {
		t.Fatalf("ReadRecords returned error: %v", err)
	}

	if len(records) != 3 {
		t.Fatalf("len(records) = %d, want 3", len(records))
	}

	del := records[2]
	if name, ok := del["NAME"].(string); !ok || name != "Deleted" {
		t.Fatalf("deleted NAME = %#v, want 'Deleted'", del["NAME"])
	}
	if flag, ok := del["_deleted"].(bool); !ok || !flag {
		t.Fatalf("_deleted flag = %#v, want true", del["_deleted"])
	}
	if age, ok := del["AGE"].(float64); !ok || age != 99 {
		t.Fatalf("deleted AGE = %#v, want 99", del["AGE"])
	}
	if bal, ok := del["BALANCE"].(float64); !ok || math.Abs(bal-1.23) > 1e-6 {
		t.Fatalf("deleted BALANCE = %#v, want approx 1.23", del["BALANCE"])
	}
	if curr, ok := del["CURR"].(float64); !ok || math.Abs(curr-0) > 1e-9 {
		t.Fatalf("deleted CURR = %#v, want 0", del["CURR"])
	}
	if active, ok := del["ACTIVE"]; !ok || active != nil {
		t.Fatalf("deleted ACTIVE = %#v, want nil", del["ACTIVE"])
	}
	wantBirth := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	if birth, ok := del["BIRTH"].(time.Time); !ok || !birth.Equal(wantBirth) {
		t.Fatalf("deleted BIRTH = %#v, want %v", del["BIRTH"], wantBirth)
	}
	if count, ok := del["COUNT"].(int32); !ok || count != 7 {
		t.Fatalf("deleted COUNT = %#v, want 7", del["COUNT"])
	}
	if ratio, ok := del["RATIO"].(float64); !ok || math.Abs(ratio-2.5) > 1e-9 {
		t.Fatalf("deleted RATIO = %#v, want approx 2.5", del["RATIO"])
	}
	wantStamp := time.Date(2019, time.December, 31, 0, 0, 0, 0, time.UTC)
	if stamp, ok := del["STAMP"].(time.Time); !ok || !stamp.Equal(wantStamp) {
		t.Fatalf("deleted STAMP = %#v, want %v", del["STAMP"], wantStamp)
	}
	if memo, ok := del["MEMO"]; !ok || memo != nil {
		t.Fatalf("deleted MEMO = %#v, want nil", del["MEMO"])
	}
}

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
