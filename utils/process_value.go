package utils

import (
	"database/sql"
	"fmt"
	"time"
)

func ProcessValue(val interface{}) (interface{}, error) {
	switch v := val.(type) {
	case nil:
		return nil, nil
	case bool:
		return v, nil
	case int:
		return v, nil
	case int8:
		return v, nil
	case int16:
		return v, nil
	case int32:
		return v, nil
	case int64:
		return v, nil
	case uint:
		return v, nil
	case uint8:
		return v, nil
	case uint16:
		return v, nil
	case uint32:
		return v, nil
	case uint64:
		return v, nil
	case float32:
		return v, nil
	case float64:
		return v, nil
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case time.Time:
		return v.Format(time.RFC3339), nil
	case sql.NullBool:
		if v.Valid {
			return v.Bool, nil
		}
		return nil, nil
	case sql.NullInt64:
		if v.Valid {
			return v.Int64, nil
		}
		return nil, nil
	case sql.NullFloat64:
		if v.Valid {
			return v.Float64, nil
		}
		return nil, nil
	case sql.NullString:
		if v.Valid {
			return v.String, nil
		}
		return nil, nil
	case sql.NullTime:
		if v.Valid {
			return v.Time.Format(time.RFC3339), nil
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}
}
