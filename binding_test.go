package gqlparser_test

import (
	"errors"
	"testing"

	"github.com/karupanerura/gqlparser"
)

func TestBindingResolver_Resolve_NamedBinding(t *testing.T) {
	resolver := &gqlparser.BindingResolver{
		Named: map[string]any{
			"name":  "John",
			"age":   30,
			"email": "john@example.com",
		},
	}

	tests := []struct {
		name     string
		binding  gqlparser.BindingVariable
		expected any
		wantErr  bool
	}{
		{
			name:     "existing named binding",
			binding:  &gqlparser.NamedBinding{Name: "name"},
			expected: "John",
			wantErr:  false,
		},
		{
			name:     "existing named binding with number",
			binding:  &gqlparser.NamedBinding{Name: "age"},
			expected: 30,
			wantErr:  false,
		},
		{
			name:     "non-existing named binding",
			binding:  &gqlparser.NamedBinding{Name: "nonexistent"},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.Resolve(tt.binding)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("Resolve() = %v, expected %v", result, tt.expected)
			}
			if tt.wantErr && !errors.Is(err, gqlparser.ErrBindValue) {
				t.Errorf("Resolve() error should be ErrBindValue, got %v", err)
			}
		})
	}
}

func TestBindingResolver_Resolve_IndexedBinding(t *testing.T) {
	resolver := &gqlparser.BindingResolver{
		Indexed: []any{"first", "second", 42, true},
	}

	tests := []struct {
		name     string
		binding  gqlparser.BindingVariable
		expected any
		wantErr  bool
	}{
		{
			name:     "valid index 1",
			binding:  &gqlparser.IndexedBinding{Index: 1},
			expected: "first",
			wantErr:  false,
		},
		{
			name:     "valid index 3",
			binding:  &gqlparser.IndexedBinding{Index: 3},
			expected: 42,
			wantErr:  false,
		},
		{
			name:     "valid index 4",
			binding:  &gqlparser.IndexedBinding{Index: 4},
			expected: true,
			wantErr:  false,
		},
		{
			name:     "out of bounds index",
			binding:  &gqlparser.IndexedBinding{Index: 10},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.Resolve(tt.binding)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("Resolve() = %v, expected %v", result, tt.expected)
			}
			if tt.wantErr && !errors.Is(err, gqlparser.ErrBindValue) {
				t.Errorf("Resolve() error should be ErrBindValue, got %v", err)
			}
		})
	}
}

func TestBindingResolver_Resolve_NilMaps(t *testing.T) {
	resolver := &gqlparser.BindingResolver{
		Named:   nil,
		Indexed: nil,
	}

	t.Run("nil named map", func(t *testing.T) {
		binding := &gqlparser.NamedBinding{Name: "test"}
		result, err := resolver.Resolve(binding)
		if err == nil {
			t.Error("Expected error for nil Named map")
		}
		if result != nil {
			t.Errorf("Expected nil result, got %v", result)
		}
		if !errors.Is(err, gqlparser.ErrBindValue) {
			t.Errorf("Expected ErrBindValue, got %v", err)
		}
	})

	t.Run("nil indexed slice", func(t *testing.T) {
		binding := &gqlparser.IndexedBinding{Index: 0}
		result, err := resolver.Resolve(binding)
		if err == nil {
			t.Error("Expected error for nil Indexed slice")
		}
		if result != nil {
			t.Errorf("Expected nil result, got %v", result)
		}
		if !errors.Is(err, gqlparser.ErrBindValue) {
			t.Errorf("Expected ErrBindValue, got %v", err)
		}
	})
}

func TestBindingResolver_Resolve_EmptyCollections(t *testing.T) {
	resolver := &gqlparser.BindingResolver{
		Named:   make(map[string]any),
		Indexed: make([]any, 0),
	}

	t.Run("empty named map", func(t *testing.T) {
		binding := &gqlparser.NamedBinding{Name: "test"}
		result, err := resolver.Resolve(binding)
		if err == nil {
			t.Error("Expected error for missing key in empty Named map")
		}
		if result != nil {
			t.Errorf("Expected nil result, got %v", result)
		}
	})

	t.Run("empty indexed slice", func(t *testing.T) {
		binding := &gqlparser.IndexedBinding{Index: 1}
		result, err := resolver.Resolve(binding)
		if err == nil {
			t.Error("Expected error for index in empty Indexed slice")
		}
		if result != nil {
			t.Errorf("Expected nil result, got %v", result)
		}
	})
}
