package models

// shouldApplyPrefixToMetric returns whether the prefix should be prepended
// to the metric name, taking into account a limit of consecutive prefixes that
// the metric name can contain.
func shouldApplyPrefixToMetric(name, prefix string, consecutiveNamePrefixLimit int) bool {
	// not using the limit
	if consecutiveNamePrefixLimit <= 0 {
		return true
	}

	lenName := len(name)
	lenPrefix := len(prefix)
	// cannot reach limit of consecutive prefixes
	if lenName < (lenPrefix * consecutiveNamePrefixLimit) {
		return true
	}

	for i := lenPrefix; i < lenName; i += lenPrefix {
		subst := name[i-lenPrefix : i]
		// the current prefix length subst of the metric name
		// does not equal prefix, cannot reach limit
		if subst != prefix {
			return true
		}
	}

	return false
}
