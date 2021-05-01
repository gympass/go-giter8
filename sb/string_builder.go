package sb

type StringBuilder struct {
	inner []rune
	cur   int
}

// New creates a new StringBuilder instance
func New() *StringBuilder {
	return &StringBuilder{
		inner: make([]rune, 0, 2048),
		cur:   0,
	}
}

// WriteRunes appends a given rune to the StringBuilder
func (sb *StringBuilder) WriteRune(r rune) {
	sb.inner = append(sb.inner, r)
	sb.cur++
}

// DeleteLast deletes the last rune in this StringBuilder
func (sb *StringBuilder) DeleteLast() {
	sb.Delete(1)
}

// Delete removes a given amount of runes from this StringBuilder
func (sb *StringBuilder) Delete(delta int) {
	sb.cur -= delta
	sb.inner = sb.inner[0:sb.cur]
}

// String composes and returns a given string based on this StringBuilder
// contents
func (sb *StringBuilder) String() string {
	return string(sb.inner[0:sb.cur])
}

// Reset clears all contents from this StringBuilder
func (sb *StringBuilder) Reset() {
	sb.inner = sb.inner[:0]
	sb.cur = 0
}

// Len returns the amount of runes contained within this StringBuilder
func (sb *StringBuilder) Len() int {
	return sb.cur
}
