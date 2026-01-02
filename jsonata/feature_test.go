package jsonata

import "testing"

const (
	testFeatureDotNavigation        = true
	testFeatureArrayIndexNavigation = true
	testFeatureEqualityFilter       = true
	testFeatureFunctionCalls        = false
	testFeaturePredicate            = false
	testFeaturePathOperators        = false
	testFeatureNumericOperators     = false
	testFeatureBooleanOperators     = false
	testFeatureBooleanFunctions     = false
	testFeatureStringFunctions      = false
	testFeatureNumericFunctions     = false
	testFeatureAggregationFunctions = false
	testFeatureArrayFunctions       = false
	testFeatureObjectFunctions      = false
	testFeatureHigherOrderFunctions = false
	testFeatureDateTimeFunctions    = false
	testFeatureRegex                = false
)

// groupStatus controls which test groups are enabled.
// If a group is present in this map, it runs in "skip on failure" mode.
// If a group is missing, it runs in "strict" mode (failures break the build).
// Eventually, this map should be empty as all groups are fixed.
var groupStatus = map[string]bool{
	"array-constructor":           true,
	"blocks":                      true,
	"boolean-expresssions":        true,
	"closures":                    true,
	"coalescing-operator":         true,
	"comments":                    true,
	"comparison-operators":        true,
	"conditionals":                true,
	"context":                     true,
	"default-operator":            true,
	"descendent-operator":         true,
	"encoding":                    true,
	"errors":                      true,
	"fields":                      true,
	"flattening":                  true,
	"function-abs":                true,
	"function-append":             true,
	"function-applications":       true,
	"function-assert":             true,
	"function-average":            true,
	"function-boolean":            true,
	"function-ceil":               true,
	"function-contains":           true,
	"function-count":              true,
	"function-decodeUrl":          true,
	"function-decodeUrlComponent": true,
	"function-each":               true,
	"function-encodeUrl":          true,
	"function-encodeUrlComponent": true,
	"function-error":              true,
	"function-eval":               true,
	"function-exists":             true,
	"function-floor":              true,
	"function-formatBase":         true,
	"function-formatNumber":       true,
	"function-fromMillis":         true,
	"function-join":               true,
	"function-keys":               true,
	"function-length":             true,
	"function-lookup":             true,
	"function-lowercase":          true,
	"function-max":                true,
	"function-merge":              true,
	"function-number":             true,
	"function-pad":                true,
	"function-power":              true,
	"function-replace":            true,
	"function-reverse":            true,
	"function-round":              true,
	"function-shuffle":            true,
	"function-sift":               true,
	"function-signatures":         true,
	"function-sort":               true,
	"function-split":              true,
	"function-spread":             true,
	"function-sqrt":               true,
	"function-string":             true,
	"function-substring":          true,
	"function-substringAfter":     true,
	"function-substringBefore":    true,
	"function-sum":                true,
	"function-tomillis":           true,
	"function-trim":               true,
	"function-typeOf":             true,
	"function-uppercase":          true,
	"function-zip":                true,
	"higher-order-functions":      true,
	"hof-filter":                  true,
	"hof-map":                     true,
	"hof-reduce":                  true,
	"hof-single":                  true,
	"hof-zip-map":                 true,
	"inclusion-operator":          true,
	"lambdas":                     true,
	"literals":                    true,
	"matchers":                    true,
	"multiple-array-selectors":    true,
	"null":                        true,
	"numeric-operators":           true,
	"object-constructor":          true,
	"parentheses":                 true,
	"partial-application":         true,
	"performance":                 true,
	"predicates":                  true,
	"quoted-selectors":            true,
	"range-operator":              true,
	"regex":                       true,
	"simple-array-selectors":      true,
	"sorting":                     true,
	"string-concat":               true,
	"tail-recursion":              true,
	"token-conversion":            true,
	"transform":                   true,
	"transforms":                  true,
	"variables":                   true,
	"wildcards":                   true,
	// "missing-paths": true, // Not present means STRICT PASS
}

func skipIf(t *testing.T, feature bool, name string) {
	if !feature {
		t.Skip("feature: " + name + " not ready for implementation")
	}
}
