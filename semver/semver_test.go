package semver_test

import (
	"errors"
	"testing"

	"github.com/dmitrymomot/gokit/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Run("valid version without v prefix", func(t *testing.T) {
		v, err := semver.Parse("1.2.3")
		require.NoError(t, err)
		assert.Equal(t, uint64(1), v.Major)
		assert.Equal(t, uint64(2), v.Minor)
		assert.Equal(t, uint64(3), v.Patch)
		assert.Empty(t, v.Prerelease)
		assert.Empty(t, v.Build)
	})

	t.Run("valid version with v prefix", func(t *testing.T) {
		v, err := semver.Parse("v1.2.3")
		require.NoError(t, err)
		assert.Equal(t, uint64(1), v.Major)
		assert.Equal(t, uint64(2), v.Minor)
		assert.Equal(t, uint64(3), v.Patch)
	})

	t.Run("version with prerelease", func(t *testing.T) {
		v, err := semver.Parse("1.2.3-alpha.1")
		require.NoError(t, err)
		assert.Equal(t, uint64(1), v.Major)
		assert.Equal(t, uint64(2), v.Minor)
		assert.Equal(t, uint64(3), v.Patch)
		assert.Equal(t, "alpha.1", v.Prerelease)
		assert.Empty(t, v.Build)
	})

	t.Run("version with build metadata", func(t *testing.T) {
		v, err := semver.Parse("1.2.3+20130313144700")
		require.NoError(t, err)
		assert.Equal(t, uint64(1), v.Major)
		assert.Equal(t, uint64(2), v.Minor)
		assert.Equal(t, uint64(3), v.Patch)
		assert.Empty(t, v.Prerelease)
		assert.Equal(t, "20130313144700", v.Build)
	})

	t.Run("version with prerelease and build metadata", func(t *testing.T) {
		v, err := semver.Parse("1.2.3-alpha.1+20130313144700")
		require.NoError(t, err)
		assert.Equal(t, uint64(1), v.Major)
		assert.Equal(t, uint64(2), v.Minor)
		assert.Equal(t, uint64(3), v.Patch)
		assert.Equal(t, "alpha.1", v.Prerelease)
		assert.Equal(t, "20130313144700", v.Build)
	})

	t.Run("empty version", func(t *testing.T) {
		_, err := semver.Parse("")
		require.Error(t, err)
		assert.ErrorIs(t, err, semver.ErrEmptyVersion)
	})

	t.Run("invalid version format", func(t *testing.T) {
		_, err := semver.Parse("1.2")
		require.Error(t, err)
		assert.True(t, errors.Is(err, semver.ErrInvalidVersion))
	})

	t.Run("non-numeric version parts", func(t *testing.T) {
		_, err := semver.Parse("a.b.c")
		require.Error(t, err)
		assert.True(t, errors.Is(err, semver.ErrInvalidVersion))
	})
}

func TestMustParse(t *testing.T) {
	t.Run("valid version", func(t *testing.T) {
		v := semver.MustParse("1.2.3")
		assert.Equal(t, uint64(1), v.Major)
		assert.Equal(t, uint64(2), v.Minor)
		assert.Equal(t, uint64(3), v.Patch)
	})

	t.Run("invalid version panics", func(t *testing.T) {
		assert.Panics(t, func() {
			semver.MustParse("invalid")
		})
	})
}

func TestValidate(t *testing.T) {
	t.Run("valid version", func(t *testing.T) {
		assert.True(t, semver.Validate("1.2.3"))
		assert.True(t, semver.Validate("v1.2.3"))
		assert.True(t, semver.Validate("1.2.3-alpha"))
		assert.True(t, semver.Validate("1.2.3+build"))
		assert.True(t, semver.Validate("1.2.3-alpha+build"))
	})

	t.Run("invalid version", func(t *testing.T) {
		assert.False(t, semver.Validate(""))
		assert.False(t, semver.Validate("1.2"))
		assert.False(t, semver.Validate("1.2.3.4"))
		assert.False(t, semver.Validate("1.2.a"))
		assert.False(t, semver.Validate("a.b.c"))
	})
}

func TestString(t *testing.T) {
	t.Run("version without prerelease or build", func(t *testing.T) {
		v := semver.Version{Major: 1, Minor: 2, Patch: 3}
		assert.Equal(t, "1.2.3", v.String())
	})

	t.Run("version with prerelease", func(t *testing.T) {
		v := semver.Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha.1"}
		assert.Equal(t, "1.2.3-alpha.1", v.String())
	})

	t.Run("version with build", func(t *testing.T) {
		v := semver.Version{Major: 1, Minor: 2, Patch: 3, Build: "20130313144700"}
		assert.Equal(t, "1.2.3+20130313144700", v.String())
	})

	t.Run("version with prerelease and build", func(t *testing.T) {
		v := semver.Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha.1", Build: "20130313144700"}
		assert.Equal(t, "1.2.3-alpha.1+20130313144700", v.String())
	})
}

