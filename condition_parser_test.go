package gqlparser

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestConstructConditionAST(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []Token
		want    conditionAST
		wantErr error
	}{
		// Basic value types
		{
			name:   "SymbolToken",
			tokens: []Token{&SymbolToken{Content: "field", Position: 0}},
			want:   &conditionField{sym: &SymbolToken{Content: "field", Position: 0}},
		},
		{
			name:   "BooleanToken_true",
			tokens: []Token{&BooleanToken{Value: true, RawContent: "true", Position: 0}},
			want:   &conditionValue{b: &BooleanToken{Value: true, RawContent: "true", Position: 0}},
		},
		{
			name:   "BooleanToken_false",
			tokens: []Token{&BooleanToken{Value: false, RawContent: "false", Position: 0}},
			want:   &conditionValue{b: &BooleanToken{Value: false, RawContent: "false", Position: 0}},
		},
		{
			name:   "StringToken_field",
			tokens: []Token{&StringToken{Quote: '`', Content: "field", RawContent: "`field`", Position: 0}},
			want:   &conditionField{str: &StringToken{Quote: '`', Content: "field", RawContent: "`field`", Position: 0}},
		},
		{
			name:   "StringToken_value",
			tokens: []Token{&StringToken{Quote: '"', Content: "value", RawContent: "\"value\"", Position: 0}},
			want:   &conditionValue{s: &StringToken{Quote: '"', Content: "value", RawContent: "\"value\"", Position: 0}},
		},
		{
			name:   "NumericToken_integer",
			tokens: []Token{&NumericToken{Int64: 42, Floating: false, RawContent: "42", Position: 0}},
			want:   &conditionValue{n: &NumericToken{Int64: 42, Floating: false, RawContent: "42", Position: 0}},
		},
		{
			name:   "NumericToken_float",
			tokens: []Token{&NumericToken{Float64: 3.14, Floating: true, RawContent: "3.14", Position: 0}},
			want:   &conditionValue{n: &NumericToken{Float64: 3.14, Floating: true, RawContent: "3.14", Position: 0}},
		},
		{
			name:   "BindingToken_named",
			tokens: []Token{&BindingToken{Name: "param", Position: 0}},
			want:   &conditionValue{bind: &BindingToken{Name: "param", Position: 0}},
		},
		{
			name:   "BindingToken_indexed",
			tokens: []Token{&BindingToken{Index: 1, Position: 0}},
			want:   &conditionValue{bind: &BindingToken{Index: 1, Position: 0}},
		},

		// field access
		{
			name:   "NestedSymbolToken",
			tokens: []Token{&SymbolToken{Content: "field", Position: 0}, &OperatorToken{Type: ".", Position: 5}, &SymbolToken{Content: "field", Position: 6}},
			want: &conditionFieldAccess{
				left:  &conditionField{sym: &SymbolToken{Content: "field", Position: 0}},
				right: &conditionField{sym: &SymbolToken{Content: "field", Position: 6}},
			},
		},
		{
			name:   "MoreNestedSymbolToken",
			tokens: []Token{&SymbolToken{Content: "field", Position: 0}, &OperatorToken{Type: ".", Position: 5}, &SymbolToken{Content: "field", Position: 6}, &OperatorToken{Type: ".", Position: 11}, &SymbolToken{Content: "field", Position: 12}},
			want: &conditionFieldAccess{
				left: &conditionFieldAccess{
					left:  &conditionField{sym: &SymbolToken{Content: "field", Position: 0}},
					right: &conditionField{sym: &SymbolToken{Content: "field", Position: 6}},
				},
				right: &conditionField{sym: &SymbolToken{Content: "field", Position: 12}},
			},
		},

		// Keyword types
		{
			name: "KeywordToken_KEY",
			tokens: []Token{
				&KeywordToken{Name: "KEY", RawContent: "KEY", Position: 0},
				&OperatorToken{Type: "(", RawContent: "(", Position: 3},
				&SymbolToken{Content: "EntityName", Position: 4},
				&OperatorToken{Type: ",", RawContent: ",", Position: 14},
				&NumericToken{Int64: 123, Floating: false, RawContent: "123", Position: 15},
				&OperatorToken{Type: ")", RawContent: ")", Position: 18},
			},
			want: &conditionKey{
				keyKeyword: &KeywordToken{Name: "KEY", RawContent: "KEY", Position: 0},
				key: &Key{
					Path: []*KeyPath{
						{
							Kind: "EntityName",
							ID:   int64(123),
						},
					},
				},
			},
		},
		{
			name:   "KeywordToken_NULL",
			tokens: []Token{&KeywordToken{Name: "NULL", RawContent: "NULL", Position: 0}},
			want:   &conditionValue{null: &KeywordToken{Name: "NULL", RawContent: "NULL", Position: 0}},
		},

		// Infix operators - Either operators (field = value)
		{
			name: "InfixOperator_equals_forward",
			tokens: []Token{
				&SymbolToken{Content: "field", Position: 0},
				&WhitespaceToken{Content: " ", Position: 5},
				&OperatorToken{Type: "=", RawContent: "=", Position: 6},
				&WhitespaceToken{Content: " ", Position: 7},
				&StringToken{Quote: '"', Content: "value", RawContent: "\"value\"", Position: 8},
			},
			want: &forwardComparatorCondition{
				left:   &conditionField{sym: &SymbolToken{Content: "field", Position: 0}},
				op:     &OperatorToken{Type: "=", RawContent: "=", Position: 6},
				opType: "=",
				right:  &conditionValue{s: &StringToken{Quote: '"', Content: "value", RawContent: "\"value\"", Position: 8}},
			},
		},
		{
			name: "InfixOperator_not_equals_backward",
			tokens: []Token{
				&StringToken{Quote: '"', Content: "value", RawContent: "\"value\"", Position: 0},
				&WhitespaceToken{Content: " ", Position: 7},
				&OperatorToken{Type: "!=", RawContent: "!=", Position: 8},
				&WhitespaceToken{Content: " ", Position: 10},
				&SymbolToken{Content: "field", Position: 11},
			},
			want: &backwardComparatorCondition{
				left:   &conditionValue{s: &StringToken{Quote: '"', Content: "value", RawContent: "\"value\"", Position: 0}},
				op:     &OperatorToken{Type: "!=", RawContent: "!=", Position: 8},
				opType: "!=",
				right:  &conditionField{sym: &SymbolToken{Content: "field", Position: 11}},
			},
		},

		// Compound operators
		{
			name: "InfixOperator_AND",
			tokens: []Token{
				&SymbolToken{Content: "field1", Position: 0},
				&WhitespaceToken{Content: " ", Position: 6},
				&OperatorToken{Type: "=", RawContent: "=", Position: 7},
				&WhitespaceToken{Content: " ", Position: 8},
				&StringToken{Quote: '"', Content: "value1", RawContent: "\"value1\"", Position: 9},
				&WhitespaceToken{Content: " ", Position: 17},
				&OperatorToken{Type: "AND", RawContent: "AND", Position: 18},
				&WhitespaceToken{Content: " ", Position: 21},
				&SymbolToken{Content: "field2", Position: 22},
				&WhitespaceToken{Content: " ", Position: 28},
				&OperatorToken{Type: "=", RawContent: "=", Position: 29},
				&WhitespaceToken{Content: " ", Position: 30},
				&StringToken{Quote: '"', Content: "value2", RawContent: "\"value2\"", Position: 31},
			},
			want: &compoundComparatorCondition{
				left: &forwardComparatorCondition{
					left:   &conditionField{sym: &SymbolToken{Content: "field1", Position: 0}},
					op:     &OperatorToken{Type: "=", RawContent: "=", Position: 7},
					opType: "=",
					right:  &conditionValue{s: &StringToken{Quote: '"', Content: "value1", RawContent: "\"value1\"", Position: 9}},
				},
				op: &OperatorToken{Type: "AND", RawContent: "AND", Position: 18},
				right: &forwardComparatorCondition{
					left:   &conditionField{sym: &SymbolToken{Content: "field2", Position: 22}},
					op:     &OperatorToken{Type: "=", RawContent: "=", Position: 29},
					opType: "=",
					right:  &conditionValue{s: &StringToken{Quote: '"', Content: "value2", RawContent: "\"value2\"", Position: 31}},
				},
			},
		},

		// Forward operators
		{
			name: "InfixOperator_CONTAINS",
			tokens: []Token{
				&SymbolToken{Content: "field", Position: 0},
				&WhitespaceToken{Content: " ", Position: 5},
				&OperatorToken{Type: "CONTAINS", RawContent: "CONTAINS", Position: 6},
				&WhitespaceToken{Content: " ", Position: 14},
				&StringToken{Quote: '"', Content: "substring", RawContent: "\"substring\"", Position: 15},
			},
			want: &forwardComparatorCondition{
				left:   &conditionField{sym: &SymbolToken{Content: "field", Position: 0}},
				op:     &OperatorToken{Type: "CONTAINS", RawContent: "CONTAINS", Position: 6},
				opType: "CONTAINS",
				right:  &conditionValue{s: &StringToken{Quote: '"', Content: "substring", RawContent: "\"substring\"", Position: 15}},
			},
		},
		{
			name: "ForwardOperator_IS_NULL",
			tokens: []Token{
				&SymbolToken{Content: "field", Position: 0},
				&WhitespaceToken{Content: " ", Position: 5},
				&OperatorToken{Type: "IS", RawContent: "IS", Position: 6},
				&WhitespaceToken{Content: " ", Position: 8},
				&KeywordToken{Name: "NULL", RawContent: "NULL", Position: 9},
			},
			want: &forwardComparatorCondition{
				left:   &conditionField{sym: &SymbolToken{Content: "field", Position: 0}},
				op:     &OperatorToken{Type: "IS", RawContent: "IS", Position: 6},
				opType: "IS",
				right:  &conditionValue{null: &KeywordToken{Name: "NULL", RawContent: "NULL", Position: 9}},
			},
		},
		{
			name: "SpecialOperator_HAS_ANCESTOR",
			tokens: []Token{
				&SymbolToken{Content: "__key__"},
				&WhitespaceToken{Content: " ", Position: 7},
				&OperatorToken{Type: "HAS", RawContent: "HAS", Position: 8},
				&WhitespaceToken{Content: " ", Position: 11},
				&OperatorToken{Type: "ANCESTOR", RawContent: "ANCESTOR", Position: 12},
				&WhitespaceToken{Content: " ", Position: 20},
				&KeywordToken{Name: "KEY", RawContent: "KEY", Position: 21},
				&OperatorToken{Type: "(", RawContent: "(", Position: 22},
				&SymbolToken{Content: "EntityName", Position: 23},
				&OperatorToken{Type: ",", RawContent: ",", Position: 33},
				&NumericToken{Int64: 123, Floating: false, RawContent: "123", Position: 34},
				&OperatorToken{Type: ")", RawContent: ")", Position: 37},
			},
			want: &forwardComparatorCondition{
				left:   &conditionField{sym: &SymbolToken{Content: "__key__"}},
				op:     &OperatorToken{Type: "HAS", RawContent: "HAS", Position: 8},
				opType: "HAS ANCESTOR",
				right: &conditionKey{
					keyKeyword: &KeywordToken{Name: "KEY", RawContent: "KEY", Position: 21},
					key: &Key{
						Path: []*KeyPath{
							{
								Kind: "EntityName",
								ID:   int64(123),
							},
						},
					},
				},
			},
		},
		{
			name: "GroupedCondition_Normal",
			tokens: []Token{
				&OperatorToken{Type: "(", RawContent: "(", Position: 0},
				&SymbolToken{Content: "field", Position: 1},
				&WhitespaceToken{Content: " ", Position: 6},
				&OperatorToken{Type: "=", RawContent: "=", Position: 7},
				&WhitespaceToken{Content: " ", Position: 8},
				&NumericToken{Int64: 1, Floating: false, RawContent: "1", Position: 9},
				&OperatorToken{Type: ")", RawContent: ")", Position: 10},
			},
			want: &forwardComparatorCondition{
				left:   &conditionField{sym: &SymbolToken{Content: "field", Position: 1}},
				op:     &OperatorToken{Type: "=", RawContent: "=", Position: 7},
				opType: "=",
				right:  &conditionValue{n: &NumericToken{Int64: 1, Floating: false, RawContent: "1", Position: 9}},
			},
		},

		// Backward operators
		{
			name: "InfixOperator_HAS_DESCENDANT",
			tokens: []Token{
				&StringToken{Quote: '"', Content: "value", RawContent: "\"value\"", Position: 0},
				&WhitespaceToken{Content: " ", Position: 7},
				&OperatorToken{Type: "HAS", RawContent: "HAS", Position: 8},
				&WhitespaceToken{Content: " ", Position: 11},
				&OperatorToken{Type: "DESCENDANT", RawContent: "DESCENDANT", Position: 12},
				&WhitespaceToken{Content: " ", Position: 22},
				&SymbolToken{Content: "field", Position: 23},
			},
			want: &backwardComparatorCondition{
				left:   &conditionValue{s: &StringToken{Quote: '"', Content: "value", RawContent: "\"value\"", Position: 0}},
				op:     &OperatorToken{Type: "HAS", RawContent: "HAS", Position: 8},
				opType: "HAS DESCENDANT",
				right:  &conditionField{sym: &SymbolToken{Content: "field", Position: 23}},
			},
		},

		// Special operators
		{
			name: "SpecialOperator_NOT_IN",
			tokens: []Token{
				&SymbolToken{Content: "field", Position: 0},
				&WhitespaceToken{Content: " ", Position: 5},
				&OperatorToken{Type: "NOT", RawContent: "NOT", Position: 6},
				&WhitespaceToken{Content: " ", Position: 9},
				&OperatorToken{Type: "IN", RawContent: "IN", Position: 10},
				&WhitespaceToken{Content: " ", Position: 12},
				&KeywordToken{Name: "ARRAY", RawContent: "ARRAY", Position: 13},
				&OperatorToken{Type: "(", RawContent: "(", Position: 18},
				&StringToken{Quote: '"', Content: "val1", RawContent: "\"val1\"", Position: 19},
				&OperatorToken{Type: ",", RawContent: ",", Position: 25},
				&StringToken{Quote: '"', Content: "val2", RawContent: "\"val2\"", Position: 26},
				&OperatorToken{Type: ")", RawContent: ")", Position: 32},
			},
			want: &forwardComparatorCondition{
				left:   &conditionField{sym: &SymbolToken{Content: "field", Position: 0}},
				op:     &OperatorToken{Type: "NOT", RawContent: "NOT", Position: 6},
				opType: "NOT IN",
				right: &conditionArray{
					arrayKeyword: &KeywordToken{Name: "ARRAY", RawContent: "ARRAY", Position: 13},
					values: []valueAST{
						&conditionValue{s: &StringToken{Quote: '"', Content: "val1", RawContent: "\"val1\"", Position: 19}},
						&conditionValue{s: &StringToken{Quote: '"', Content: "val2", RawContent: "\"val2\"", Position: 26}},
					},
				},
			},
		},

		// Operator precedence tests
		{
			name: "OperatorPrecedence_AND_OR",
			tokens: []Token{
				&SymbolToken{Content: "a", Position: 0},
				&WhitespaceToken{Content: " ", Position: 1},
				&OperatorToken{Type: "=", RawContent: "=", Position: 2},
				&WhitespaceToken{Content: " ", Position: 3},
				&NumericToken{Int64: 1, Floating: false, RawContent: "1", Position: 4},
				&WhitespaceToken{Content: " ", Position: 5},
				&OperatorToken{Type: "OR", RawContent: "OR", Position: 6},
				&WhitespaceToken{Content: " ", Position: 8},
				&SymbolToken{Content: "b", Position: 9},
				&WhitespaceToken{Content: " ", Position: 10},
				&OperatorToken{Type: "=", RawContent: "=", Position: 11},
				&WhitespaceToken{Content: " ", Position: 12},
				&NumericToken{Int64: 2, Floating: false, RawContent: "2", Position: 13},
				&WhitespaceToken{Content: " ", Position: 14},
				&OperatorToken{Type: "AND", RawContent: "AND", Position: 15},
				&WhitespaceToken{Content: " ", Position: 18},
				&SymbolToken{Content: "c", Position: 19},
				&WhitespaceToken{Content: " ", Position: 20},
				&OperatorToken{Type: "=", RawContent: "=", Position: 21},
				&WhitespaceToken{Content: " ", Position: 22},
				&NumericToken{Int64: 3, Floating: false, RawContent: "3", Position: 23},
			},
			want: &compoundComparatorCondition{
				left: &forwardComparatorCondition{
					left:   &conditionField{sym: &SymbolToken{Content: "a", Position: 0}},
					op:     &OperatorToken{Type: "=", RawContent: "=", Position: 2},
					opType: "=",
					right:  &conditionValue{n: &NumericToken{Int64: 1, Floating: false, RawContent: "1", Position: 4}},
				},
				op: &OperatorToken{Type: "OR", RawContent: "OR", Position: 6},
				right: &compoundComparatorCondition{
					left: &forwardComparatorCondition{
						left:   &conditionField{sym: &SymbolToken{Content: "b", Position: 9}},
						op:     &OperatorToken{Type: "=", RawContent: "=", Position: 11},
						opType: "=",
						right:  &conditionValue{n: &NumericToken{Int64: 2, Floating: false, RawContent: "2", Position: 13}},
					},
					op: &OperatorToken{Type: "AND", RawContent: "AND", Position: 15},
					right: &forwardComparatorCondition{
						left:   &conditionField{sym: &SymbolToken{Content: "c", Position: 19}},
						op:     &OperatorToken{Type: "=", RawContent: "=", Position: 21},
						opType: "=",
						right:  &conditionValue{n: &NumericToken{Int64: 3, Floating: false, RawContent: "3", Position: 23}},
					},
				},
			},
		},

		// Error cases
		{
			name:    "Error_EmptyTokens",
			tokens:  []Token{},
			wantErr: ErrNoTokens,
		},
		{
			name: "Error_UnexpectedToken",
			tokens: []Token{
				&OrderToken{Descending: false, RawContent: "ASC", Position: 0},
			},
			wantErr: ErrUnexpectedToken,
		},
		{
			name: "Error_UnexpectedKeyword",
			tokens: []Token{
				&KeywordToken{Name: "UNKNOWN", RawContent: "UNKNOWN", Position: 0},
			},
			wantErr: ErrUnexpectedToken,
		},
		{
			name: "Error_PrefixOperator_MissingClosingParen",
			tokens: []Token{
				&OperatorToken{Type: "(", RawContent: "(", Position: 0},
				&SymbolToken{Content: "field", Position: 1},
			},
			wantErr: ErrUnexpectedToken,
		},
		{
			name: "Error_PrefixOperator_InvalidNumericToken",
			tokens: []Token{
				&OperatorToken{Type: "+", RawContent: "+", Position: 0},
				&StringToken{Quote: '"', Content: "invalid", RawContent: "\"invalid\"", Position: 1},
			},
			wantErr: ErrUnexpectedToken,
		},
		{
			name: "Error_InvalidPrefixOperator",
			tokens: []Token{
				&OperatorToken{Type: "INVALID", RawContent: "INVALID", Position: 0},
			},
			wantErr: ErrUnexpectedToken,
		},
		{
			name: "Error_PrefixOnlySpecialOperator",
			tokens: []Token{
				&StringToken{Quote: '"', Content: "value", RawContent: "\"value\"", Position: 0},
				&WhitespaceToken{Content: " ", Position: 7},
				&OperatorToken{Type: "HAS", RawContent: "HAS", Position: 8},
				&WhitespaceToken{Content: " ", Position: 11},
				&SymbolToken{Content: "field", Position: 20},
			},
			wantErr: ErrUnexpectedToken,
		},
		{
			name: "Error_SpecialOperatorWithoutWhitespace",
			tokens: []Token{
				&StringToken{Quote: '"', Content: "value", RawContent: "\"value\"", Position: 0},
				&WhitespaceToken{Content: " ", Position: 7},
				&OperatorToken{Type: "HAS", RawContent: "HAS", Position: 8},
				&SymbolToken{Content: "field", Position: 20},
			},
			wantErr: ErrUnexpectedToken,
		},
		{
			name: "Error_BrokenSpecialOperator",
			tokens: []Token{
				&StringToken{Quote: '"', Content: "value", RawContent: "\"value\"", Position: 0},
				&WhitespaceToken{Content: " ", Position: 7},
				&OperatorToken{Type: "HAS", RawContent: "HAS", Position: 8},
				&WhitespaceToken{Content: " ", Position: 11},
			},
			wantErr: ErrUnexpectedToken,
		},
		{
			name: "Error_InvalidSpecialOperator",
			tokens: []Token{
				&StringToken{Quote: '"', Content: "value", RawContent: "\"value\"", Position: 0},
				&WhitespaceToken{Content: " ", Position: 7},
				&OperatorToken{Type: "HAS", RawContent: "HAS", Position: 8},
				&WhitespaceToken{Content: " ", Position: 11},
				&OperatorToken{Type: "INVALID", RawContent: "INVALID", Position: 12},
				&WhitespaceToken{Content: " ", Position: 19},
				&SymbolToken{Content: "field", Position: 20},
			},
			wantErr: ErrUnexpectedToken,
		},
		{
			name: "ErrKeyNoBody",
			tokens: []Token{
				&KeywordToken{Name: "KEY", RawContent: "KEY", Position: 0},
				&OperatorToken{Type: "(", RawContent: "(", Position: 3},
			},
			wantErr: ErrNoTokens,
		},
		{
			name: "ErrArrayNoBody",
			tokens: []Token{
				&KeywordToken{Name: "ARRAY", RawContent: "ARRAY", Position: 0},
				&OperatorToken{Type: "(", RawContent: "(", Position: 5},
			},
			wantErr: ErrNoTokens,
		},
		{
			name: "ErrBlobNoBody",
			tokens: []Token{
				&KeywordToken{Name: "BLOB", RawContent: "BLOB", Position: 0},
				&OperatorToken{Type: "(", RawContent: "(", Position: 5},
			},
			wantErr: ErrNoTokens,
		},
		{
			name: "ErrKeyDateTimeBody",
			tokens: []Token{
				&KeywordToken{Name: "DATETIME", RawContent: "DATETIME", Position: 0},
				&OperatorToken{Type: "(", RawContent: "(", Position: 3},
			},
			wantErr: ErrNoTokens,
		},
		{
			name: "SpecialOperator_NOT_UNKNOWN_Error",
			tokens: []Token{
				&SymbolToken{Content: "field", Position: 0},
				&WhitespaceToken{Content: " ", Position: 5},
				&OperatorToken{Type: "NOT", RawContent: "NOT", Position: 6},
				&WhitespaceToken{Content: " ", Position: 9},
				&OperatorToken{Type: "UNKNOWN", RawContent: "UNKNOWN", Position: 10},
			},
			wantErr: ErrUnexpectedToken,
		},
		{
			name: "DotOperator_RightNotPropertyAST_Error",
			tokens: []Token{
				&SymbolToken{Content: "field", Position: 0},
				&OperatorToken{Type: ".", RawContent: ".", Position: 5},
				&NumericToken{Int64: 123, Floating: false, RawContent: "123", Position: 6},
			},
			wantErr: ErrUnexpectedToken,
		},
		{
			name: "GroupedCondition_Empty_Error",
			tokens: []Token{
				&OperatorToken{Type: "(", RawContent: "(", Position: 0},
				&OperatorToken{Type: ")", RawContent: ")", Position: 1},
			},
			wantErr: ErrUnexpectedToken,
		},
		{
			name: "BackwardOperator_RightNotPropertyAST_Error",
			tokens: []Token{
				&NumericToken{Int64: 123, Floating: false, RawContent: "123", Position: 0},
				&OperatorToken{Type: "=", RawContent: "=", Position: 3},
				&NumericToken{Int64: 456, Floating: false, RawContent: "456", Position: 4},
			},
			wantErr: ErrUnexpectedToken,
		},
		{
			name: "ForwardOperator_RightNotValueAST_Error",
			tokens: []Token{
				&SymbolToken{Content: "field", Position: 0},
				&OperatorToken{Type: "=", RawContent: "=", Position: 5},
				&SymbolToken{Content: "field2", Position: 6},
			},
			wantErr: ErrUnexpectedToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create token reader
			tokenSource := defaultTokenSourceFactory.NewSliceTokenSource(tt.tokens)

			// Call constructAST
			got, err := constructConditionAST(tokenSource, 0)

			// Check error conditions
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("constructAST() expected no error, got %v", err)
				}
			} else {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("constructAST() error = %v, want error containing %v", err, tt.wantErr)
				}
				return // Skip further checks if we expect an error
			}

			// Compare results
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(
				conditionField{}, conditionFieldAccess{}, conditionValue{}, conditionKey{}, conditionArray{},
				conditionBlob{}, conditionDateTime{},
				forwardComparatorCondition{}, backwardComparatorCondition{},
				compoundComparatorCondition{},
			)); diff != "" {
				t.Errorf("constructAST() mismatch (-want +got):\n%s", diff)
			}

			if token, err := tokenSource.Read(); err != ErrEndOfToken {
				if err == nil {
					t.Errorf("constructAST() expected end of token, got %v", token)
				} else {
					t.Errorf("constructAST() expected end of token, got %v", err)
				}
			}
		})
	}
}

