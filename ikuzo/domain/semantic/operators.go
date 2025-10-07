package semantic

import "fmt"

// Operator represents a filter operator type.
type Operator string

// Filter operators supported by the semantic API.
const (
	// OpEqual performs exact match comparison
	OpEqual Operator = "eq"

	// OpNotEqual performs negation comparison
	OpNotEqual Operator = "neq"

	// OpIn matches if value is in the provided list
	OpIn Operator = "in"

	// OpNotIn matches if value is not in the provided list
	OpNotIn Operator = "nin"

	// OpGreaterThan performs numeric/date greater than comparison
	OpGreaterThan Operator = "gt"

	// OpGreaterEqual performs numeric/date greater than or equal comparison
	OpGreaterEqual Operator = "gte"

	// OpLessThan performs numeric/date less than comparison
	OpLessThan Operator = "lt"

	// OpLessEqual performs numeric/date less than or equal comparison
	OpLessEqual Operator = "lte"

	// OpContains performs substring match
	OpContains Operator = "contains"

	// OpStartsWith performs prefix match
	OpStartsWith Operator = "startswith"

	// OpExists checks if field exists (has any value)
	OpExists Operator = "exists"

	// Geospatial operators
	// OpBBox performs bounding box geographic query
	OpBBox Operator = "bbox"

	// OpWithin performs distance-based geographic query
	OpWithin Operator = "within"

	// OpPolygon performs polygon-based geographic query
	OpPolygon Operator = "polygon"

	// OpIntersects checks if geometries intersect
	OpIntersects Operator = "intersects"
)

// AllOperators returns all supported operators.
func AllOperators() []Operator {
	return []Operator{
		OpEqual,
		OpNotEqual,
		OpIn,
		OpNotIn,
		OpGreaterThan,
		OpGreaterEqual,
		OpLessThan,
		OpLessEqual,
		OpContains,
		OpStartsWith,
		OpExists,
		OpBBox,
		OpWithin,
		OpPolygon,
		OpIntersects,
	}
}

// String returns the string representation of the operator.
func (op Operator) String() string {
	return string(op)
}

// IsValid checks if the operator is valid.
func (op Operator) IsValid() bool {
	for _, valid := range AllOperators() {
		if op == valid {
			return true
		}
	}
	return false
}

// IsComparison returns true if this is a comparison operator (eq, gt, etc).
func (op Operator) IsComparison() bool {
	switch op {
	case OpEqual, OpNotEqual, OpGreaterThan, OpGreaterEqual, OpLessThan, OpLessEqual:
		return true
	default:
		return false
	}
}

// IsTextSearch returns true if this is a text search operator.
func (op Operator) IsTextSearch() bool {
	switch op {
	case OpContains, OpStartsWith:
		return true
	default:
		return false
	}
}

// IsListOperator returns true if this operator works with lists.
func (op Operator) IsListOperator() bool {
	switch op {
	case OpIn, OpNotIn:
		return true
	default:
		return false
	}
}

// IsGeospatial returns true if this is a geospatial operator.
func (op Operator) IsGeospatial() bool {
	switch op {
	case OpBBox, OpWithin, OpPolygon, OpIntersects:
		return true
	default:
		return false
	}
}

// RequiresValue returns false only for operators that don't need a value (like exists).
func (op Operator) RequiresValue() bool {
	return op != OpExists
}

// Validate validates the operator against a value.
func (op Operator) Validate(value interface{}) error {
	if !op.IsValid() {
		return fmt.Errorf("invalid operator: %s", op)
	}

	if op.RequiresValue() && value == nil {
		return fmt.Errorf("operator %s requires a value", op)
	}

	if !op.RequiresValue() && value != nil {
		return fmt.Errorf("operator %s does not accept a value", op)
	}

	return nil
}

// OperatorDescription returns a human-readable description of the operator.
func (op Operator) Description() string {
	descriptions := map[Operator]string{
		OpEqual:        "Exact match",
		OpNotEqual:     "Not equal to",
		OpIn:           "Matches any value in list",
		OpNotIn:        "Does not match any value in list",
		OpGreaterThan:  "Greater than",
		OpGreaterEqual: "Greater than or equal to",
		OpLessThan:     "Less than",
		OpLessEqual:    "Less than or equal to",
		OpContains:     "Contains substring",
		OpStartsWith:   "Starts with prefix",
		OpExists:       "Field has any value",
		OpBBox:         "Within bounding box",
		OpWithin:       "Within distance of point",
		OpPolygon:      "Within polygon",
		OpIntersects:   "Intersects with geometry",
	}

	if desc, ok := descriptions[op]; ok {
		return desc
	}

	return "Unknown operator"
}

// HydraMapping returns a Hydra-compatible description for API documentation.
func (op Operator) HydraMapping() map[string]interface{} {
	return map[string]interface{}{
		"@type":              "hub3:FilterOperator",
		"hub3:operator":      op.String(),
		"schema:name":        op.String(),
		"schema:description": op.Description(),
		"hub3:requiresValue": op.RequiresValue(),
		"hub3:isComparison":  op.IsComparison(),
		"hub3:isTextSearch":  op.IsTextSearch(),
		"hub3:isGeospatial":  op.IsGeospatial(),
	}
}