func TestCompare(t *testing.T) {
	t.Run("equal versions", func(t *testing.T) {
		v1 := semver.Version{Major: 1, Minor: 2, Patch: 3}
		v2 := semver.Version{Major: 1, Minor: 2, Patch: 3}
		assert.Equal(t, 0, v1.Compare(v2))
	})

	t.Run("major version difference", func(t *testing.T) {
		v1 := semver.Version{Major: 1, Minor: 2, Patch: 3}
		v2 := semver.Version{Major: 2, Minor: 2, Patch: 3}
		assert.Equal(t, -1, v1.Compare(v2))
		assert.Equal(t, 1, v2.Compare(v1))
	})

	t.Run("minor version difference", func(t *testing.T) {
		v1 := semver.Version{Major: 1, Minor: 2, Patch: 3}
		v2 := semver.Version{Major: 1, Minor: 3, Patch: 3}
		assert.Equal(t, -1, v1.Compare(v2))
		assert.Equal(t, 1, v2.Compare(v1))
	})

	t.Run("patch version difference", func(t *testing.T) {
		v1 := semver.Version{Major: 1, Minor: 2, Patch: 3}
		v2 := semver.Version{Major: 1, Minor: 2, Patch: 4}
		assert.Equal(t, -1, v1.Compare(v2))
		assert.Equal(t, 1, v2.Compare(v1))
	})

	t.Run("prerelease precedence", func(t *testing.T) {
		v1 := semver.Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha"}
		v2 := semver.Version{Major: 1, Minor: 2, Patch: 3}
		assert.Equal(t, -1, v1.Compare(v2))
		assert.Equal(t, 1, v2.Compare(v1))
	})

	t.Run("prerelease ordering", func(t *testing.T) {
		v1 := semver.Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha"}
		v2 := semver.Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "beta"}
		assert.Equal(t, -1, v1.Compare(v2))
		assert.Equal(t, 1, v2.Compare(v1))
	})

	t.Run("numeric prerelease identifiers", func(t *testing.T) {
		v1 := semver.Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha.1"}
		v2 := semver.Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha.2"}
		assert.Equal(t, -1, v1.Compare(v2))
		assert.Equal(t, 1, v2.Compare(v1))
	})

	t.Run("build metadata is ignored", func(t *testing.T) {
		v1 := semver.Version{Major: 1, Minor: 2, Patch: 3, Build: "1"}
		v2 := semver.Version{Major: 1, Minor: 2, Patch: 3, Build: "2"}
		assert.Equal(t, 0, v1.Compare(v2))
	})
}

func TestComparators(t *testing.T) {
	v1 := semver.Version{Major: 1, Minor: 2, Patch: 3}
	v2 := semver.Version{Major: 2, Minor: 0, Patch: 0}

	t.Run("LessThan", func(t *testing.T) {
		assert.True(t, v1.LessThan(v2))
		assert.False(t, v2.LessThan(v1))
		assert.False(t, v1.LessThan(v1))
	})

	t.Run("GreaterThan", func(t *testing.T) {
		assert.False(t, v1.GreaterThan(v2))
		assert.True(t, v2.GreaterThan(v1))
		assert.False(t, v1.GreaterThan(v1))
	})

	t.Run("Equal", func(t *testing.T) {
		assert.False(t, v1.Equal(v2))
		assert.True(t, v1.Equal(v1))
	})

	t.Run("LessThanOrEqual", func(t *testing.T) {
		assert.True(t, v1.LessThanOrEqual(v2))
		assert.False(t, v2.LessThanOrEqual(v1))
		assert.True(t, v1.LessThanOrEqual(v1))
	})

	t.Run("GreaterThanOrEqual", func(t *testing.T) {
		assert.False(t, v1.GreaterThanOrEqual(v2))
		assert.True(t, v2.GreaterThanOrEqual(v1))
		assert.True(t, v1.GreaterThanOrEqual(v1))
	})
}

func TestInRange(t *testing.T) {
	v1 := semver.MustParse("1.0.0")
	v2 := semver.MustParse("2.0.0")
	v3 := semver.MustParse("3.0.0")

	t.Run("version in range", func(t *testing.T) {
		assert.True(t, v2.InRange(v1, v3))
	})

	t.Run("version at lower bound", func(t *testing.T) {
		assert.True(t, v1.InRange(v1, v3))
	})

	t.Run("version at upper bound", func(t *testing.T) {
		assert.True(t, v3.InRange(v1, v3))
	})

	t.Run("version below range", func(t *testing.T) {
		below := semver.MustParse("0.9.0")
		assert.False(t, below.InRange(v1, v3))
	})

	t.Run("version above range", func(t *testing.T) {
		above := semver.MustParse("3.1.0")
		assert.False(t, above.InRange(v1, v3))
	})
}

