package jsonata

import "testing"

const (
	testFeatureDotNavigation         = true
	testFeatureArrayIndexNavigation  = true
	testFeatureEqualityFilter        = true
	testFeatureFunctionCalls         = false
	testFeatureFieldsGroup           = false
	testFeatureArrayConstructorGroup = false
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

func skipIf(t *testing.T, feature bool, name string) {
	if !feature {
		t.Skip("feature: " + name + " not ready for implementation")
	}
}