func TestConstructAST_EdgeCases(t *testing.T) {
	t.Run("MinBP_FilteringLowPrecedence", func(t *testing.T) {
		// Test with higher minBP to filter out low precedence operators
		tokens := []Token{
			&SymbolToken{Content: "field", Position: 0},
			&WhitespaceToken{Content: " ", Position: 5},
			&OperatorToken{Type: "OR", RawContent: "OR", Position: 6}, // BP = 1
			&WhitespaceToken{Content: " ", Position: 8},
			&SymbolToken{Content: "field2", Position: 9},
		}

		tokenSource := defaultTokenSourceFactory.NewSliceTokenSource(tokens)
		got, err := constructConditionAST(tokenSource, 2) // minBP = 2, should filter out OR (BP=1)

		if err != nil {
			t.Errorf("constructAST() unexpected error = %v", err)
			return
		}

		want := &conditionField{sym: &SymbolToken{Content: "field", Position: 0}}
		if diff := cmp.Diff(want, got, cmp.AllowUnexported(conditionField{})); diff != "" {
			t.Errorf("constructAST() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("ComplexNesting", func(t *testing.T) {
		// Test nested parentheses with complex expressions
		tokens := []Token{
			&OperatorToken{Type: "(", RawContent: "(", Position: 0},
			&SymbolToken{Content: "a", Position: 1},
			&WhitespaceToken{Content: " ", Position: 2},
			&OperatorToken{Type: "=", RawContent: "=", Position: 3},
			&WhitespaceToken{Content: " ", Position: 4},
			&NumericToken{Int64: 1, Floating: false, RawContent: "1", Position: 5},
			&WhitespaceToken{Content: " ", Position: 6},
			&OperatorToken{Type: "AND", RawContent: "AND", Position: 7},
			&WhitespaceToken{Content: " ", Position: 10},
			&SymbolToken{Content: "b", Position: 11},
			&OperatorToken{Type: ".", Position: 12},
			&SymbolToken{Content: "c", Position: 13},
			&OperatorToken{Type: ".", Position: 14},
			&SymbolToken{Content: "d", Position: 15},
			&OperatorToken{Type: ".", Position: 16},
			&SymbolToken{Content: "e", Position: 17},
			&OperatorToken{Type: ".", Position: 18},
			&SymbolToken{Content: "f", Position: 19},
			&WhitespaceToken{Content: " ", Position: 20},
			&OperatorToken{Type: "=", RawContent: "=", Position: 21},
			&WhitespaceToken{Content: " ", Position: 22},
			&NumericToken{Int64: 2, Floating: false, RawContent: "2", Position: 23},
			&OperatorToken{Type: ")", RawContent: ")", Position: 24},
		}

		tokenSource := defaultTokenSourceFactory.NewSliceTokenSource(tokens)
		got, err := constructConditionAST(tokenSource, 0)

		if err != nil {
			t.Errorf("constructAST() unexpected error = %v", err)
			return
		}

		want := &compoundComparatorCondition{
			left: &forwardComparatorCondition{
				left:   &conditionField{sym: &SymbolToken{Content: "a", Position: 1}},
				op:     &OperatorToken{Type: "=", RawContent: "=", Position: 3},
				opType: "=",
				right:  &conditionValue{n: &NumericToken{Int64: 1, Floating: false, RawContent: "1", Position: 5}},
			},
			op: &OperatorToken{Type: "AND", RawContent: "AND", Position: 7},
			right: &forwardComparatorCondition{
				left: &conditionFieldAccess{
					left: &conditionFieldAccess{
						left: &conditionFieldAccess{
							left: &conditionFieldAccess{
								left:  &conditionField{sym: &SymbolToken{Content: "b", Position: 11}},
								right: &conditionField{sym: &SymbolToken{Content: "c", Position: 13}},
							},
							right: &conditionField{sym: &SymbolToken{Content: "d", Position: 15}},
						},
						right: &conditionField{sym: &SymbolToken{Content: "e", Position: 17}},
					},
					right: &conditionField{sym: &SymbolToken{Content: "f", Position: 19}},
				},
				op:     &OperatorToken{Type: "=", RawContent: "=", Position: 21},
				opType: "=",
				right:  &conditionValue{n: &NumericToken{Int64: 2, Floating: false, RawContent: "2", Position: 23}},
			},
		}

		if diff := cmp.Diff(want, got, cmp.AllowUnexported(
			conditionField{}, conditionFieldAccess{}, conditionValue{}, forwardComparatorCondition{}, compoundComparatorCondition{},
		)); diff != "" {
			t.Errorf("constructAST() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("ReadError", func(t *testing.T) {
		t.Parallel()
		for _, tt := range []struct {
			name     string
			position int
		}{
			{
				name:     "FirstToken",
				position: 0,
			},
			{
				name:     "FirstWhitespace",
				position: 1,
			},
			{
				name:     "OperatorToken",
				position: 2,
			},
			{
				name:     "SecondWhitespace",
				position: 3,
			},
			{
				name:     "NumericToken",
				position: 4,
			},
		} {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				ts := defaultTokenSourceFactory.NewErrorTokenSource([]Token{
					&SymbolToken{Content: "field", Position: 0},
					&WhitespaceToken{Content: " ", Position: 5},
					&OperatorToken{Type: "=", RawContent: "=", Position: 6},
					&WhitespaceToken{Content: " ", Position: 7},
					&NumericToken{Int64: 42, Floating: false, RawContent: "42", Position: 8},
				}, map[int]error{
					tt.position: io.ErrUnexpectedEOF,
				})

				got, err := constructConditionAST(ts, 0)
				if err == nil {
					t.Errorf("constructAST() expected error, got = %v", got)
				}
			})
		}
	})
}

func TestConstructAST_AdvancedCases(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []Token
		minBP   uint8
		want    conditionAST
		wantErr error
	}{
		// Test with ARRAY keyword
		{
			name: "KeywordToken_ARRAY",
			tokens: []Token{
				&KeywordToken{Name: "ARRAY", RawContent: "ARRAY", Position: 0},
				&OperatorToken{Type: "(", RawContent: "(", Position: 5},
				&StringToken{Quote: '"', Content: "val1", RawContent: "\"val1\"", Position: 6},
				&OperatorToken{Type: ",", RawContent: ",", Position: 12},
				&NumericToken{Int64: 42, Floating: false, RawContent: "42", Position: 13},
				&OperatorToken{Type: ")", RawContent: ")", Position: 15},
			},
			minBP: 0,
			want: &conditionArray{
				arrayKeyword: &KeywordToken{Name: "ARRAY", RawContent: "ARRAY", Position: 0},
				values: []valueAST{
					&conditionValue{s: &StringToken{Quote: '"', Content: "val1", RawContent: "\"val1\"", Position: 6}},
					&conditionValue{n: &NumericToken{Int64: 42, Floating: false, RawContent: "42", Position: 13}},
				},
			},
		},
		// Test with BLOB keyword
		{
			name: "KeywordToken_BLOB",
			tokens: []Token{
				&KeywordToken{Name: "BLOB", RawContent: "BLOB", Position: 0},
				&OperatorToken{Type: "(", RawContent: "(", Position: 4},
				&StringToken{Quote: '"', Content: "SGVsbG8gV29ybGQ", RawContent: "\"SGVsbG8gV29ybGQ\"", Position: 5},
				&OperatorToken{Type: ")", RawContent: ")", Position: 22},
			},
			minBP: 0,
			want: &conditionBlob{
				blobKeyword: &KeywordToken{Name: "BLOB", RawContent: "BLOB", Position: 0},
				b:           []byte("Hello World"),
			},
		},
		// Test with DATETIME keyword
		{
			name: "KeywordToken_DATETIME",
			tokens: []Token{
				&KeywordToken{Name: "DATETIME", RawContent: "DATETIME", Position: 0},
				&OperatorToken{Type: "(", RawContent: "(", Position: 8},
				&StringToken{Quote: '"', Content: "2023-01-15T10:30:00Z", RawContent: "\"2023-01-15T10:30:00Z\"", Position: 9},
				&OperatorToken{Type: ")", RawContent: ")", Position: 32},
			},
			minBP: 0,
			want: &conditionDateTime{
				dateTimeKeyword: &KeywordToken{Name: "DATETIME", RawContent: "DATETIME", Position: 0},
				t:               mustParseTime("2023-01-15T10:30:00Z"),
			},
		},
		// Test IS operator with NULL
		{
			name: "InfixOperator_IS_NULL",
			tokens: []Token{
				&SymbolToken{Content: "field", Position: 0},
				&WhitespaceToken{Content: " ", Position: 5},
				&OperatorToken{Type: "IS", RawContent: "IS", Position: 6},
				&WhitespaceToken{Content: " ", Position: 8},
				&KeywordToken{Name: "NULL", RawContent: "NULL", Position: 9},
			},
			minBP: 0,
			want: &forwardComparatorCondition{
				left:   &conditionField{sym: &SymbolToken{Content: "field", Position: 0}},
				op:     &OperatorToken{Type: "IS", RawContent: "IS", Position: 6},
				opType: "IS",
				right:  &conditionValue{null: &KeywordToken{Name: "NULL", RawContent: "NULL", Position: 9}},
			},
		},
		// Test comparison operators - less than, greater than, etc.
		{
			name: "InfixOperator_LessThan",
			tokens: []Token{
				&SymbolToken{Content: "age", Position: 0},
				&WhitespaceToken{Content: " ", Position: 3},
				&OperatorToken{Type: "<", RawContent: "<", Position: 4},
				&WhitespaceToken{Content: " ", Position: 5},
				&NumericToken{Int64: 30, Floating: false, RawContent: "30", Position: 6},
			},
			minBP: 0,
			want: &forwardComparatorCondition{
				left:   &conditionField{sym: &SymbolToken{Content: "age", Position: 0}},
				op:     &OperatorToken{Type: "<", RawContent: "<", Position: 4},
				opType: "<",
				right:  &conditionValue{n: &NumericToken{Int64: 30, Floating: false, RawContent: "30", Position: 6}},
			},
		},
		{
			name: "InfixOperator_GreaterThanEqual",
			tokens: []Token{
				&SymbolToken{Content: "score", Position: 0},
				&WhitespaceToken{Content: " ", Position: 5},
				&OperatorToken{Type: ">=", RawContent: ">=", Position: 6},
				&WhitespaceToken{Content: " ", Position: 8},
				&NumericToken{Float64: 85.5, Floating: true, RawContent: "85.5", Position: 9},
			},
			minBP: 0,
			want: &forwardComparatorCondition{
				left:   &conditionField{sym: &SymbolToken{Content: "score", Position: 0}},
				op:     &OperatorToken{Type: ">=", RawContent: ">=", Position: 6},
				opType: ">=",
				right:  &conditionValue{n: &NumericToken{Float64: 85.5, Floating: true, RawContent: "85.5", Position: 9}},
			},
		},
		// Test backward comparisons
		{
			name: "InfixOperator_Backward_LessThan",
			tokens: []Token{
				&NumericToken{Int64: 100, Floating: false, RawContent: "100", Position: 0},
				&WhitespaceToken{Content: " ", Position: 3},
				&OperatorToken{Type: "<", RawContent: "<", Position: 4},
				&WhitespaceToken{Content: " ", Position: 5},
				&SymbolToken{Content: "maxValue", Position: 6},
			},
			minBP: 0,
			want: &backwardComparatorCondition{
				left:   &conditionValue{n: &NumericToken{Int64: 100, Floating: false, RawContent: "100", Position: 0}},
				op:     &OperatorToken{Type: "<", RawContent: "<", Position: 4},
				opType: "<", // Not inverted here, inversion happens in toCondition()
				right:  &conditionField{sym: &SymbolToken{Content: "maxValue", Position: 6}},
			},
		},
		// Test IN operator
		{
			name: "InfixOperator_IN",
			tokens: []Token{
				&SymbolToken{Content: "category", Position: 0},
				&WhitespaceToken{Content: " ", Position: 8},
				&OperatorToken{Type: "IN", RawContent: "IN", Position: 9},
				&WhitespaceToken{Content: " ", Position: 11},
				&KeywordToken{Name: "ARRAY", RawContent: "ARRAY", Position: 12},
				&OperatorToken{Type: "(", RawContent: "(", Position: 17},
				&StringToken{Quote: '"', Content: "A", RawContent: "\"A\"", Position: 18},
				&OperatorToken{Type: ",", RawContent: ",", Position: 21},
				&StringToken{Quote: '"', Content: "B", RawContent: "\"B\"", Position: 22},
				&OperatorToken{Type: ")", RawContent: ")", Position: 25},
			},
			minBP: 0,
			want: &forwardComparatorCondition{
				left:   &conditionField{sym: &SymbolToken{Content: "category", Position: 0}},
				op:     &OperatorToken{Type: "IN", RawContent: "IN", Position: 9},
				opType: "IN",
				right: &conditionArray{
					arrayKeyword: &KeywordToken{Name: "ARRAY", RawContent: "ARRAY", Position: 12},
					values: []valueAST{
						&conditionValue{s: &StringToken{Quote: '"', Content: "A", RawContent: "\"A\"", Position: 18}},
						&conditionValue{s: &StringToken{Quote: '"', Content: "B", RawContent: "\"B\"", Position: 22}},
					},
				},
			},
		},
		// Error case: missing right operand
		{
			name: "Error_MissingRightOperand",
			tokens: []Token{
				&SymbolToken{Content: "field", Position: 0},
				&WhitespaceToken{Content: " ", Position: 5},
				&OperatorToken{Type: "=", RawContent: "=", Position: 6},
				&WhitespaceToken{Content: " ", Position: 7},
				// Missing right operand - constructAST will call recursively and get ErrEndOfToken
			},
			minBP:   0,
			wantErr: ErrUnexpectedToken,
		},
		// Error case: incomplete special operator
		{
			name: "Error_IncompleteSpecialOperator",
			tokens: []Token{
				&SymbolToken{Content: "field", Position: 0},
				&WhitespaceToken{Content: " ", Position: 5},
				&OperatorToken{Type: "HAS", RawContent: "HAS", Position: 6},
				&WhitespaceToken{Content: " ", Position: 9},
				&OperatorToken{Type: "INVALID", RawContent: "INVALID", Position: 10},
			},
			minBP:   0,
			wantErr: ErrUnexpectedToken,
		},
		// Error case: wrong type after special operator
		{
			name: "Error_WrongTypeAfterSpecialOperator",
			tokens: []Token{
				&SymbolToken{Content: "field", Position: 0},
				&WhitespaceToken{Content: " ", Position: 5},
				&OperatorToken{Type: "NOT", RawContent: "NOT", Position: 6},
				&WhitespaceToken{Content: " ", Position: 9},
				&StringToken{Quote: '"', Content: "wrong", RawContent: "\"wrong\"", Position: 10},
			},
			minBP:   0,
			wantErr: ErrUnexpectedToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create token reader
			tokenSource := defaultTokenSourceFactory.NewSliceTokenSource(tt.tokens)

			// Call constructAST
			got, err := constructConditionAST(tokenSource, tt.minBP)

			// Check error conditions
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("constructAST() expected no error, got %v", err)
				}
			} else {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("constructAST() error = %v, want error containing %v", err, tt.wantErr)
				}
				return // Skip further checks if we expect an error
			}

			// Check for unexpected errors
			if err != nil {
				t.Errorf("constructAST() unexpected error = %v", err)
				return
			}

			// Compare results
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(
				conditionField{}, conditionFieldAccess{}, conditionValue{}, conditionKey{}, conditionArray{},
				conditionBlob{}, conditionDateTime{},
				forwardComparatorCondition{}, backwardComparatorCondition{},
				compoundComparatorCondition{},
			)); diff != "" {
				t.Errorf("constructAST() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// Helper function to parse time for tests
func mustParseTime(timeStr string) time.Time {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		panic("Invalid time format in test: " + timeStr)
	}
	return t
}
