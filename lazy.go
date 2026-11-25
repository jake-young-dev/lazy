package lazy

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"math/rand/v2"
	"reflect"
	"regexp"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	DB_TAG    = "db"   //tag used to parse sql fields in example struct
	LAZY_TAG  = "lazy" //tag label for KEY_VALUE
	KEY_VALUE = "key"  //tag value for LAZY_TAG
	//bind values
	BindQuestion Binds = 1
	BindDollar   Binds = 2
	BindNamed    Binds = 3
	BindAt       Binds = 4
)

type Binds int

// mock data generated based on config
type Mock struct {
	Query   string
	Columns []string
	Rows    [][]driver.Value
}

// configuration for RandomGenerate
type Config struct {
	Query    string
	Example  any
	Keys     []any //length of keys must = RowCount, if set
	RowCount int
	BindType Binds
}

// generates mock data based on configuration to be used for sqlmock
func RandomGenerate(m Config) (*Mock, error) {
	//example struct cannot be nil and must be a struct
	if m.Example == nil {
		return nil, errors.New("example value cannot be nil")
	}
	if reflect.ValueOf(m.Example).Kind() != reflect.Struct {
		return nil, errors.New("example value must be a struct")
	}

	//any weirdness, just pull one row
	if m.RowCount <= 0 {
		m.RowCount = 1
	}

	//if keys are set, ensure there are enough to populate requested rows
	primaryKey := false
	if len(m.Keys) > 0 {
		primaryKey = true
		if len(m.Keys) != m.RowCount {
			return nil, errors.New("you must provide a key for each row")
		}
	}

	retType := reflect.TypeOf(m.Example)
	maxFieldCount := retType.NumField()
	columns := make([]string, 0, maxFieldCount)
	rows := make([][]driver.Value, 0)

	for y := 0; y < m.RowCount; y++ {
		rows = append(rows, make([]driver.Value, 0))
		for x := 0; x < maxFieldCount; x++ {
			field := retType.Field(x)
			dbTag := field.Tag.Get(DB_TAG)
			if dbTag == "" {
				continue
			}

			//track columns only once
			if y == 0 {
				columns = append(columns, dbTag)
			}

			//if field has lazy:"key" tag and tags are present inject primary key
			if field.Tag.Get(LAZY_TAG) == KEY_VALUE && primaryKey {
				rows[y] = append(rows[y], m.Keys[y])
				continue
			}

			nv := kindToRandom(field)
			if nv == nil {
				return nil, fmt.Errorf("could not match type: %s", retType.Name())
			}

			rows[y] = append(rows[y], nv)
		}
	}

	return &Mock{
		//sql is rebound and escaped for sqlmock.ExpectQuery
		Query:   sqlx.Rebind(int(m.BindType), regexp.QuoteMeta(m.Query)),
		Columns: columns,
		Rows:    rows,
	}, nil
}

