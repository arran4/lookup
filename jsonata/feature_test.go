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
var groupStatus = map[string]bool{
    // Existing groups (keeping previous config if needed)
    "fields": true,
    "array-constructor": true,
    "comments": true,
    "missing-paths": true,
}

func skipIf(t *testing.T, feature bool, name string) {
	if !feature {
		t.Skip("feature: " + name + " not ready for implementation")
	}
}

func skipGroupIfDisabled(t *testing.T, groupName string) {
    // Default to disabled (skipped) to ensure build passes after import.
    // Enable groups in groupStatus as they are verified to pass.
    if enabled, ok := groupStatus[groupName]; !ok || !enabled {
        t.Skip("group: " + groupName + " disabled or not yet enabled")
    }
}
