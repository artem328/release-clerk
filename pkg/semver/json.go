package semver

import "encoding/json"

func (v Version) MarshalJSON() ([]byte, error) {
	type jsonModel Version

	var object = struct {
		String  string
		Version jsonModel
	}{
		String:  v.String(),
		Version: jsonModel(v),
	}

	return json.Marshal(object)
}

func (v *Version) UnmarshalJSON(data []byte) error {
	type jsonModel Version

	var object struct {
		Version jsonModel
	}

	if err := json.Unmarshal(data, &object); err != nil {
		return err
	}

	*v = Version(object.Version)

	return nil
}
