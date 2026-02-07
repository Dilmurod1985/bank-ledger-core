package services

import (
	"errors"
	"math/big"
	"strings"
)

func parseDecimal(s string) (*big.Float, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, errors.New("empty decimal string")
	}

	decimal := new(big.Float)
	_, success := decimal.SetString(s)
	if !success {
		return nil, errors.New("invalid decimal format")
	}

	return decimal, nil
}
