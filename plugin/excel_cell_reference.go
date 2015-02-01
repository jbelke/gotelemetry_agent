package plugin

import (
	"math"
	"strconv"
	"strings"
)

type excelCellReference struct {
	Row    int
	Column int
}

func newExcelCellReference(column, row string) excelCellReference {
	result := excelCellReference{}

	column = strings.ToUpper(column)

	l := len(column) - 1

	for index, char := range column {
		result.Column += int(float64((char-'A')+1)*(math.Pow(26, float64(l-index)))) - 1
	}

	var err error

	result.Row, err = strconv.Atoi(row)

	if err != nil {
		panic(err) // This should never happen, since we always come here from a regex that only matches digits
	}

	return result
}
