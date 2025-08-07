package gqlparser_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/karupanerura/gqlparser"
)

func TestConditionBind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		resolver  *gqlparser.BindingResolver
		condition gqlparser.Condition
		want      gqlparser.Condition
		wantErr   bool
	}{
		{
			name:     "ExpandIndexedBinding",
			resolver: &gqlparser.BindingResolver{Indexed: []any{int64(10), int64(20)}},
			condition: &gqlparser.OrCompoundCondition{
				Left: &gqlparser.EitherComparatorCondition{
					Comparator: gqlparser.GreaterThanEitherComparator,
					Property:   gqlparser.Property{Name: "a"},
					Value:      &gqlparser.IndexedBinding{Index: 1},
				},
				Right: &gqlparser.ForwardComparatorCondition{
					Comparator: gqlparser.ContainsForwardComparator,
					Property:   gqlparser.Property{Name: "a"},
					Value:      &gqlparser.IndexedBinding{Index: 2},
				},
			},
			want: &gqlparser.OrCompoundCondition{
				Left: &gqlparser.EitherComparatorCondition{
					Comparator: gqlparser.GreaterThanEitherComparator,
					Property:   gqlparser.Property{Name: "a"},
					Value:      int64(10),
				},
				Right: &gqlparser.ForwardComparatorCondition{
					Comparator: gqlparser.ContainsForwardComparator,
					Property:   gqlparser.Property{Name: "a"},
					Value:      int64(20),
				},
			},
			wantErr: false,
		},
		{
			name: "ExpandNamedBinding",
			resolver: &gqlparser.BindingResolver{
				Named: map[string]any{
					"ancestor": &gqlparser.Key{
						Path: []*gqlparser.KeyPath{
							{Kind: "Parent", Name: "foo"},
						},
					},
					"list": []any{int64(10), int64(20)},
				},
			},
			condition: &gqlparser.AndCompoundCondition{
				Left: &gqlparser.ForwardComparatorCondition{
					Comparator: gqlparser.HasAncestorForwardComparator,
					Property:   gqlparser.Property{Name: "__key__"},
					Value:      &gqlparser.NamedBinding{Name: "ancestor"},
				},
				Right: &gqlparser.OrCompoundCondition{
					Left: &gqlparser.IsNullCondition{Property: gqlparser.Property{Name: "a"}},
					Right: &gqlparser.BackwardComparatorCondition{
						Comparator: gqlparser.InBackwardComparator,
						Property:   gqlparser.Property{Name: "a"},
						Value:      &gqlparser.NamedBinding{Name: "list"},
					},
				},
			},
			want: &gqlparser.AndCompoundCondition{
				Left: &gqlparser.ForwardComparatorCondition{
					Comparator: gqlparser.HasAncestorForwardComparator,
					Property:   gqlparser.Property{Name: "__key__"},
					Value: &gqlparser.Key{
						Path: []*gqlparser.KeyPath{
							{Kind: "Parent", Name: "foo"},
						},
					},
				},
				Right: &gqlparser.OrCompoundCondition{
					Left: &gqlparser.IsNullCondition{Property: gqlparser.Property{Name: "a"}},
					Right: &gqlparser.BackwardComparatorCondition{
						Comparator: gqlparser.InBackwardComparator,
						Property:   gqlparser.Property{Name: "a"},
						Value:      []any{int64(10), int64(20)},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.condition.Bind(tt.resolver)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bind() error = %+v, wantErr %+v", err, tt.wantErr)
				return
			}
			if df := cmp.Diff(tt.want, tt.condition); df != "" {
				t.Errorf("ParseQuery() = %+v, want %+v, diff = %s", tt.condition, tt.want, df)
			}
		})
	}
}

func TestPropertyString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		property gqlparser.Property
		want     string
	}{
		{
			name:     "SimpleProperty",
			property: gqlparser.Property{Name: "name"},
			want:     "name",
		},
		{
			name: "NestedProperty",
			property: gqlparser.Property{
				Name: "user",
				Child: &gqlparser.Property{Name: "email"},
			},
			want: "user.email",
		},
		{
			name: "DeeplyNestedProperty",
			property: gqlparser.Property{
				Name: "user",
				Child: &gqlparser.Property{
					Name: "profile",
					Child: &gqlparser.Property{Name: "address"},
				},
			},
			want: "user.profile.address",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.property.String()
			if got != tt.want {
				t.Errorf("Property.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