func TestVersionModifiers(t *testing.T) {
	base := semver.Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha", Build: "001"}

	t.Run("WithMajor", func(t *testing.T) {
		modified := base.WithMajor(5)
		assert.Equal(t, uint64(5), modified.Major)
		assert.Equal(t, base.Minor, modified.Minor)
		assert.Equal(t, base.Patch, modified.Patch)
		assert.Equal(t, base.Prerelease, modified.Prerelease)
		assert.Equal(t, base.Build, modified.Build)
	})

	t.Run("WithMinor", func(t *testing.T) {
		modified := base.WithMinor(5)
		assert.Equal(t, base.Major, modified.Major)
		assert.Equal(t, uint64(5), modified.Minor)
		assert.Equal(t, base.Patch, modified.Patch)
		assert.Equal(t, base.Prerelease, modified.Prerelease)
		assert.Equal(t, base.Build, modified.Build)
	})

	t.Run("WithPatch", func(t *testing.T) {
		modified := base.WithPatch(5)
		assert.Equal(t, base.Major, modified.Major)
		assert.Equal(t, base.Minor, modified.Minor)
		assert.Equal(t, uint64(5), modified.Patch)
		assert.Equal(t, base.Prerelease, modified.Prerelease)
		assert.Equal(t, base.Build, modified.Build)
	})

	t.Run("WithPrerelease", func(t *testing.T) {
		modified := base.WithPrerelease("beta")
		assert.Equal(t, base.Major, modified.Major)
		assert.Equal(t, base.Minor, modified.Minor)
		assert.Equal(t, base.Patch, modified.Patch)
		assert.Equal(t, "beta", modified.Prerelease)
		assert.Equal(t, base.Build, modified.Build)
	})

	t.Run("WithBuild", func(t *testing.T) {
		modified := base.WithBuild("002")
		assert.Equal(t, base.Major, modified.Major)
		assert.Equal(t, base.Minor, modified.Minor)
		assert.Equal(t, base.Patch, modified.Patch)
		assert.Equal(t, base.Prerelease, modified.Prerelease)
		assert.Equal(t, "002", modified.Build)
	})
}

func TestIncrement(t *testing.T) {
	base := semver.Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha", Build: "001"}

	t.Run("increment major", func(t *testing.T) {
		modified, err := base.Increment("major")
		require.NoError(t, err)
		assert.Equal(t, uint64(2), modified.Major)
		assert.Equal(t, uint64(0), modified.Minor)
		assert.Equal(t, uint64(0), modified.Patch)
		assert.Empty(t, modified.Prerelease)
		assert.Empty(t, modified.Build)
	})

	t.Run("increment minor", func(t *testing.T) {
		modified, err := base.Increment("minor")
		require.NoError(t, err)
		assert.Equal(t, uint64(1), modified.Major)
		assert.Equal(t, uint64(3), modified.Minor)
		assert.Equal(t, uint64(0), modified.Patch)
		assert.Empty(t, modified.Prerelease)
		assert.Empty(t, modified.Build)
	})

	t.Run("increment patch", func(t *testing.T) {
		modified, err := base.Increment("patch")
		require.NoError(t, err)
		assert.Equal(t, uint64(1), modified.Major)
		assert.Equal(t, uint64(2), modified.Minor)
		assert.Equal(t, uint64(4), modified.Patch)
		assert.Empty(t, modified.Prerelease)
		assert.Empty(t, modified.Build)
	})

	t.Run("invalid component", func(t *testing.T) {
		_, err := base.Increment("invalid")
		require.Error(t, err)
	})
}

func TestMajorMinorPatch(t *testing.T) {
	t.Run("with prerelease and build", func(t *testing.T) {
		v := semver.Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha", Build: "001"}
		assert.Equal(t, "1.2.3", v.MajorMinorPatch())
	})

	t.Run("without prerelease and build", func(t *testing.T) {
		v := semver.Version{Major: 1, Minor: 2, Patch: 3}
		assert.Equal(t, "1.2.3", v.MajorMinorPatch())
	})
}

func TestIsValid(t *testing.T) {
	t.Run("valid version", func(t *testing.T) {
		v := semver.Version{Major: 1, Minor: 2, Patch: 3}
		assert.True(t, v.IsValid())
	})

	t.Run("valid version with prerelease", func(t *testing.T) {
		v := semver.Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha.1"}
		assert.True(t, v.IsValid())
	})

	t.Run("valid version with build", func(t *testing.T) {
		v := semver.Version{Major: 1, Minor: 2, Patch: 3, Build: "20130313144700"}
		assert.True(t, v.IsValid())
	})
}
