// Copyright 2013 Sonia Keys
// License MIT: http://www.opensource.org/licenses/MIT

// Solar: Chapter 25, Solar Coordinates.
//
// Partial implementation:
//
// 1. Higher accuracy positions are not computed with Appendix III but with
// full VSOP87 as implemented in package planetposition.
//
// 2. Higher accuracy correction for aberration (using the formula for
// variation Δλ on p. 168) is not implemented.  Results for example 25.b
// already match the full VSOP87 values on p. 165 even with the low accuracy
// correction for aberration, thus there are no more significant digits that
// would check a more accurate result.  Also the size of the formula presents
// significant chance of typographical error.
package solar

import (
	"math"

	"github.com/soniakeys/meeus/base"
	"github.com/soniakeys/meeus/coord"
	"github.com/soniakeys/meeus/nutation"
	pp "github.com/soniakeys/meeus/planetposition"
)

// True returns true geometric longitude and anomaly of the sun referenced to the mean equinox of date.
//
// Argument T is the number of Julian centuries since J2000.
// See base.J2000Century.
//
// Results:
//	s = true geometric longitude, ☉, in radians
//	ν = true anomaly in radians
func True(T float64) (s, ν float64) {
	// (25.2) p. 163
	L0 := base.Horner(T, 280.46646, 36000.76983, 0.0003032) *
		math.Pi / 180
	M := MeanAnomaly(T)
	C := (base.Horner(T, 1.914602, -0.004817, -.000014)*
		math.Sin(M) +
		(0.019993-.000101*T)*math.Sin(2*M) +
		0.000289*math.Sin(3*M)) * math.Pi / 180
	return base.PMod(L0+C, 2*math.Pi), base.PMod(M+C, 2*math.Pi)
}

// MeanAnomaly returns the mean anomaly of Earth at the given T.
//
// Argument T is the number of Julian centuries since J2000.
// See base.J2000Century.
//
// Result is in radians and is not normalized to the range 0..2π.
func MeanAnomaly(T float64) float64 {
	// (25.3) p. 163
	return base.Horner(T, 357.52911, 35999.05029, -0.0001537) * math.Pi / 180
}

// Eccentricity returns eccentricity of the Earth's orbit around the sun.
//
// Argument T is the number of Julian centuries since J2000.
// See base.J2000Century.
func Eccentricity(T float64) float64 {
	// (25.4) p. 163
	return base.Horner(T, 0.016708634, -0.000042037, -0.0000001267)
}

// Radius returns the Sun-Earth distance in AU.
//
// Argument T is the number of Julian centuries since J2000.
// See base.J2000Century.
func Radius(T float64) float64 {
	_, ν := True(T)
	e := Eccentricity(T)
	// (25.5) p. 164
	return 1.000001018 * (1 - e*e) / (1 + e*math.Cos(ν))
}

// ApparentLongitude returns apparent longitude of the Sun referenced to the true equinox of date.
//
// Argument T is the number of Julian centuries since J2000.
// See base.J2000Century.
//
// Result includes correction for nutation and aberration.  Unit is radians.
func ApparentLongitude(T float64) float64 {
	Ω := node(T)
	s, _ := True(T)
	return s - .00569*math.Pi/180 - .00478*math.Pi/180*math.Sin(Ω)
}

func node(T float64) float64 {
	return 125.04*math.Pi/180 - 1934.136*math.Pi/180*T
}

// True2000 returns true geometric longitude and anomaly of the sun referenced to equinox J2000.
//
// Argument T is the number of Julian centuries since J2000.
// See base.J2000Century.
//
// Results are accurate to .01 degree for years 1900 to 2100.
//
// Results:
//	s = true geometric longitude, ☉, in radians
//	ν = true anomaly in radians
func True2000(T float64) (s, ν float64) {
	s, ν = True(T)
	s -= .01397 * math.Pi / 180 * T * 100
	return
}

