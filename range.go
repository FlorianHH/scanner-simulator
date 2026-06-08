package main

import (
	"fmt"
	"strconv"
)

func ssccCheckDigit(s string) byte {
	sum := 0
	for i, ch := range s {
		d := int(ch - '0')
		if (len(s)-1-i)%2 == 0 {
			d *= 3
		}
		sum += d
	}
	return byte('0' + (10-sum%10)%10)
}

func generateRange(start uint64, count int, checkDigit bool) []string {
	items := make([]string, count)
	for i := 0; i < count; i++ {
		v := start + uint64(i)
		if checkDigit {
			payload := fmt.Sprintf("%017d", v)
			items[i] = payload + string(ssccCheckDigit(payload))
		} else {
			items[i] = fmt.Sprintf("%d", v)
		}
	}
	return items
}

func (a *App) StartRange(start string, count int, delayMs int, checkDigit bool) error {
	if count <= 0 {
		return fmt.Errorf("count must be > 0")
	}
	startVal, err := strconv.ParseUint(start, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid start number")
	}
	if checkDigit {
		const maxSSCC = uint64(99999999999999999) // 17-digit max
		if startVal > maxSSCC {
			return fmt.Errorf("start exceeds 17-digit maximum")
		}
		if uint64(count-1) > maxSSCC-startVal {
			return fmt.Errorf("range end exceeds 17-digit maximum")
		}
	}
	items := generateRange(startVal, count, checkDigit)
	return a.StartBatch(items, delayMs, 1)
}