// converts basic reflect Kind's to psuedo-random data, slices are given a random amount of entries
func kindToRandom(field reflect.StructField) any {
	//this isn't ideal, but since the sql types are structs they were being filled in a weird
	//manner. For now we check those types first, and will clean this up later.
	switch field.Type {
	case reflect.TypeOf(sql.NullTime{}):
		var temp sql.NullTime
		if !isNull() {
			temp.Valid = true
			temp.Time = time.Now()
		}
		return temp
	case reflect.TypeOf(sql.NullInt16{}):
		var temp sql.NullInt16
		if isNull() {
			temp.Valid = false
		} else {
			temp.Valid = true
			temp.Int16 = int16(rand.Int())
		}
		return temp
	case reflect.TypeOf(sql.NullInt32{}):
		var temp sql.NullInt32
		if isNull() {
			temp.Valid = false
		} else {
			temp.Valid = true
			temp.Int32 = rand.Int32()
		}
		return temp
	case reflect.TypeOf(sql.NullInt64{}):
		var temp sql.NullInt64
		if isNull() {
			temp.Valid = false
		} else {
			temp.Valid = true
			temp.Int64 = rand.Int64()
		}
		return temp
	case reflect.TypeOf(sql.NullBool{}):
		var temp sql.NullBool
		if isNull() {
			temp.Valid = false
		} else {
			temp.Valid = true
			temp.Bool = (rand.Int()%2 == 0)
		}
		return temp
	case reflect.TypeOf(sql.NullFloat64{}):
		var temp sql.NullFloat64
		if isNull() {
			temp.Valid = false
		} else {
			temp.Valid = true
			temp.Float64 = rand.Float64()
		}
		return temp
	case reflect.TypeOf(sql.NullString{}):
		var temp sql.NullString
		if isNull() {
			temp.Valid = false
		} else {
			temp.Valid = true
			temp.String = fmt.Sprintf("random %d", rand.Int())
		}
		return temp
	}
	kind := field.Type.Kind()
	switch kind {
	case reflect.Int:
		return rand.Int()
	case reflect.Int32:
		return rand.Int32()
	case reflect.Int64:
		return rand.Int64()
	case reflect.Float32:
		return rand.Float32()
	case reflect.Float64:
		return rand.Float64()
	case reflect.String:
		return fmt.Sprintf("random %d", rand.Int())
	case reflect.Bool:
		return rand.Int()%2 == 0
	case reflect.TypeOf(time.Time{}).Kind():
		return time.Now()
	case reflect.Array:
		underlying := field.Type.Elem().Kind()
		count := reflect.ValueOf(field).Len() //fill entire length
		switch underlying {
		case reflect.Int:
			var intslice []int
			for x := 0; x < count; x++ {
				intslice = append(intslice, rand.Int())
			}
			return intslice
		case reflect.Int32:
			var int32slice []int32
			for x := 0; x < count; x++ {
				int32slice = append(int32slice, rand.Int32())
			}
			return int32slice
		case reflect.Int64:
			var int64slice []int64
			for x := 0; x < count; x++ {
				int64slice = append(int64slice, rand.Int64())
			}
			return int64slice
		case reflect.Float32:
			var float32slice []float32
			for x := 0; x < count; x++ {
				float32slice = append(float32slice, rand.Float32())
			}
			return float32slice
		case reflect.Float64:
			var float64slice []float64
			for x := 0; x < count; x++ {
				float64slice = append(float64slice, rand.Float64())
			}
			return float64slice
		case reflect.String:
			var strslice []string
			for x := 0; x < count; x++ {
				strslice = append(strslice, fmt.Sprintf("%d", rand.Int()))
			}
			return strslice
		case reflect.Bool:
			var boolslice []bool
			for x := 0; x < count; x++ {
				boolslice = append(boolslice, (rand.Int()%2 == 0))
			}
			return boolslice
		}
	case reflect.Slice:
		underlying := field.Type.Elem().Kind()
		count := (rand.Int() % 10) + 1 //amount of entries to append to slice
		switch underlying {
		case reflect.Int:
			var intslice []int
			for x := 0; x < count; x++ {
				intslice = append(intslice, rand.Int())
			}
			return intslice
		case reflect.Int32:
			var int32slice []int32
			for x := 0; x < count; x++ {
				int32slice = append(int32slice, rand.Int32())
			}
			return int32slice
		case reflect.Int64:
			var int64slice []int64
			for x := 0; x < count; x++ {
				int64slice = append(int64slice, rand.Int64())
			}
			return int64slice
		case reflect.Float32:
			var float32slice []float32
			for x := 0; x < count; x++ {
				float32slice = append(float32slice, rand.Float32())
			}
			return float32slice
		case reflect.Float64:
			var float64slice []float64
			for x := 0; x < count; x++ {
				float64slice = append(float64slice, rand.Float64())
			}
			return float64slice
		case reflect.String:
			var strslice []string
			for x := 0; x < count; x++ {
				strslice = append(strslice, fmt.Sprintf("random %d", rand.Int()))
			}
			return strslice
		case reflect.Bool:
			var boolslice []bool
			for x := 0; x < count; x++ {
				boolslice = append(boolslice, (rand.Int()%2 == 0))
			}
			return boolslice
		}
	}

	return nil
}

// a quick and dirty rng function to determine if nulltypes are null or
// if we will generate a random value
func isNull() bool {
	x := rand.IntN(2)
	if x%2 == 0 {
		return true
	} else {
		return false
	}
}
