package dbfmini

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// --------------------------- Tipos públicos ---------------------------

type ReadMode string

const (
	ReadStrict ReadMode = "strict"
	ReadLoose  ReadMode = "loose"
)

// Encoding permite um default e (opcional) overrides por campo.
type Encoding struct {
	Default  string
	PerField map[string]string // ex.: {"NOME": "CP1252"}
}

type OpenOptions struct {
	ReadMode       ReadMode
	Encoding       Encoding // Default: ISO-8859-1
	IncludeDeleted bool
}

type Field struct {
	Name          string
	Type          byte // 'C','N','F','Y','L','D','I','M','T','B'
	Size          uint8
	DecimalPlaces uint8
}

type DBF struct {
	Path          string
	RecordCount   uint32
	DateOfLastUpd time.Time
	Fields        []Field

	// internos
	version     byte
	headerLen   uint16
	recordLen   uint16
	memoPath    string
	opt         OpenOptions
	recordsRead uint32
}

// Record é um mapa com valores tipados por Go nativo.
// Se IncludeDeleted=true, adicionamos a chave "_deleted": bool.
type Record map[string]any

// --------------------------- Abertura ---------------------------

func Open(path string, opts *OpenOptions) (*DBF, error) {
	if opts == nil {
		opts = &OpenOptions{}
	}
	normOptions(opts)

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	hdr := make([]byte, 32)
	if _, err := io.ReadFull(f, hdr); err != nil {
		return nil, fmt.Errorf("lendo header: %w", err)
	}

	version := hdr[0]
	yy := int(hdr[1]) + 1900
	mm := int(hdr[2])
	dd := int(hdr[3])
	date := time.Date(yy, time.Month(mm), dd, 0, 0, 0, 0, time.UTC)

	recCount := binary.LittleEndian.Uint32(hdr[4:8])
	headerLen := binary.LittleEndian.Uint16(hdr[8:10])
	recordLen := binary.LittleEndian.Uint16(hdr[10:12])

	// Checa versão (apenas se strict)
	if opts.ReadMode == ReadStrict && !isValidVersion(version) {
		return nil, fmt.Errorf("versão dBase não suportada: 0x%02x", version)
	}

	// Localiza arquivo de memo (quando aplicável)
	memoPath := ""
	ext := strings.ToLower(filepath.Ext(path))
	if version == 0x83 || version == 0x8b { // dBase III/IV com memo .dbt
		for _, e := range []string{".dbt", ".DBT"} {
			p := path[:len(path)-len(ext)] + e
			if _, err := os.Stat(p); err == nil {
				memoPath = p
				break
			}
		}
		if memoPath == "" && opts.ReadMode == ReadStrict {
			return nil, errors.New("memo .DBT não encontrado (modo strict)")
		}
	}
	if version == 0x30 || version == 0xf5 { // VFP/FoxPro podem usar .fpt
		for _, e := range []string{".fpt", ".FPT"} {
			p := path[:len(path)-len(ext)] + e
			if _, err := os.Stat(p); err == nil {
				memoPath = p
				break
			}
		}
	}

	// Lê descritores de campos (32 bytes cada) até 0x0D
	var fields []Field
	pos := int64(32)
	for {
		if pos >= int64(headerLen) {
			break
		}
		des := make([]byte, 32)
		if _, err := f.ReadAt(des, pos); err != nil {
			return nil, fmt.Errorf("lendo field descriptor: %w", err)
		}
		pos += 32

		if des[0] == 0x0D { // terminador
			break
		}

		// Nome (0..10 com NUL terminator)
		nameRaw := des[0:11]
		zero := bytes.IndexByte(nameRaw, 0)
		if zero == -1 {
			zero = 11
		}
		name := decodeBytes(nameRaw[:zero], fieldEncoding(opts.Encoding, "")) // nome usa encoding default

		ftype := des[11] // 0x0B
		size := des[16]  // field length
		dec := des[17]   // decimal count

		field := Field{
			Name:          strings.TrimSpace(name),
			Type:          ftype,
			Size:          size,
			DecimalPlaces: dec,
		}

		// Validações básicas em modo strict
		if opts.ReadMode == ReadStrict {
			if err := validateField(field, version); err != nil {
				return nil, err
			}
			for _, ex := range fields {
				if ex.Name == field.Name {
					return nil, fmt.Errorf("nome de campo duplicado: %s", field.Name)
				}
			}
		}
		fields = append(fields, field)
	}

	// Confere comprimento de registro
	calculated := calcRecordLen(fields)
	if opts.ReadMode == ReadStrict && calculated != recordLen {
		return nil, fmt.Errorf("tamanho de registro inconsistente: header=%d calculado=%d", recordLen, calculated)
	}

	db := &DBF{
		Path:          path,
		RecordCount:   recCount,
		DateOfLastUpd: date,
		Fields:        fields,
		version:       version,
		headerLen:     headerLen,
		recordLen:     recordLen,
		memoPath:      memoPath,
		opt:           *opts,
	}
	return db, nil
}

// --------------------------- Leitura de registros ---------------------------

