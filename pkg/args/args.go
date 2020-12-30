package args

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

const (
	emptyValue = ""
	separate   = ','
)

type List struct {
	values []string
	idx    int
}

func New(s string) *List {
	values := strings.FieldsFunc(s, func(r rune) bool {
		return r == separate
	})

	return &List{values: values}
}

func (l *List) Next() string {
	idx := l.idx
	l.idx += 1

	if idx < len(l.values) {
		return l.values[idx]
	}

	return emptyValue
}

func (l *List) NextUUID() uuid.UUID {
	v := l.Next()
	id, _ := uuid.FromString(v)
	return id
}

func (l *List) NextInt() int {
	v := l.Next()
	i, _ := strconv.Atoi(v)
	return i
}

func (l *List) NextInt64() int64 {
	v := l.Next()
	i, _ := strconv.ParseInt(v, 10, 64)
	return i
}

func (l *List) NextDecimal() decimal.Decimal {
	v := l.Next()
	d, _ := decimal.NewFromString(v)
	return d
}

func (l *List) NextDuration() time.Duration {
	v := l.Next()
	d, _ := time.ParseDuration(v)
	return d
}
