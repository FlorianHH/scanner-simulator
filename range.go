package main

import "fmt"

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
