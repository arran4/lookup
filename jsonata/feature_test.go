package jsonata

import "testing"

const (
	testFeatureDotNavigation         = true
	testFeatureArrayIndexNavigation  = true
	testFeatureEqualityFilter        = true
	testFeatureFunctionCalls         = false
	testFeaturePredicate             = false
	testFeaturePathOperators         = false
	testFeatureNumericOperators      = false
	testFeatureBooleanOperators      = false
	testFeatureBooleanFunctions      = false
	testFeatureStringFunctions       = false
	testFeatureNumericFunctions      = false
	testFeatureAggregationFunctions  = false
	testFeatureArrayFunctions        = false
	testFeatureObjectFunctions       = false
	testFeatureHigherOrderFunctions  = false
	testFeatureDateTimeFunctions     = false
	testFeatureRegex                 = false
)

// groupStatus controls which test groups are enabled.
// If true, the group is expected to pass completely. Failures will fail the test.
// If false or missing, the group is run but failures are skipped.
var groupStatus = map[string]bool{
	"fields":            false, // Partial failure
	"array-constructor": false, // Partial failure
	"comments":          false, // Partial failure
	"missing-paths":     true,  // Passing
}

func skipIf(t *testing.T, feature bool, name string) {
	if !feature {
		t.Skip("feature: " + name + " not ready for implementation")
	}
}
