package date

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
)

type Date time.Time

const FormatDate = "2006-01-02"

func (d Date) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(d).Format(FormatDate))
	return []byte(stamp), nil
}

func (d *Date) UnmarshalJSON(bytes []byte) error {
	t, err := time.Parse(FormatDate, string(bytes[1:len(bytes)-1]))
	if err != nil {
		return errors.Wrapf(err, "can't parse date %s", bytes)
	}
	*d = Date(t)
	return nil
}

func (d Date) Equal(r Date) bool {
	return time.Time(d).Equal(time.Time(r))
}

func (d Date) EqualTime(r time.Time) bool {
	return time.Time(d).Equal(r)
}