// ReadRecords lê até maxCount; se maxCount<=0 lê até o fim.
func (d *DBF) ReadRecords(maxCount int) ([]Record, error) {
	if maxCount <= 0 {
		maxCount = int(d.RecordCount - d.recordsRead)
	}
	if d.recordsRead >= d.RecordCount || maxCount == 0 {
		return []Record{}, nil
	}

	f, err := os.Open(d.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	start := int64(d.headerLen) + int64(d.recordsRead)*int64(d.recordLen)

	var out []Record
	for i := 0; i < maxCount && d.recordsRead < d.RecordCount; i++ {
		buf := make([]byte, d.recordLen)
		if _, err := f.ReadAt(buf, start); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("lendo registro: %w", err)
		}
		start += int64(d.recordLen)
		d.recordsRead++

		rec, skip, err := d.parseRecord(buf)
		if err != nil {
			if d.opt.ReadMode == ReadStrict {
				return nil, err
			}
			if skip {
				continue
			}
		}
		if rec != nil {
			out = append(out, rec)
		}
	}
	return out, nil
}

// Reset reinicia o cursor interno para nova leitura.
func (d *DBF) Reset() { d.recordsRead = 0 }

// Version retorna o byte de versão do arquivo DBF.
func (d *DBF) Version() byte { return d.version }

// --------------------------- Parsing de registro ---------------------------

func (d *DBF) parseRecord(b []byte) (Record, bool, error) {
	if len(b) == 0 {
		return nil, true, nil
	}
	deleted := b[0] == 0x2A // '*' = deletado; ' ' = normal
	if deleted && !d.opt.IncludeDeleted {
		return nil, true, nil
	}

	rec := Record{}
	offset := 1

	for _, f := range d.Fields {
		if offset+int(f.Size) > len(b) {
			return nil, true, fmt.Errorf("registro truncado")
		}
		fieldBytes := b[offset : offset+int(f.Size)]
		offset += int(f.Size)

		enc := fieldEncoding(d.opt.Encoding, f.Name)
		switch f.Type {
		case 'C': // texto
			val := rtrimSpaces(decodeBytes(fieldBytes, enc))
			rec[f.Name] = val

		case 'N', 'F': // número/float (ASCII)
			s := strings.TrimSpace(decodeBytes(fieldBytes, enc))
			if s == "" {
				rec[f.Name] = nil
				break
			}
			// aceita vírgula decimal
			if strings.Contains(s, ",") && !strings.Contains(s, ".") {
				s = strings.ReplaceAll(s, ",", ".")
			}
			fl, err := strconv.ParseFloat(s, 64)
			if err != nil {
				if d.opt.ReadMode == ReadStrict {
					return nil, true, fmt.Errorf("%s: número inválido: %q", f.Name, s)
				}
				rec[f.Name] = nil
			} else {
				rec[f.Name] = fl
			}

		case 'Y': // currency 64 bits int / 10000 (LE)
			if len(fieldBytes) != 8 {
				rec[f.Name] = nil
				break
			}
			u := binary.LittleEndian.Uint64(fieldBytes)
			i := int64(u)
			rec[f.Name] = float64(i) / 10000.0

		case 'L': // lógico
			c := byte(' ')
			if len(fieldBytes) > 0 {
				c = fieldBytes[0]
			}
			switch c {
			case 'T', 't', 'Y', 'y':
				rec[f.Name] = true
			case 'F', 'f', 'N', 'n':
				rec[f.Name] = false
			default:
				rec[f.Name] = nil
			}

		case 'D': // data "YYYYMMDD"
			s := decodeBytes(fieldBytes, enc)
			s = strings.TrimSpace(s)
			if len(s) != 8 || strings.Contains(s, " ") || s == "00000000" {
				rec[f.Name] = nil
				break
			}
			t, err := time.Parse("20060102", s)
			if err != nil {
				rec[f.Name] = nil
			} else {
				rec[f.Name] = t
			}

		case 'I': // int32 LE
			if len(fieldBytes) != 4 {
				rec[f.Name] = nil
				break
			}
			rec[f.Name] = int32(binary.LittleEndian.Uint32(fieldBytes))

		case 'B': // double LE
			if len(fieldBytes) != 8 {
				rec[f.Name] = nil
				break
			}
			bits := binary.LittleEndian.Uint64(fieldBytes)
			rec[f.Name] = math.Float64frombits(bits)

		case 'T': // VFP DateTime (julian int32 LE, msSinceMidnight int32 LE)
			if len(fieldBytes) != 8 {
				rec[f.Name] = nil
				break
			}
			jd := int32(binary.LittleEndian.Uint32(fieldBytes[:4]))
			ms := int32(binary.LittleEndian.Uint32(fieldBytes[4:8]))
			rec[f.Name] = vfpDateTimeToUTC(int(jd), int(ms))

		case 'M':
			// TODO: Implementar leitura de memo (DBT/FPT). Por ora:
			if d.opt.ReadMode == ReadStrict {
				return nil, true, fmt.Errorf("campo memo não suportado (ainda)")
			}
			rec[f.Name] = nil

		default:
			if d.opt.ReadMode == ReadStrict {
				return nil, true, fmt.Errorf("tipo de campo não suportado: %q", string(f.Type))
			}
			// loose -> ignora
		}
	}

	if deleted {
		rec["_deleted"] = true
	}
	return rec, false, nil
}

