package main

func frame(data string) []byte {
	b := []byte(data)
	out := make([]byte, 0, len(b)+2)
	out = append(out, 0x02)
	out = append(out, b...)
	out = append(out, 0x03)
	return out
}
