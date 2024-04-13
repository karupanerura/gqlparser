package gqlparser

import (
	"errors"
	"fmt"
)

var ErrBindValue = errors.New("no bind value")

type BindingResolver struct {
	Indexed []any
	Named   map[string]any
}

func (r *BindingResolver) Resolve(value BindingVariable) (any, error) {
	return value.resolveBy(r)
}

func (r *BindingResolver) getNamed(name string) (any, error) {
	if r.Named == nil {
		return nil, fmt.Errorf("%w: name=%s", ErrBindValue, name)
	}
	if v, ok := r.Named[name]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("%w: name=%s", ErrBindValue, name)
}

func (r *BindingResolver) getIndexed(index int64) (any, error) {
	if r.Indexed == nil {
		return nil, fmt.Errorf("%w: index=%d", ErrBindValue, index)
	}
	if index <= int64(len(r.Indexed)) {
		return r.Indexed[index-1], nil
	}
	return nil, fmt.Errorf("%w: index=%d", ErrBindValue, index)
}

type BindingVariable interface {
	resolveBy(resolver *BindingResolver) (any, error)
}

type NamedBinding struct {
	Name string
}

func (b *NamedBinding) resolveBy(resolver *BindingResolver) (any, error) {
	return resolver.getNamed(b.Name)
}

type IndexedBinding struct {
	Index int64
}

func (b *IndexedBinding) resolveBy(resolver *BindingResolver) (any, error) {
	return resolver.getIndexed(b.Index)
}
