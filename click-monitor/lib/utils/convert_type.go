package utils

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
)

func ConvertToInt(from interface{}) (int, error) {
	switch from.(type) {
	case nil:
		return 0, nil
	case float32:
		return int(from.(float32)), nil
	case float64:
		return int(from.(float64)), nil
	case int8:
		return int(from.(int8)), nil
	case int16:
		return int(from.(int16)), nil
	case int32:
		return int(from.(int32)), nil
	case int64:
		return int(from.(int64)), nil
	case uint:
		return int(from.(uint)), nil
	case uint8:
		return int(from.(uint8)), nil
	case uint16:
		return int(from.(uint16)), nil
	case uint32:
		return int(from.(uint32)), nil
	case uint64:
		return int(from.(uint64)), nil
	case string:
		return strconv.Atoi(from.(string))
	case int:
		return from.(int), nil
	case bool:
		b := from.(bool)
		if b {
			return 1, nil
		}
		return 0, nil
	}

	return 0, errors.New(fmt.Sprintf("ConvertToInt error: unknown FROM type %T", from))
}

func ConvertToBool(from interface{}) (bool, error) {
	switch from.(type) {
	case nil:
		return false, nil
	case string:
		tmp, err := ConvertToString(from)
		return tmp != "", err
	case int, int32, int16, int8, int64, uint32, uint8, uint64, uint16, uint, float32, float64:
		tmp, err := ConvertToInt(from)
		return tmp > 0, err
	case bool:
		return from.(bool), nil
	}

	return false, errors.New(fmt.Sprintf("ConvertToBool error: unknown FROM type %T", from))
}

func ConvertToUint32(from interface{}) (uint32, error) {
	val, err := ConvertToInt(from)

	return uint32(val), err
}

func ConvertToUint8(from interface{}) (uint8, error) {
	val, err := ConvertToInt(from)

	return uint8(val), err
}

func ConvertToSqlNullInt32(from interface{}) (sql.NullInt32, error) {
	val, err := ConvertToInt(from)

	return sql.NullInt32{
		Int32: int32(val),
		Valid: int32(val) != 0,
	}, err
}

func ConvertToString(from interface{}) (string, error) {
	switch from.(type) {
	case nil:
		return "", nil
	case float32, float64:
		return fmt.Sprintf("%f", from), nil
	case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64:
		return strconv.Itoa(from.(int)), nil
	case string:
		return from.(string), nil
	}

	return "", errors.New(fmt.Sprintf("ConvertToString error: unknown FROM type %T", from))
}

func ConvertToSQlNullString(from interface{}) (sql.NullString, error) {
	val, err := ConvertToString(from)

	return sql.NullString{
		String: val,
		Valid:  val != "",
	}, err
}
