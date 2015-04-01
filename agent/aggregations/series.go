package aggregations

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

type FunctionType int

const (
	Sum FunctionType = iota
)

type Series struct {
	context *Context
	Name    string
}

var cachedSeries = map[string]*Series{}

var seriesNameRegex = regexp.MustCompile("^[A-Za-z_][A-Za-z0-9_]*$")

func GetSeries(context *Context, name string) (*Series, error) {
	if err := validateSeriesName(name); err != nil {
		return nil, err
	}

	result := &Series{
		context: context,
		Name:    name,
	}

	if _, ok := cachedSeries[name]; !ok {
		if err := createSeries(context, name); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (s *Series) Push(timestamp *time.Time, value float64) error {
	if timestamp == nil {
		timestamp = &time.Time{}
		*timestamp = time.Now()
	}

	return s.exec("INSERT INTO ?? (ts, value) VALUES (?, ?)", *timestamp, value)
}

func (s *Series) last() (map[string]interface{}, error) {
	return s.fetchRow("SELECT rowid, ts, value FROM ?? ORDER BY ts DESC LIMIT 1")
}

func (s *Series) Last() (map[string]interface{}, error) {
	row, err := s.last()

	if err != nil {
		return nil, err
	}

	delete(row, "rowid")

	return row, nil
}

func (s *Series) Pop(shouldDelete bool) (map[string]interface{}, error) {
	row, err := s.last()

	if err != nil {
		return nil, err
	}

	rowId := row["rowid"]

	delete(row, "rowid")

	if shouldDelete {
		s.exec("DELETE FROM ?? WHERE rowid = ?", rowId)
	}

	return row, nil
}

func (s *Series) Compute(functionType FunctionType, start, end *time.Time) (float64, error) {
	var operation string

	switch functionType {
	case Sum:
		operation = "SUM"

	default:
		return 0.0, errors.New(fmt.Sprintf("Unknown operation %d", functionType))
	}

	if start == nil {
		start = &time.Time{}
		*start = time.Unix(0, 0)
	}

	if end == nil {
		end = &time.Time{}
		*end = time.Now()
	}

	row, err := s.fetchRow("SELECT "+operation+"(value) AS result FROM ?? WHERE ts BETWEEN ? AND ?", *start, *end)

	if err != nil {
		return 0.0, err
	}

	return row["result"].(float64), nil
}