// --------------------------- Utilitários ---------------------------

func normOptions(o *OpenOptions) {
	if o.ReadMode != ReadStrict && o.ReadMode != ReadLoose {
		o.ReadMode = ReadStrict
	}
	if o.Encoding.Default == "" {
		o.Encoding.Default = "ISO-8859-1"
	}
	if o.Encoding.PerField == nil {
		o.Encoding.PerField = map[string]string{}
	}
}

func isValidVersion(v byte) bool {
	switch v {
	case 0x03, 0x83, 0x8b, 0x30, 0xf5:
		return true
	default:
		return false
	}
}

func validateField(f Field, version byte) error {
	if f.Name == "" || len(f.Name) > 10 {
		return fmt.Errorf("nome de campo inválido: %q", f.Name)
	}
	switch f.Type {
	case 'C', 'N', 'F', 'Y', 'L', 'D', 'I', 'M', 'T', 'B':
	default:
		return fmt.Errorf("tipo não suportado: %q", string(f.Type))
	}
	// checks de tamanho básicos (equivalentes ao exemplo TS)
	if f.Type == 'C' && f.Size > 255 {
		return fmt.Errorf("%s: tamanho >255", f.Name)
	}
	if (f.Type == 'N' || f.Type == 'F') && f.Size > 20 {
		return fmt.Errorf("%s: tamanho >20", f.Name)
	}
	if f.Type == 'Y' && f.Size != 8 {
		return fmt.Errorf("%s: currency deve ter 8 bytes", f.Name)
	}
	if f.Type == 'L' && f.Size != 1 {
		return fmt.Errorf("%s: lógico deve ter 1 byte", f.Name)
	}
	if f.Type == 'D' && f.Size != 8 {
		return fmt.Errorf("%s: data deve ter 8 bytes", f.Name)
	}
	if f.Type == 'T' && f.Size != 8 {
		return fmt.Errorf("%s: datetime deve ter 8 bytes", f.Name)
	}
	if f.Type == 'B' && f.Size != 8 {
		return fmt.Errorf("%s: double deve ter 8 bytes", f.Name)
	}
	// memo size (DBT dBaseIII=10, VFP=4)
	memoSize := uint8(10)
	if version == 0x30 {
		memoSize = 4
	}
	if f.Type == 'M' && f.Size != memoSize {
		return fmt.Errorf("%s: memo size inválido (esperado %d)", f.Name, memoSize)
	}
	return nil
}

func calcRecordLen(fields []Field) uint16 {
	sum := 1 // flag de deletado
	for _, f := range fields {
		sum += int(f.Size)
	}
	return uint16(sum)
}

func rtrimSpaces(s string) string {
	return strings.TrimRight(s, " ")
}

func fieldEncoding(enc Encoding, field string) string {
	if v, ok := enc.PerField[field]; ok && v != "" {
		return v
	}
	return enc.Default
}

func decodeBytes(data []byte, enc string) string {
	enc = strings.ToUpper(strings.TrimSpace(enc))
	var tr *transform.Reader
	switch enc {
	case "CP850":
		tr = transform.NewReader(bytes.NewReader(data), charmap.CodePage850.NewDecoder())
	case "CP437":
		tr = transform.NewReader(bytes.NewReader(data), charmap.CodePage437.NewDecoder())
	case "CP1252", "WINDOWS-1252":
		tr = transform.NewReader(bytes.NewReader(data), charmap.Windows1252.NewDecoder())
	case "ISO-8859-1", "ISO8859-1", "LATIN1":
		tr = transform.NewReader(bytes.NewReader(data), charmap.ISO8859_1.NewDecoder())
	case "UTF-8", "UTF8":
		return string(data)
	default:
		// fallback: ISO-8859-1
		tr = transform.NewReader(bytes.NewReader(data), charmap.ISO8859_1.NewDecoder())
	}
	out, _ := io.ReadAll(tr)
	return string(out)
}

// ---------------- VFP DateTime (juliano <-> UTC) ----------------

func vfpDateTimeToUTC(julianDay int, msSinceMidnight int) time.Time {
	// Algoritmo equivalente ao do seu TS
	s1 := julianDay + 68569
	n := (4 * s1) / 146097
	s2 := s1 - (146097*n+3)/4
	i := (4000 * (s2 + 1)) / 1461001
	s3 := s2 - (1461*i)/4 + 31
	q := (80 * s3) / 2447
	s4 := q / 11
	year := 100*(n-49) + i + s4
	month := q + 2 - 12*s4
	day := s3 - (2447*q)/80

	sec := msSinceMidnight / 1000
	min := sec / 60
	hour := min / 60
	min = min % 60
	sec = sec % 60
	return time.Date(year, time.Month(month), day, hour, min, sec, 0, time.UTC)
}
