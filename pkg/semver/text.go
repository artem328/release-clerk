package semver

func (v Version) MarshalText() ([]byte, error) {
	return v.bytes(), nil
}

func (v *Version) UnmarshalText(text []byte) error {
	vv, err := ParseBytes(text)
	if err != nil {
		return err
	}

	*v = vv

	return nil
}
