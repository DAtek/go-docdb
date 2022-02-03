package docdb

import (
	sqldriver "database/sql/driver"
	"io"
	"time"

	"github.com/domonda/go-errs"
	"github.com/domonda/go-pretty"
	"github.com/domonda/go-types/nullable"
)

const (
	// VersionTimeFormat is the string format of a version time
	// returned by VersionTime.String() and parsed by VersionTimeFromString.
	VersionTimeFormat = "2006-01-02_15-04-05.000"

	sqlTimeFormat = "2006-01-02 15:04:05.999"
)

// VersionTime of a document.
// VersionTime implements the database/sql.Scanner and database/sql/driver.Valuer interfaces
// and will treat a zero VersionTime value as SQL NULL value.
type VersionTime struct {
	Time time.Time
}

// VersionTimeFrom returns a VersionTime for the given time translated to UTC and truncated to milliseconds
func VersionTimeFrom(t time.Time) VersionTime {
	if t.IsZero() {
		return VersionTime{}
	}
	return VersionTime{Time: t.UTC().Truncate(time.Millisecond)}
}

// VersionTimeFromString parses a string as VersionTime.
// The strings "", "null", "NULL" will be parsed as null VersionTime.
func VersionTimeFromString(str string) (VersionTime, error) {
	if str == "" || str == "null" || str == "NULL" {
		return VersionTime{}, nil
	}
	t, err := time.ParseInLocation(VersionTimeFormat, str, time.UTC)
	if err != nil {
		// Try again with SQL time format:
		t, err = time.ParseInLocation(sqlTimeFormat, str, time.UTC)
		if err != nil {
			return VersionTime{}, errs.Errorf("error parsing %q as docdb.VersionTime: %w", str, err)
		}
	}
	return VersionTime{Time: t}, nil
}

// String implements the fmt.Stringer interface.
func (v VersionTime) String() string {
	if v.IsNull() {
		return ""
	}
	return v.Time.Format(VersionTimeFormat)
}

// NullableTime returns the version time as nullable.Time
func (v VersionTime) NullableTime() nullable.Time {
	return nullable.TimeFrom(v.Time)
}

// PrettyPrint implements the pretty.Printable interface
func (v VersionTime) PrettyPrint(w io.Writer) {
	if v.IsNull() {
		pretty.Fprint(w, nil)
	} else {
		pretty.Fprint(w, v.Time.Format(VersionTimeFormat))
	}
}

// MarshalText implements the encoding.TextMarshaler interface
func (v VersionTime) MarshalText() (text []byte, err error) {
	return []byte(v.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (v *VersionTime) UnmarshalText(text []byte) error {
	vt, err := VersionTimeFromString(string(text))
	if err != nil {
		return err
	}
	*v = vt
	return nil
}

func (v VersionTime) After(other VersionTime) bool {
	// Truncate(time.Millisecond) on both times just to make sure it's comparable
	return v.Time.Truncate(time.Millisecond).After(other.Time.Truncate(time.Millisecond))
}

func (v VersionTime) Before(other VersionTime) bool {
	// Truncate(time.Millisecond) on both times just to make sure it's comparable
	return v.Time.Truncate(time.Millisecond).Before(other.Time.Truncate(time.Millisecond))
}

func (v VersionTime) Equal(other VersionTime) bool {
	// Truncate(time.Millisecond) on both times just to make sure it's comparable
	return v.Time.Truncate(time.Millisecond).Equal(other.Time.Truncate(time.Millisecond))
}

func (v VersionTime) IsNull() bool {
	return v.Time.IsZero()
}

func (v VersionTime) IsNotNull() bool {
	return !v.Time.IsZero()
}

// Scan implements the database/sql.Scanner interface.
func (v *VersionTime) Scan(value interface{}) error {
	switch t := value.(type) {
	case nil:
		*v = VersionTime{}
		return nil

	case time.Time:
		*v = VersionTimeFrom(t)
		return nil

	case []byte:
		vt, err := VersionTimeFromString(string(t))
		if err != nil {
			return err
		}
		*v = vt
		return nil

	case string:
		vt, err := VersionTimeFromString(t)
		if err != nil {
			return err
		}
		*v = vt
		return nil

	default:
		return errs.Errorf("can't scan %T as docdb.VersionTime", value)
	}
}

// Value implements the driver database/sql/driver.Valuer interface.
func (v VersionTime) Value() (sqldriver.Value, error) {
	if v.IsNull() {
		return nil, nil
	}
	return v.Time.Truncate(time.Millisecond), nil
}
