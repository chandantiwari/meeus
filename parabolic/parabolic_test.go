package parabolic_test

import (
	"fmt"
	"math"

	"github.com/soniakeys/meeus/julian"
	"github.com/soniakeys/meeus/parabolic"
)

func ExampleElements_AnomalyDistance() {
	// Example 34.a, p. 243
	e := &parabolic.Elements{
		TimeP: julian.CalendarGregorianToJD(1998, 4, 14.4358),
		PDis:  1.487469,
	}
	j := julian.CalendarGregorianToJD(1998, 8, 5)
	ν, r := e.AnomalyDistance(j)
	fmt.Printf("%.5f deg\n", ν*180/math.Pi)
	fmt.Printf("%.6f AU\n", r)
	// Output:
	// 66.78862 deg
	// 2.133911 AU
}
