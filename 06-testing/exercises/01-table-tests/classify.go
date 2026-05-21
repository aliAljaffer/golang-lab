// Package classify maps a numeric score to a letter grade.
package classify

// Classify returns the letter grade for an integer score in [0, 100].
// Returns "?" for scores outside that range.
func Classify(score int) string {
	switch {
	case score < 0 || score > 100:
		return "?"
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}