// TrueEquatorial returns the true geometric position of the Sun as equatorial coordinates.
func TrueEquatorial(jde float64) (α, δ float64) {
	s, _ := True(base.J2000Century(jde))
	ε := nutation.MeanObliquity(jde)
	ss, cs := math.Sincos(s)
	sε, cε := math.Sincos(ε)
	// (25.6, 25.7) p. 165
	return math.Atan2(cε*ss, cs), sε * ss
}

// ApparentEquatorial returns the apparent position of the Sun as equatorial coordinates.
//
//	α: right ascension in radians
//	δ: declination in radians
func ApparentEquatorial(jde float64) (α, δ float64) {
	T := base.J2000Century(jde)
	λ := ApparentLongitude(T)
	ε := nutation.MeanObliquity(jde)
	sλ, cλ := math.Sincos(λ)
	// (25.8) p. 165
	sε, cε := math.Sincos(ε + .00256*math.Pi/180*math.Cos(node(T)))
	return math.Atan2(cε*sλ, cλ), math.Asin(sε * sλ)
}

// TrueVSOP87 returns the true geometric position of the sun as ecliptic coordinates.
//
// Result computed by full VSOP87 theory.  Result is at equator and equinox
// of date in the FK5 frame.  It does not include nutation or aberration.
//
//	s: ecliptic longitude in radians
//	β: ecliptic latitude in radians
//	R: range in AU
func TrueVSOP87(e *pp.V87Planet, jde float64) (s, β, R float64) {
	l, b, r := e.Position(jde)
	s = l + math.Pi
	// FK5 correction.
	λp := base.Horner(base.J2000Century(jde),
		s, -1.397*math.Pi/180, -.00031*math.Pi/180)
	sλp, cλp := math.Sincos(λp)
	Δβ := .03916 / 3600 * math.Pi / 180 * (cλp - sλp)
	// (25.9) p. 166
	return base.PMod(s-.09033/3600*math.Pi/180, 2*math.Pi), Δβ - b, r
}

// ApparentVSOP87 returns the apparent position of the sun as ecliptic coordinates.
//
// Result computed by VSOP87, at equator and equinox of date in the FK5 frame,
// and includes effects of nutation and aberration.
//
//  λ: ecliptic longitude in radians
//  β: ecliptic latitude in radians
//  R: range in AU
func ApparentVSOP87(e *pp.V87Planet, jde float64) (λ, β, R float64) {
	// note: see duplicated code in ApparentEquatorialVSOP87.
	s, β, R := TrueVSOP87(e, jde)
	Δψ, _ := nutation.Nutation(jde)
	a := aberration(R)
	return s + Δψ + a, β, R
}

// ApparentEquatorialVSOP87 returns the apparent position of the sun as equatorial coordinates.
//
// Result computed by VSOP87, at equator and equinox of date in the FK5 frame,
// and includes effects of nutation and aberration.
//
//	α: right ascension in radians
//	δ: declination in radians
//	R: range in AU
func ApparentEquatorialVSOP87(e *pp.V87Planet, jde float64) (α, δ, R float64) {
	// note: duplicate code from ApparentVSOP87 so we can keep Δε.
	// see also duplicate code in time.E().
	s, β, R := TrueVSOP87(e, jde)
	Δψ, Δε := nutation.Nutation(jde)
	a := aberration(R)
	λ := s + Δψ + a
	ε := nutation.MeanObliquity(jde) + Δε
	sε, cε := math.Sincos(ε)
	α, δ = coord.EclToEq(λ, β, sε, cε)
	return
}

// Low precision formula.  The high precision formula is not implemented
// because the low precision formula already gives position results to the
// accuracy given on p. 165.  The high precision formula the represents lots
// of typing with associated chance of typos, and no way to test the result.
func aberration(R float64) float64 {
	// (25.10) p. 167
	return -20.4898 / 3600 * math.Pi / 180 / R
}
