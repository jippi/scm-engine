package gitlab

// Extending the EvalContext with Expr-lang "valuer" interface support
// allowing custom types to act as native Expr-lang types
//
// See: https://github.com/expr-lang/expr/blob/master/patcher/value/value.go
// See: pkg/stdlib/stdlib.go

// MergeRequestState is a ENUM type
func (d MergeRequestState) AsString() string {
	return d.String()
}

// UserState is a ENUM type
func (d UserState) AsString() string {
	return d.String()
}

// MergeStatus is a ENUM type
func (d MergeStatus) AsString() string {
	return d.String()
}

// DetailedMergeStatus is a ENUM type
func (d DetailedMergeStatus) AsString() string {
	return d.String()
}

// PipelineStatusEnum is a ENUM type
func (d PipelineStatusEnum) AsString() string {
	return d.String()
}
