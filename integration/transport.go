package integration

import (
	"fmt"
	"net"
)

const (
	portRangeMax = 65535
)

var (
	errPortMaxExceeded   = portError{message: "port range exceeded"}
	errNegativePortValue = portError{message: "cannot have a negative port value"}
)

type portError struct{ message string }

// Error returns a string with the port-related error message
func (p portError) Error() string {
	return fmt.Sprintf("port discovery error: %s", p.message)
}

// GetOpenPortInRange finds an unused port within a specified range
func GetOpenPortInRange(lowerBound, upperBound int) (int, error) {
	if lowerBound < 0 {
		return -1, errNegativePortValue
	}
	for lowerBound <= portRangeMax && lowerBound <= upperBound {
		if _, err := net.Dial("tcp", fmt.Sprintf(":%d", lowerBound)); err != nil {
			return lowerBound, nil
		}
		lowerBound++
	}
	return -1, errPortMaxExceeded
}
