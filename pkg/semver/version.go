package semver

import (
	"strconv"
	"strings"
)

type Version struct {
	Major, Minor, Patch uint
	PreRelease          []string
	Build               []string
}

func (v Version) String() string {
	return string(v.bytes())
}

func (v Version) BumpMajor() Version {
	v.Major++
	v.Minor = 0
	v.Patch = 0

	v.PreRelease = nil
	v.Build = nil

	return v
}

func (v Version) BumpMinor() Version {
	v.Minor++
	v.Patch = 0

	v.PreRelease = nil
	v.Build = nil

	return v
}

func (v Version) BumpPatch() Version {
	v.Patch++

	v.PreRelease = nil
	v.Build = nil

	return v
}

func (v Version) IsPreRelease() bool {
	return len(v.PreRelease) > 0
}

func (v Version) ResetPreRelease() Version {
	v.PreRelease = nil

	return v
}

func (v Version) SetPreRelease(p []string) Version {
	v.PreRelease = p

	return v
}

func (v Version) HasBuild() bool {
	return len(v.Build) > 0
}

func (v Version) ResetBuild() Version {
	v.Build = nil

	return v
}

func (v Version) SetBuild(b []string) Version {
	v.Build = b

	return v
}

func (v Version) Compare(other Version) int {
	// major
	{
		if v.Major > other.Major {
			return 1
		}

		if v.Major < other.Major {
			return -1
		}
	}

	// minor
	{
		if v.Minor > other.Minor {
			return 1
		}

		if v.Minor < other.Minor {
			return -1
		}
	}

	// patch
	{
		if v.Patch > other.Patch {
			return 1
		}

		if v.Patch < other.Patch {
			return -1
		}
	}

	if len(v.PreRelease) == 0 && len(other.PreRelease) == 0 {
		return 0
	}

	// at this point major.minor.patch are equal so we compare pre-release

	// When major, minor, and patch are equal, a pre-release version has lower precedence than a normal version
	if len(v.PreRelease) == 0 && len(other.PreRelease) > 0 {
		return 1
	}

	if len(v.PreRelease) > 0 && len(other.PreRelease) == 0 {
		return -1
	}

	return comparePreRelease(v.PreRelease, other.PreRelease)
}

func (v Version) IsZero() bool {
	return v.Major == 0 && v.Minor == 0 && v.Patch == 0 && len(v.PreRelease) == 0 && len(v.Build) == 0
}

func (v Version) bytes() []byte {
	buf := make([]byte, 0, 16)

	buf = strconv.AppendUint(buf, uint64(v.Major), 10)
	buf = append(buf, '.')
	buf = strconv.AppendUint(buf, uint64(v.Minor), 10)
	buf = append(buf, '.')
	buf = strconv.AppendUint(buf, uint64(v.Patch), 10)

	if len(v.PreRelease) > 0 {
		buf = append(buf, '-')
		for i, pre := range v.PreRelease {
			if i > 0 {
				buf = append(buf, '.')
			}
			buf = append(buf, pre...)
		}
	}

	if len(v.Build) > 0 {
		buf = append(buf, '+')
		for i, build := range v.Build {
			if i > 0 {
				buf = append(buf, '.')
			}
			buf = append(buf, build...)
		}
	}

	return buf
}

func comparePreRelease(a, b []string) int {
	// Precedence for two pre-release versions with the same major, minor, and patch version MUST be determined
	// by comparing each dot separated identifier from left to right until a difference is found as follows:
	//	1. Identifiers consisting of only digits are compared numerically.
	//	2. Identifiers with letters or hyphens are compared lexically in ASCII sort order.
	//	3. Numeric identifiers always have lower precedence than non-numeric identifiers.
	//	4. A larger set of pre-release fields has a higher precedence than a smaller set, if all of the preceding identifiers are equal.
	//
	// Example: 1.0.0-alpha < 1.0.0-alpha.1 < 1.0.0-alpha.beta < 1.0.0-beta < 1.0.0-beta.2 < 1.0.0-beta.11 < 1.0.0-rc.1 < 1.0.0.

	var i int

	for i = 0; i < len(a) && i < len(b); i++ {
		cmp := comparePreReleaseIdentifier(a[i], b[i])
		if cmp != 0 {
			return cmp
		}
	}

	// 4. A larger set of pre-release fields has a higher precedence than a smaller set, if all of the preceding identifiers are equal.
	{
		if len(a) > len(b) {
			return 1
		}

		if len(a) < len(b) {
			return -1
		}
	}

	return 0
}

func comparePreReleaseIdentifier(a, b string) int {
	var (
		isANumeric, isBNumeric bool
		aNumeric, bNumeric     uint64
	)

	if len(a) > 0 && isPositiveDigit(a[0]) || (isDigit(a[0]) && len(a) == 1) {
		var err error
		aNumeric, err = strconv.ParseUint(a, 10, 64)
		isANumeric = err == nil
	}

	if len(b) > 0 && isPositiveDigit(b[0]) || (isDigit(a[0]) && len(a) == 1) {
		var err error
		bNumeric, err = strconv.ParseUint(b, 10, 64)
		isBNumeric = err == nil
	}

	//1. Identifiers consisting of only digits are compared numerically.
	if isANumeric && isBNumeric {
		if aNumeric < bNumeric {
			return -1
		}

		if aNumeric > bNumeric {
			return 1
		}

		return 0
	}

	// 3. Numeric identifiers always have lower precedence than non-numeric identifiers.
	{
		if isANumeric {
			return -1
		}

		if isBNumeric {
			return 1
		}
	}

	// 2. Identifiers with letters or hyphens are compared lexically in ASCII sort order.
	return strings.Compare(a, b)
}

func isPositiveDigit(a byte) bool {
	return a > '0' && a <= '9'
}
