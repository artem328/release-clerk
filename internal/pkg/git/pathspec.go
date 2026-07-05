package git

func PathSpec(from, to string) string {
	if to == "" {
		to = "HEAD"
	}

	if from != "" {
		return from + ".." + to
	}

	return to
}
