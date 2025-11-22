package time

import (
	"fmt"
	"time"
)

const timeStr = "2024-08-26T20:02:37.726556848+03:00"
const ConstantTimestamp = 1724684557 // Unix timestamp

type ConstantProvider struct {
	time time.Time
}

func NewConstantProvider() *ConstantProvider {
	parsedTime, err := time.Parse(time.RFC3339Nano, timeStr)
	if err != nil {
		panic(fmt.Sprintf("time parse: %v", err))
	}
	return &ConstantProvider{time: parsedTime}
}

func (p ConstantProvider) Now() time.Time {
	return p.time
}

