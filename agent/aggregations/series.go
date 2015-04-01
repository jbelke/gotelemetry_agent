package aggregations

import (
	"time"
)

type Series struct {
	context *Context
	Name    string
}

func createSeries(context *Context, name string) error {
	result := &Series{
		context: context,
		Name:    name,
	}

	if err := result.createTable(); err != nil {
		return err
	}

	cachedSeries[name] = result

	return nil
}

var cachedSeries = map[string]*Series{}

func GetSeries(context *Context, name string) (*Series, error) {
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

func (s *Series) createTable() error {
	c := s.context.conn

	if err := c.Exec("CREATE TABLE IF NOT EXISTS `" + s.Name + "` (timestamp DATETIME NOT NULL, value FLOAT)"); err != nil {
		return err
	}

	if err := c.Exec("CREATE INDEX IF NOT EXISTS `" + s.Name + "_index` ON " + s.Name + " (timestamp)"); err != nil {
		return err
	}

	return nil
}

func (s *Series) Push(timestamp *time.Time, value float64) error {
	if timestamp == nil {
		timestamp = &time.Time{}
		*timestamp = time.Now()
	}

	return s.context.conn.Exec("INSERT INTO `"+s.Name+"` (timestamp, value) VALUES (?, ?)", *timestamp, value)
}
