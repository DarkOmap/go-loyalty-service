package luhnalg

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
)

var ErrInvalidNumber error = fmt.Errorf("invalid order number")

func CheckNumber(number int) bool {
	var sum int

	for i := 0; number != 0; i++ {
		a := number % 10

		if i%2 != 0 {
			a *= 2

			if a > 9 {
				a -= 9
			}
		}

		sum += a
		number /= 10
	}

	return sum%10 == 0
}

func GetNumberFromBody(body io.ReadCloser) (int, error) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(body)

	if err != nil {
		return 0, fmt.Errorf("invalid body: %w", err)
	}

	id, err := strconv.Atoi(buf.String())

	if err != nil {
		return 0, fmt.Errorf("ivalid parsing: %w", err)
	}

	validID := CheckNumber(id)

	if !validID {
		return 0, ErrInvalidNumber
	}

	return id, nil
}
