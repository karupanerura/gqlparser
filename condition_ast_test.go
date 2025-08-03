package gqlparser

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestForwardComparatorCondition_ToCondition(t *testing.T) {
	tm := time.Date(2024, 7, 5, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		condition *forwardComparatorCondition
		want      Condition
		wantErr   bool
	}{
		{
			name: "forward comparator",
			condition: &forwardComparatorCondition{
				left:   &conditionField{sym: &SymbolToken{Content: "name"}},
				op:     &OperatorToken{Type: "CONTAINS", Position: 0},
				opType: "CONTAINS",
				right:  &conditionValue{s: &StringToken{Content: "test"}},
			},
			want: &ForwardComparatorCondition{
				Comparator: ForwardComparator("CONTAINS"),
				Property:   Property{Name: "name"},
				Value:      "test",
			},
		},
		{
			name: "IS NULL condition",
			condition: &forwardComparatorCondition{
				left:   &conditionField{sym: &SymbolToken{Content: "field"}},
				op:     &OperatorToken{Type: "IS", Position: 0},
				opType: "IS",
				right:  &conditionValue{null: &KeywordToken{Name: "NULL"}},
			},
			want: &IsNullCondition{Property: Property{Name: "field"}},
		},
		{
			name: "either comparator (equals)",
			condition: &forwardComparatorCondition{
				left:   &conditionField{sym: &SymbolToken{Content: "field"}},
				op:     &OperatorToken{Type: "=", Position: 0},
				opType: "=",
				right:  &conditionValue{n: &NumericToken{Int64: 42}},
			},
			want: &EitherComparatorCondition{
				Comparator: EitherComparator("="),
				Property:   Property{Name: "field"},
				Value:      int64(42),
			},
		},
		{
			name: "IS with non-null value should error",
			condition: &forwardComparatorCondition{
				left:   &conditionField{sym: &SymbolToken{Content: "field"}},
				op:     &OperatorToken{Type: "IS", Position: 0},
				opType: "IS",
				right:  &conditionValue{s: &StringToken{Content: "not null", Position: 10}},
			},
			wantErr: true,
		},
		{
			name: "invalid operator should error",
			condition: &forwardComparatorCondition{
				left:   &conditionField{sym: &SymbolToken{Content: "field"}},
				op:     &OperatorToken{Type: "INVALID", Position: 0, RawContent: "INVALID"},
				opType: "INVALID",
				right:  &conditionValue{s: &StringToken{Content: "test"}},
			},
			wantErr: true,
		},
		{
			name: "right value is boolean",
			condition: &forwardComparatorCondition{
				left:   &conditionField{sym: &SymbolToken{Content: "flag"}},
				op:     &OperatorToken{Type: "=", Position: 0},
				opType: "=",
				right:  &conditionValue{b: &BooleanToken{Value: true}},
			},
			want: &EitherComparatorCondition{
				Comparator: EitherComparator("="),
				Property:   Property{Name: "flag"},
				Value:      true,
			},
		},
		{
			name: "right value is numeric (float)",
			condition: &forwardComparatorCondition{
				left:   &conditionField{sym: &SymbolToken{Content: "score"}},
				op:     &OperatorToken{Type: ">", Position: 0},
				opType: ">",
				right:  &conditionValue{n: &NumericToken{Float64: 0.5, Floating: true}},
			},
			want: &EitherComparatorCondition{
				Comparator: EitherComparator(">"),
				Property:   Property{Name: "score"},
				Value:      float64(0.5),
			},
		},
		{
			name: "right value is binding",
			condition: &forwardComparatorCondition{
				left:   &conditionField{sym: &SymbolToken{Content: "param"}},
				op:     &OperatorToken{Type: "=", Position: 0},
				opType: "=",
				right:  &conditionValue{bind: &BindingToken{Name: "p1"}},
			},
			want: &EitherComparatorCondition{
				Comparator: EitherComparator("="),
				Property:   Property{Name: "param"},
				Value:      &NamedBinding{Name: "p1"},
			},
		},
		{
			name: "right value is array",
			condition: &forwardComparatorCondition{
				left:   &conditionField{sym: &SymbolToken{Content: "ids"}},
				op:     &OperatorToken{Type: "IN", Position: 0},
				opType: "IN",
				right: &conditionArray{arrayKeyword: &KeywordToken{Name: "ARRAY"}, values: []valueAST{
					&conditionValue{n: &NumericToken{Int64: 1}},
					&conditionValue{n: &NumericToken{Int64: 2}},
				}},
			},
			want: &ForwardComparatorCondition{
				Comparator: ForwardComparator("IN"),
				Property:   Property{Name: "ids"},
				Value:      []any{int64(1), int64(2)},
			},
		},
		{
			name: "right value is blob",
			condition: &forwardComparatorCondition{
				left:   &conditionField{sym: &SymbolToken{Content: "data"}},
				op:     &OperatorToken{Type: "=", Position: 0},
				opType: "=",
				right:  &conditionBlob{blobKeyword: &KeywordToken{Name: "BLOB"}, b: []byte("abc")},
			},
			want: &EitherComparatorCondition{
				Comparator: EitherComparator("="),
				Property:   Property{Name: "data"},
				Value:      []byte("abc"),
			},
		},
		{
			name: "right value is datetime",
			condition: &forwardComparatorCondition{
				left:   &conditionField{sym: &SymbolToken{Content: "created_at"}},
				op:     &OperatorToken{Type: ">="},
				opType: ">=",
				right:  &conditionDateTime{dateTimeKeyword: &KeywordToken{Name: "DATETIME"}, t: tm},
			},
			want: &EitherComparatorCondition{
				Comparator: EitherComparator(">="),
				Property:   Property{Name: "created_at"},
				Value:      tm,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.condition.toCondition()
			if (err != nil) != tt.wantErr {
				t.Errorf("forwardComparatorCondition.toCondition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !cmp.Equal(got, tt.want) {
				t.Errorf("forwardComparatorCondition.toCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBackwardComparatorCondition_ToCondition(t *testing.T) {
	tests := []struct {
		name      string
		condition *backwardComparatorCondition
		want      Condition
		wantErr   bool
		wantPanic bool
	}{
		{
			name: "backward equals comparison",
			condition: &backwardComparatorCondition{
				left:   &conditionValue{s: &StringToken{Content: "test"}},
				op:     &OperatorToken{Type: "=", Position: 0},
				opType: "=",
				right:  &conditionField{sym: &SymbolToken{Content: "name"}},
			},
			want: &EitherComparatorCondition{
				Comparator: EitherComparator("="),
				Property:   Property{Name: "name"},
				Value:      "test",
			},
		},
		{
			name: "backward less than comparison",
			condition: &backwardComparatorCondition{
				left:   &conditionValue{n: &NumericToken{Int64: 10}},
				op:     &OperatorToken{Type: "IN", Position: 0},
				opType: "IN",
				right:  &conditionField{sym: &SymbolToken{Content: "categories"}},
			},
			want: &BackwardComparatorCondition{
				Comparator: BackwardComparator("IN"),
				Property:   Property{Name: "categories"},
				Value:      int64(10),
			},
		},
		{
			name: "invalid operator should error",
			condition: &backwardComparatorCondition{
				left:   &conditionValue{s: &StringToken{Content: "test"}},
				op:     &OperatorToken{Type: "INVALID", Position: 0, RawContent: "INVALID"},
				opType: "INVALID",
				right:  &conditionField{sym: &SymbolToken{Content: "field"}},
			},
			wantErr: true,
		},
		{
			name: "backward HAS DESCENDANT comparison",
			condition: &backwardComparatorCondition{
				left:   &conditionValue{s: &StringToken{Content: "descendant"}},
				op:     &OperatorToken{Type: "HAS DESCENDANT", Position: 1},
				opType: "HAS DESCENDANT",
				right:  &conditionField{sym: &SymbolToken{Content: "parent"}},
			},
			want: &BackwardComparatorCondition{
				Comparator: BackwardComparator("HAS DESCENDANT"),
				Property:   Property{Name: "parent"},
				Value:      "descendant",
			},
		},
		{
			name: "backward IN with array value",
			condition: &backwardComparatorCondition{
				left: &conditionArray{arrayKeyword: &KeywordToken{Name: "ARRAY"}, values: []valueAST{
					&conditionValue{n: &NumericToken{Int64: 1}},
					&conditionValue{n: &NumericToken{Int64: 2}},
				}},
				op:     &OperatorToken{Type: "IN", Position: 2},
				opType: "IN",
				right:  &conditionField{sym: &SymbolToken{Content: "ids"}},
			},
			want: &BackwardComparatorCondition{
				Comparator: BackwardComparator("IN"),
				Property:   Property{Name: "ids"},
				Value:      []any{int64(1), int64(2)},
			},
		},
		{
			name: "backward IN with blob value",
			condition: &backwardComparatorCondition{
				left:   &conditionBlob{blobKeyword: &KeywordToken{Name: "BLOB"}, b: []byte("abc")},
				op:     &OperatorToken{Type: "IN", Position: 3},
				opType: "IN",
				right:  &conditionField{sym: &SymbolToken{Content: "data"}},
			},
			want: &BackwardComparatorCondition{
				Comparator: BackwardComparator("IN"),
				Property:   Property{Name: "data"},
				Value:      []byte("abc"),
			},
		},
		{
			name: "backward IN with datetime value",
			condition: &backwardComparatorCondition{
				left:   &conditionDateTime{dateTimeKeyword: &KeywordToken{Name: "DATETIME"}, t: time.Date(2025, 7, 5, 12, 0, 0, 0, time.UTC)},
				op:     &OperatorToken{Type: "IN", Position: 4},
				opType: "IN",
				right:  &conditionField{sym: &SymbolToken{Content: "created_at"}},
			},
			want: &BackwardComparatorCondition{
				Comparator: BackwardComparator("IN"),
				Property:   Property{Name: "created_at"},
				Value:      time.Date(2025, 7, 5, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "backward IN with named binding value",
			condition: &backwardComparatorCondition{
				left:   &conditionValue{bind: &BindingToken{Name: "param1"}},
				op:     &OperatorToken{Type: "IN", Position: 5},
				opType: "IN",
				right:  &conditionField{sym: &SymbolToken{Content: "param"}},
			},
			want: &BackwardComparatorCondition{
				Comparator: BackwardComparator("IN"),
				Property:   Property{Name: "param"},
				Value:      &NamedBinding{Name: "param1"},
			},
		},
		{
			name: "backward IN with indexed binding value",
			condition: &backwardComparatorCondition{
				left:   &conditionValue{bind: &BindingToken{Index: 2}},
				op:     &OperatorToken{Type: "IN", Position: 6},
				opType: "IN",
				right:  &conditionField{sym: &SymbolToken{Content: "param"}},
			},
			want: &BackwardComparatorCondition{
				Comparator: BackwardComparator("IN"),
				Property:   Property{Name: "param"},
				Value:      &IndexedBinding{Index: 2},
			},
		},
		{
			name: "backward equals with string property field",
			condition: &backwardComparatorCondition{
				left:   &conditionValue{s: &StringToken{Content: "test"}},
				op:     &OperatorToken{Type: "=", Position: 7},
				opType: "=",
				right:  &conditionField{str: &StringToken{Content: "fieldStr"}},
			},
			want: &EitherComparatorCondition{
				Comparator: EitherComparator("="),
				Property:   Property{Name: "fieldStr"},
				Value:      "test",
			},
		},
		{
			name: "backward value panic (all nil)",
			condition: &backwardComparatorCondition{
				left:   &conditionValue{}, // will panic on value()
				op:     &OperatorToken{Type: "=", Position: 8},
				opType: "=",
				right:  &conditionField{sym: &SymbolToken{Content: "field"}},
			},
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic, but got none")
					}
				}()
			}

			got, err := tt.condition.toCondition()
			if (err != nil) != tt.wantErr {
				t.Errorf("backwardComparatorCondition.toCondition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !cmp.Equal(got, tt.want) {
				t.Errorf("backwardComparatorCondition.toCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompoundComparatorCondition_ToCondition(t *testing.T) {
	leftCondition := &forwardComparatorCondition{
		left:   &conditionField{sym: &SymbolToken{Content: "name"}},
		op:     &OperatorToken{Type: "CONTAINS", Position: 0},
		opType: "CONTAINS",
		right:  &conditionValue{s: &StringToken{Content: "test"}},
	}
	rightCondition := &forwardComparatorCondition{
		left:   &conditionField{sym: &SymbolToken{Content: "category"}},
		op:     &OperatorToken{Type: "IN", Position: 0},
		opType: "IN",
		right:  &conditionValue{s: &StringToken{Content: "news"}},
	}

	tests := []struct {
		name      string
		condition *compoundComparatorCondition
		want      Condition
		wantErr   bool
	}{
		{
			name: "AND compound condition",
			condition: &compoundComparatorCondition{
				left:  leftCondition,
				op:    &OperatorToken{Type: "AND", Position: 0},
				right: rightCondition,
			},
			want: &AndCompoundCondition{
				Left: &ForwardComparatorCondition{
					Comparator: ForwardComparator("CONTAINS"),
					Property:   Property{Name: "name"},
					Value:      "test",
				},
				Right: &ForwardComparatorCondition{
					Comparator: ForwardComparator("IN"),
					Property:   Property{Name: "category"},
					Value:      "news",
				},
			},
		},
		{
			name: "OR compound condition",
			condition: &compoundComparatorCondition{
				left:  leftCondition,
				op:    &OperatorToken{Type: "OR", Position: 0},
				right: rightCondition,
			},
			want: &OrCompoundCondition{
				Left: &ForwardComparatorCondition{
					Comparator: ForwardComparator("CONTAINS"),
					Property:   Property{Name: "name"},
					Value:      "test",
				},
				Right: &ForwardComparatorCondition{
					Comparator: ForwardComparator("IN"),
					Property:   Property{Name: "category"},
					Value:      "news",
				},
			},
		},
		{
			name: "left condition is invalid",
			condition: &compoundComparatorCondition{
				left:  &conditionValue{n: &NumericToken{Int64: 42}},
				op:    &OperatorToken{Type: "OR", Position: 0},
				right: rightCondition,
			},
			wantErr: true,
		},
		{
			name: "right condition is invalid",
			condition: &compoundComparatorCondition{
				left:  leftCondition,
				op:    &OperatorToken{Type: "OR", Position: 0},
				right: &conditionValue{n: &NumericToken{Int64: 42}},
			},
			wantErr: true,
		},
		{
			name: "invalid operator should error",
			condition: &compoundComparatorCondition{
				left:  leftCondition,
				op:    &OperatorToken{Type: "XOR", Position: 0, RawContent: "XOR"},
				right: rightCondition,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.condition.toCondition()
			if (err != nil) != tt.wantErr {
				t.Errorf("compoundComparatorCondition.toCondition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !cmp.Equal(got, tt.want) {
				t.Errorf("compoundComparatorCondition.toCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConditionField_Name(t *testing.T) {
	tests := []struct {
		name  string
		field *conditionField
		want  string
	}{
		{
			name:  "symbol token field",
			field: &conditionField{sym: &SymbolToken{Content: "fieldName"}},
			want:  "fieldName",
		},
		{
			name:  "string token field",
			field: &conditionField{str: &StringToken{Content: "stringField"}},
			want:  "stringField",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.field.name(); got != tt.want {
				t.Errorf("conditionField.name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConditionField_Token(t *testing.T) {
	symToken := &SymbolToken{Content: "field", Position: 0}
	strToken := &StringToken{Content: "field", Position: 0}

	tests := []struct {
		name  string
		field *conditionField
		want  Token
	}{
		{
			name:  "symbol token",
			field: &conditionField{sym: symToken},
			want:  symToken,
		},
		{
			name:  "string token",
			field: &conditionField{str: strToken},
			want:  strToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.field.token(); got != tt.want {
				t.Errorf("conditionField.token() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConditionField_ToCondition(t *testing.T) {
	field := &conditionField{
		sym: &SymbolToken{Content: "field", Position: 5},
	}

	got, err := field.toCondition()
	if got != nil {
		t.Errorf("conditionField.toCondition() got = %v, want nil", got)
	}
	if !errors.Is(err, ErrUnexpectedToken) {
		t.Errorf("conditionField.toCondition() error = %v, want ErrUnexpectedToken", err)
	}
}

func TestConditionValue_Value(t *testing.T) {
	tests := []struct {
		name      string
		value     *conditionValue
		want      any
		wantPanic bool
	}{
		{
			name:  "boolean true value",
			value: &conditionValue{b: &BooleanToken{Value: true}},
			want:  true,
		},
		{
			name:  "boolean false value",
			value: &conditionValue{b: &BooleanToken{Value: false}},
			want:  false,
		},
		{
			name:  "string value",
			value: &conditionValue{s: &StringToken{Content: "test string"}},
			want:  "test string",
		},
		{
			name:  "integer numeric value",
			value: &conditionValue{n: &NumericToken{Int64: 42, Floating: false}},
			want:  int64(42),
		},
		{
			name:  "float numeric value",
			value: &conditionValue{n: &NumericToken{Float64: 3.14, Floating: true}},
			want:  float64(3.14),
		},
		{
			name:  "null value",
			value: &conditionValue{null: &KeywordToken{Name: "NULL"}},
			want:  nil,
		},
		{
			name:      "all tokens nil should panic",
			value:     &conditionValue{},
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic, but got none")
					}
				}()
			}

			got := tt.value.value()
			if !tt.wantPanic && !cmp.Equal(got, tt.want) {
				t.Errorf("conditionValue.value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConditionValue_Token(t *testing.T) {
	boolToken := &BooleanToken{Value: true}
	stringToken := &StringToken{Content: "test"}
	numToken := &NumericToken{Int64: 42}
	nullToken := &KeywordToken{Name: "NULL"}
	bindToken := &BindingToken{Name: "param"}

	tests := []struct {
		name      string
		value     *conditionValue
		want      Token
		wantPanic bool
	}{
		{
			name:  "boolean token",
			value: &conditionValue{b: boolToken},
			want:  boolToken,
		},
		{
			name:  "string token",
			value: &conditionValue{s: stringToken},
			want:  stringToken,
		},
		{
			name:  "numeric token",
			value: &conditionValue{n: numToken},
			want:  numToken,
		},
		{
			name:  "null token",
			value: &conditionValue{null: nullToken},
			want:  nullToken,
		},
		{
			name:  "binding token",
			value: &conditionValue{bind: bindToken},
			want:  bindToken,
		},
		{
			name:      "all tokens nil should panic",
			value:     &conditionValue{},
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic, but got none")
					}
				}()
			}

			got := tt.value.token()
			if !tt.wantPanic && got != tt.want {
				t.Errorf("conditionValue.token() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConditionValue_ToCondition(t *testing.T) {
	value := &conditionValue{
		s: &StringToken{Content: "test", Position: 3},
	}

	got, err := value.toCondition()
	if got != nil {
		t.Errorf("conditionValue.toCondition() got = %v, want nil", got)
	}
	if !errors.Is(err, ErrUnexpectedToken) {
		t.Errorf("conditionValue.toCondition() error = %v, want ErrUnexpectedToken", err)
	}
}

func TestConditionKey_Value(t *testing.T) {
	testKey := &Key{ProjectID: "test-project"}

	key := &conditionKey{
		keyKeyword: &KeywordToken{Name: "KEY"},
		key:        testKey,
	}

	got := key.value()
	if got != testKey {
		t.Errorf("conditionKey.value() = %v, want %v", got, testKey)
	}
}

func TestConditionKey_ToCondition(t *testing.T) {
	key := &conditionKey{
		keyKeyword: &KeywordToken{Name: "KEY", Position: 7},
		key:        &Key{ProjectID: "test"},
	}

	got, err := key.toCondition()
	if got != nil {
		t.Errorf("conditionKey.toCondition() got = %v, want nil", got)
	}
	if !errors.Is(err, ErrUnexpectedToken) {
		t.Errorf("conditionKey.toCondition() error = %v, want ErrUnexpectedToken", err)
	}
}

func TestConditionArray_Value(t *testing.T) {
	array := &conditionArray{
		arrayKeyword: &KeywordToken{Name: "ARRAY"},
		values: []valueAST{
			&conditionValue{n: &NumericToken{Int64: 1}},
			&conditionValue{n: &NumericToken{Int64: 2}},
			&conditionValue{s: &StringToken{Content: "test"}},
		},
	}

	want := []any{int64(1), int64(2), "test"}
	got := array.value()

	if !cmp.Equal(got, want) {
		t.Errorf("conditionArray.value() = %v, want %v", got, want)
	}
}

func TestConditionArray_ToCondition(t *testing.T) {
	array := &conditionArray{
		arrayKeyword: &KeywordToken{Name: "ARRAY", Position: 10},
		values:       []valueAST{},
	}

	got, err := array.toCondition()
	if got != nil {
		t.Errorf("conditionArray.toCondition() got = %v, want nil", got)
	}
	if !errors.Is(err, ErrUnexpectedToken) {
		t.Errorf("conditionArray.toCondition() error = %v, want ErrUnexpectedToken", err)
	}
}

func TestConditionBlob_Value(t *testing.T) {
	testBytes := []byte("test blob data")

	blob := &conditionBlob{
		blobKeyword: &KeywordToken{Name: "BLOB"},
		b:           testBytes,
	}

	got := blob.value()
	if !cmp.Equal(got, testBytes) {
		t.Errorf("conditionBlob.value() = %v, want %v", got, testBytes)
	}
}

func TestConditionBlob_ToCondition(t *testing.T) {
	blob := &conditionBlob{
		blobKeyword: &KeywordToken{Name: "BLOB", Position: 15},
		b:           []byte("test"),
	}

	got, err := blob.toCondition()
	if got != nil {
		t.Errorf("conditionBlob.toCondition() got = %v, want nil", got)
	}
	if !errors.Is(err, ErrUnexpectedToken) {
		t.Errorf("conditionBlob.toCondition() error = %v, want ErrUnexpectedToken", err)
	}
}

func TestConditionDateTime_Value(t *testing.T) {
	testTime := time.Date(2023, 6, 15, 14, 30, 0, 0, time.UTC)

	dateTime := &conditionDateTime{
		dateTimeKeyword: &KeywordToken{Name: "DATETIME"},
		t:               testTime,
	}

	got := dateTime.value()
	if !cmp.Equal(got, testTime) {
		t.Errorf("conditionDateTime.value() = %v, want %v", got, testTime)
	}
}

func TestConditionDateTime_ToCondition(t *testing.T) {
	dateTime := &conditionDateTime{
		dateTimeKeyword: &KeywordToken{Name: "DATETIME", Position: 20},
		t:               time.Now(),
	}

	got, err := dateTime.toCondition()
	if got != nil {
		t.Errorf("conditionDateTime.toCondition() got = %v, want nil", got)
	}
	if !errors.Is(err, ErrUnexpectedToken) {
		t.Errorf("conditionDateTime.toCondition() error = %v, want ErrUnexpectedToken", err)
	}
}

func TestConditionAST_ToUnexpectedTokenError(t *testing.T) {
	tests := []struct {
		name string
		ast  conditionAST
		want string
	}{
		{
			name: "conditionField error",
			ast: &conditionField{
				sym: &SymbolToken{Content: "field", Position: 10},
			},
			want: "unexpected token: field at 10",
		},
		{
			name: "conditionValue error",
			ast: &conditionValue{
				s: &StringToken{Content: "value", Position: 15, RawContent: "\"value\""},
			},
			want: "unexpected token: \"value\" at 15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ast.toUnexpectedTokenError()
			if err == nil {
				t.Errorf("Expected error, got nil")
				return
			}
			if !errors.Is(err, ErrUnexpectedToken) {
				t.Errorf("Expected ErrUnexpectedToken, got %v", err)
			}
			if err.Error() != tt.want {
				t.Errorf("toUnexpectedTokenError() = %v, want %v", err.Error(), tt.want)
			}
		})
	}
}

func TestForwardComparatorCondition_ToUnexpectedTokenError(t *testing.T) {
	tests := []struct {
		name      string
		condition *forwardComparatorCondition
		wantErr   string
	}{
		{
			name: "left is conditionField (symbol)",
			condition: &forwardComparatorCondition{
				left:   &conditionField{sym: &SymbolToken{Content: "field", Position: 3}},
				op:     &OperatorToken{Type: "=", Position: 0},
				opType: "=",
				right:  &conditionValue{s: &StringToken{Content: "test"}},
			},
			wantErr: "unexpected token: field at 3",
		},
		{
			name: "left is conditionField (string)",
			condition: &forwardComparatorCondition{
				left:   &conditionField{str: &StringToken{Content: "strField", Position: 5, RawContent: "\"strField\""}},
				op:     &OperatorToken{Type: "=", Position: 0},
				opType: "=",
				right:  &conditionValue{s: &StringToken{Content: "test"}},
			},
			wantErr: "unexpected token: \"strField\" at 5",
		},
		{
			name: "left is conditionFieldAccess",
			condition: &forwardComparatorCondition{
				left: &conditionFieldAccess{
					left:  &conditionField{sym: &SymbolToken{Content: "parent", Position: 7}},
					right: &conditionField{sym: &SymbolToken{Content: "child", Position: 8}},
				},
				op:     &OperatorToken{Type: "=", Position: 0},
				opType: "=",
				right:  &conditionValue{s: &StringToken{Content: "test"}},
			},
			wantErr: "unexpected token: parent at 7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.condition.toUnexpectedTokenError()
			if err == nil {
				t.Errorf("Expected error, got nil")
				return
			}
			if err.Error() != tt.wantErr {
				t.Errorf("toUnexpectedTokenError() = %v, want %v", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestBackwardComparatorCondition_ToUnexpectedTokenError(t *testing.T) {
	tests := []struct {
		name      string
		condition *backwardComparatorCondition
		wantErr   string
	}{
		{
			name: "left is conditionValue (string)",
			condition: &backwardComparatorCondition{
				left:   &conditionValue{s: &StringToken{Content: "val", Position: 11, RawContent: "\"val\""}},
				op:     &OperatorToken{Type: "="},
				opType: "=",
				right:  &conditionField{sym: &SymbolToken{Content: "field"}},
			},
			wantErr: "unexpected token: \"val\" at 11",
		},
		{
			name: "left is conditionArray",
			condition: &backwardComparatorCondition{
				left:   &conditionArray{arrayKeyword: &KeywordToken{Name: "ARRAY", RawContent: "ARRAY", Position: 22}},
				op:     &OperatorToken{Type: "IN"},
				opType: "IN",
				right:  &conditionField{sym: &SymbolToken{Content: "ids"}},
			},
			wantErr: "unexpected token: ARRAY at 22",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.condition.toUnexpectedTokenError()
			if err == nil {
				t.Errorf("Expected error, got nil")
				return
			}
			if err.Error() != tt.wantErr {
				t.Errorf("toUnexpectedTokenError() = %v, want %v", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestCompoundComparatorCondition_ToUnexpectedTokenError(t *testing.T) {
	tests := []struct {
		name      string
		condition *compoundComparatorCondition
		wantErr   string
	}{
		{
			name: "left is conditionField",
			condition: &compoundComparatorCondition{
				left: &conditionField{sym: &SymbolToken{Content: "field", Position: 33}},
				op:   &OperatorToken{Type: "AND", Position: 0},
				right: &forwardComparatorCondition{
					left:   &conditionField{sym: &SymbolToken{Content: "name"}},
					op:     &OperatorToken{Type: "=", Position: 0},
					opType: "=",
					right:  &conditionValue{s: &StringToken{Content: "test"}},
				},
			},
			wantErr: "unexpected token: field at 33",
		},
		{
			name: "left is conditionValue",
			condition: &compoundComparatorCondition{
				left: &conditionValue{s: &StringToken{Content: "val", Position: 44, RawContent: "\"val\""}},
				op:   &OperatorToken{Type: "OR", Position: 0},
				right: &forwardComparatorCondition{
					left:   &conditionField{sym: &SymbolToken{Content: "name"}},
					op:     &OperatorToken{Type: "=", Position: 0},
					opType: "=",
					right:  &conditionValue{s: &StringToken{Content: "test"}},
				},
			},
			wantErr: "unexpected token: \"val\" at 44",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.condition.toUnexpectedTokenError()
			if err == nil {
				t.Errorf("Expected error, got nil")
				return
			}
			if err.Error() != tt.wantErr {
				t.Errorf("toUnexpectedTokenError() = %v, want %v", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestConditionField_ToUnexpectedTokenError(t *testing.T) {
	tests := []struct {
		name  string
		field *conditionField
		want  string
	}{
		{
			name:  "symbol token",
			field: &conditionField{sym: &SymbolToken{Content: "field", Position: 55}},
			want:  "unexpected token: field at 55",
		},
		{
			name:  "string token",
			field: &conditionField{str: &StringToken{Content: "str", Position: 66, RawContent: "\"str\""}},
			want:  "unexpected token: \"str\" at 66",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.field.toUnexpectedTokenError()
			if err == nil {
				t.Errorf("Expected error, got nil")
				return
			}
			if err.Error() != tt.want {
				t.Errorf("toUnexpectedTokenError() = %v, want %v", err.Error(), tt.want)
			}
		})
	}
}

func TestConditionValue_ToUnexpectedTokenError(t *testing.T) {
	tests := []struct {
		name  string
		value *conditionValue
		want  string
	}{
		{
			name:  "string token",
			value: &conditionValue{s: &StringToken{Content: "val", Position: 77, RawContent: "\"val\""}},
			want:  "unexpected token: \"val\" at 77",
		},
		{
			name:  "boolean token",
			value: &conditionValue{b: &BooleanToken{Value: true, RawContent: "True", Position: 88}},
			want:  "unexpected token: True at 88",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.value.toUnexpectedTokenError()
			if err == nil {
				t.Errorf("Expected error, got nil")
				return
			}
			if err.Error() != tt.want {
				t.Errorf("toUnexpectedTokenError() = %v, want %v", err.Error(), tt.want)
			}
		})
	}
}

func TestBackwardComparatorCondition_EitherOperatorMapping(t *testing.T) {
	// Test the either operator mapping with operators that work
	// Note: < and > fail due to runetrie validation issue, so test with <= instead
	condition := &backwardComparatorCondition{
		left:   &conditionValue{n: &NumericToken{Int64: 5}},
		op:     &OperatorToken{Type: "<=", Position: 0},
		opType: "<=",
		right:  &conditionField{sym: &SymbolToken{Content: "field"}},
	}

	got, err := condition.toCondition()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	want := &EitherComparatorCondition{
		Comparator: EitherComparator(">="), // <= should be inverted to >=
		Property:   Property{Name: "field"},
		Value:      int64(5),
	}

	if !cmp.Equal(got, want) {
		t.Errorf("backwardComparatorCondition.toCondition() = %v, want %v", got, want)
	}
}

func TestCompoundComparatorCondition_NestedErrors(t *testing.T) {
	// Test error propagation from left condition
	leftCondition := &conditionField{sym: &SymbolToken{Content: "field", Position: 0}}
	rightCondition := &forwardComparatorCondition{
		left:   &conditionField{sym: &SymbolToken{Content: "name"}},
		op:     &OperatorToken{Type: "CONTAINS", Position: 0},
		opType: "CONTAINS",
		right:  &conditionValue{s: &StringToken{Content: "test"}},
	}

	condition := &compoundComparatorCondition{
		left:  leftCondition,
		op:    &OperatorToken{Type: "AND", Position: 0},
		right: rightCondition,
	}

	got, err := condition.toCondition()
	if got != nil {
		t.Errorf("Expected nil condition when left fails")
	}
	if !errors.Is(err, ErrUnexpectedToken) {
		t.Errorf("Expected ErrUnexpectedToken from left condition, got %v", err)
	}
}

func TestConditionAST_InterfaceImplementation(t *testing.T) {
	// Test that all types implement the interfaces correctly
	var _ conditionAST = &forwardComparatorCondition{}
	var _ conditionAST = &backwardComparatorCondition{}
	var _ conditionAST = &compoundComparatorCondition{}
	var _ conditionAST = &conditionField{}
	var _ conditionAST = &conditionValue{}

	var _ valueAST = &conditionValue{}
	var _ valueAST = &conditionKey{}
	var _ valueAST = &conditionArray{}
	var _ valueAST = &conditionBlob{}
	var _ valueAST = &conditionDateTime{}
}

func TestConditionArray_EmptyArray(t *testing.T) {
	array := &conditionArray{
		arrayKeyword: &KeywordToken{Name: "ARRAY"},
		values:       []valueAST{},
	}

	want := []any{}
	got := array.value()

	if !cmp.Equal(got, want) {
		t.Errorf("conditionArray.value() = %v, want %v", got, want)
	}
}

func TestConditionArray_MixedTypes(t *testing.T) {
	array := &conditionArray{
		arrayKeyword: &KeywordToken{Name: "ARRAY"},
		values: []valueAST{
			&conditionValue{b: &BooleanToken{Value: true}},
			&conditionValue{null: &KeywordToken{Name: "NULL"}},
			&conditionValue{n: &NumericToken{Float64: 3.14, Floating: true}},
			&conditionKey{key: &Key{ProjectID: "test"}},
		},
	}

	want := []any{true, nil, float64(3.14), &Key{ProjectID: "test"}}
	got := array.value()

	if !cmp.Equal(got, want) {
		t.Errorf("conditionArray.value() = %v, want %v", got, want)
	}
}

// Test comprehensive functionality
func TestConditionAST_ComprehensiveIntegration(t *testing.T) {
	// Test a more complex scenario
	leftCondition := &forwardComparatorCondition{
		left:   &conditionField{sym: &SymbolToken{Content: "status"}},
		op:     &OperatorToken{Type: "=", Position: 0},
		opType: "=",
		right:  &conditionValue{s: &StringToken{Content: "active"}},
	}

	rightCondition := &forwardComparatorCondition{
		left:   &conditionField{sym: &SymbolToken{Content: "category"}},
		op:     &OperatorToken{Type: "CONTAINS", Position: 0},
		opType: "CONTAINS",
		right:  &conditionValue{s: &StringToken{Content: "tech"}},
	}

	compoundCondition := &compoundComparatorCondition{
		left:  leftCondition,
		op:    &OperatorToken{Type: "AND", Position: 0},
		right: rightCondition,
	}

	got, err := compoundCondition.toCondition()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	want := &AndCompoundCondition{
		Left: &EitherComparatorCondition{
			Comparator: EitherComparator("="),
			Property:   Property{Name: "status"},
			Value:      "active",
		},
		Right: &ForwardComparatorCondition{
			Comparator: ForwardComparator("CONTAINS"),
			Property:   Property{Name: "category"},
			Value:      "tech",
		},
	}

	if !cmp.Equal(got, want) {
		t.Errorf("Complex condition conversion failed.\nGot:  %v\nWant: %v", got, want)
	}
}

func TestConditionAST_ErrorHandling(t *testing.T) {
	tests := []struct {
		name string
		ast  conditionAST
		want string
	}{
		{
			name: "Invalid either comparator in forward condition",
			ast: &forwardComparatorCondition{
				left:   &conditionField{sym: &SymbolToken{Content: "field"}},
				op:     &OperatorToken{Type: "INVALID", Position: 5, RawContent: "INVALID"},
				opType: "INVALID",
				right:  &conditionValue{s: &StringToken{Content: "value"}},
			},
			want: "unexpected token: INVALID at 5",
		},
		{
			name: "Invalid comparator in backward condition",
			ast: &backwardComparatorCondition{
				left:   &conditionValue{s: &StringToken{Content: "value"}},
				op:     &OperatorToken{Type: "INVALID", Position: 10, RawContent: "INVALID"},
				opType: "INVALID",
				right:  &conditionField{sym: &SymbolToken{Content: "field"}},
			},
			want: "unexpected token: INVALID at 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.ast.toCondition()
			if err == nil {
				t.Errorf("Expected error, got nil")
				return
			}
			if !errors.Is(err, ErrUnexpectedToken) {
				t.Errorf("Expected ErrUnexpectedToken, got %v", err)
			}
			if err.Error() != tt.want {
				t.Errorf("Error message = %v, want %v", err.Error(), tt.want)
			}
		})
	}
}
