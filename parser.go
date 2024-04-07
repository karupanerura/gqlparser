package gqlparser

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"
)

var (
	ErrNoTokens        = errors.New("no tokens")
	ErrUnexpectedToken = errors.New("unexpected token")
)

func ParseAggregationQuery(ts TokenSource) (*AggregationQuery, error) {
	var query AggregationQuery
	acceptor := acceptAggregationQuery(&query)
	// if err := acceptor.accept(&debugTokenSource{ts}); err != nil {
	if err := acceptor.accept(ts); err != nil {
		return nil, err
	}
	if ts.Next() {
		tok, _ := ts.Read()
		return nil, fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, tok.GetContent(), tok.GetPosition())
	}
	return &query, nil
}

func acceptAggregationQuery(query *AggregationQuery) tokenAcceptor {
	return tokenAcceptors{
		skipWhitespaceToken,
		&conditionalTokenAcceptor{
			ifAccept: acceptKeyword("SELECT"),
			andThen: tokenAcceptors{
				acceptWhitespaceToken,
				acceptAggregations(&query.Aggregations),
				acceptWhitespaceToken,
				acceptKeyword("FROM"),
				acceptWhitespaceToken,
				acceptEitherToken(
					func(tok *SymbolToken) error {
						query.Kind = Kind(tok.Content)
						return nil
					},
					func(tok *StringToken) error {
						if tok.Quote != '`' {
							return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, tok.Content, tok.Position)
						}
						query.Kind = Kind(tok.Content)
						return nil
					},
				),
				&conditionalTokenAcceptor{
					ifAccept: tokenAcceptors{
						acceptWhitespaceToken,
						acceptKeyword("WHERE"),
					},
					andThen: tokenAcceptors{
						acceptWhitespaceToken,
						acceptCondition(&query.Where),
					},
					orElse: nopAcceptor,
				},
			},
			orElse: &conditionalTokenAcceptor{
				ifAccept: acceptKeyword("AGGREGATE"),
				andThen: tokenAcceptors{
					acceptWhitespaceToken,
					acceptAggregations(&query.Aggregations),
					acceptWhitespaceToken,
					acceptKeyword("OVER"),
					skipWhitespaceToken,
					acceptOperator("("),
					acceptQuery(&query.Query),
					acceptOperator(")"),
					skipWhitespaceToken,
				},
				orElse: tokenAcceptorFn(func(tr tokenReader) error {
					token, err := tr.Read()
					if err != nil {
						return err
					}
					return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, token.GetContent(), token.GetPosition())
				}),
			},
		},
	}
}

func acceptAggregations(aggregations *[]Aggregation) tokenAcceptor {
	var upTo int64
	var alias string
	var prop string
	return &conditionalTokenAcceptor{
		ifAccept: acceptKeyword("COUNT"),
		andThen: tokenAcceptors{
			skipWhitespaceToken,
			acceptOperator("("),
			skipWhitespaceToken,
			acceptWildcardToken,
			skipWhitespaceToken,
			acceptOperator(")"),
			skipWhitespaceToken,
			&conditionalTokenAcceptor{
				ifAccept: acceptKeyword("AS"),
				andThen: tokenAcceptors{
					acceptWhitespaceToken,
					acceptEitherToken(
						func(token *SymbolToken) error {
							alias = token.Content
							return nil
						},
						func(token *StringToken) error {
							if token.Quote != '`' {
								return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, token.GetContent(), token.GetPosition())
							}
							alias = token.Content
							return nil
						},
					),
					skipWhitespaceToken,
				},
				orElse: nopAcceptor,
			},
			&conditionalTokenAcceptor{
				ifAccept: acceptOperator(","),
				andThen: tokenAcceptors{
					skipWhitespaceToken,
					deferAcceptor(func() tokenAcceptor {
						*aggregations = append(*aggregations, &CountAggregation{Alias: alias})
						return acceptAggregations(aggregations)
					}),
				},
				orElse: deferAcceptor(func() tokenAcceptor {
					*aggregations = append(*aggregations, &CountAggregation{Alias: alias})
					return nopAcceptor
				}),
			},
		},
		orElse: &conditionalTokenAcceptor{
			ifAccept: acceptKeyword("COUNT_UP_TO"),
			andThen: tokenAcceptors{
				skipWhitespaceToken,
				acceptOperator("("),
				skipWhitespaceToken,
				acceptSingleToken(func(token *NumericToken) error {
					if token.Floating {
						return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, token.GetContent(), token.GetPosition())
					}
					upTo = token.Int64
					return nil
				}),
				skipWhitespaceToken,
				acceptOperator(")"),
				skipWhitespaceToken,
				&conditionalTokenAcceptor{
					ifAccept: acceptKeyword("AS"),
					andThen: tokenAcceptors{
						acceptWhitespaceToken,
						acceptEitherToken(
							func(token *SymbolToken) error {
								alias = token.Content
								return nil
							},
							func(token *StringToken) error {
								if token.Quote != '`' {
									return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, token.GetContent(), token.GetPosition())
								}
								alias = token.Content
								return nil
							},
						),
						skipWhitespaceToken,
					},
					orElse: nopAcceptor,
				},
				&conditionalTokenAcceptor{
					ifAccept: acceptOperator(","),
					andThen: tokenAcceptors{
						skipWhitespaceToken,
						deferAcceptor(func() tokenAcceptor {
							*aggregations = append(*aggregations, &CountUpToAggregation{Alias: alias, Limit: upTo})
							return acceptAggregations(aggregations)
						}),
					},
					orElse: deferAcceptor(func() tokenAcceptor {
						*aggregations = append(*aggregations, &CountUpToAggregation{Alias: alias, Limit: upTo})
						return nopAcceptor
					}),
				},
			},
			orElse: &conditionalTokenAcceptor{
				ifAccept: acceptKeyword("SUM"),
				andThen: tokenAcceptors{
					skipWhitespaceToken,
					acceptOperator("("),
					skipWhitespaceToken,
					acceptEitherToken(
						func(token *SymbolToken) error {
							prop = token.Content
							return nil
						},
						func(token *StringToken) error {
							if token.Quote != '`' {
								return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, token.GetContent(), token.GetPosition())
							}
							prop = token.Content
							return nil
						},
					),
					skipWhitespaceToken,
					acceptOperator(")"),
					skipWhitespaceToken,
					&conditionalTokenAcceptor{
						ifAccept: acceptKeyword("AS"),
						andThen: tokenAcceptors{
							acceptWhitespaceToken,
							acceptEitherToken(
								func(token *SymbolToken) error {
									alias = token.Content
									return nil
								},
								func(token *StringToken) error {
									if token.Quote != '`' {
										return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, token.GetContent(), token.GetPosition())
									}
									alias = token.Content
									return nil
								},
							),
							skipWhitespaceToken,
						},
						orElse: nopAcceptor,
					},
					&conditionalTokenAcceptor{
						ifAccept: acceptOperator(","),
						andThen: tokenAcceptors{
							skipWhitespaceToken,
							deferAcceptor(func() tokenAcceptor {
								*aggregations = append(*aggregations, &SumAggregation{Alias: alias, Property: prop})
								return acceptAggregations(aggregations)
							}),
						},
						orElse: deferAcceptor(func() tokenAcceptor {
							*aggregations = append(*aggregations, &SumAggregation{Alias: alias, Property: prop})
							return nopAcceptor
						}),
					},
				},
				orElse: &conditionalTokenAcceptor{
					ifAccept: acceptKeyword("AVG"),
					andThen: tokenAcceptors{
						skipWhitespaceToken,
						acceptOperator("("),
						skipWhitespaceToken,
						acceptEitherToken(
							func(token *SymbolToken) error {
								prop = token.Content
								return nil
							},
							func(token *StringToken) error {
								if token.Quote != '`' {
									return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, token.GetContent(), token.GetPosition())
								}
								prop = token.Content
								return nil
							},
						),
						skipWhitespaceToken,
						acceptOperator(")"),
						skipWhitespaceToken,
						&conditionalTokenAcceptor{
							ifAccept: acceptKeyword("AS"),
							andThen: tokenAcceptors{
								acceptWhitespaceToken,
								acceptEitherToken(
									func(token *SymbolToken) error {
										alias = token.Content
										return nil
									},
									func(token *StringToken) error {
										if token.Quote != '`' {
											return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, token.GetContent(), token.GetPosition())
										}
										alias = token.Content
										return nil
									},
								),
								skipWhitespaceToken,
							},
							orElse: nopAcceptor,
						},
						&conditionalTokenAcceptor{
							ifAccept: acceptOperator(","),
							andThen: tokenAcceptors{
								skipWhitespaceToken,
								deferAcceptor(func() tokenAcceptor {
									*aggregations = append(*aggregations, &AvgAggregation{Alias: alias, Property: prop})
									return acceptAggregations(aggregations)
								}),
							},
							orElse: deferAcceptor(func() tokenAcceptor {
								*aggregations = append(*aggregations, &AvgAggregation{Alias: alias, Property: prop})
								return nopAcceptor
							}),
						},
					},
					orElse: tokenAcceptorFn(func(tr tokenReader) error {
						token, err := tr.Read()
						if err != nil {
							return err
						}
						return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, token.GetContent(), token.GetPosition())
					}),
				},
			},
		},
	}
}

func ParseQuery(ts TokenSource) (*Query, error) {
	var query Query
	acceptor := acceptQuery(&query)
	// if err := acceptor.accept(&debugTokenSource{ts}); err != nil {
	if err := acceptor.accept(ts); err != nil {
		return nil, err
	}
	if ts.Next() {
		tok, _ := ts.Read()
		return nil, fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, tok.GetContent(), tok.GetPosition())
	}
	return &query, nil
}

func acceptQuery(query *Query) tokenAcceptor {
	return tokenAcceptors{
		skipWhitespaceToken,
		acceptKeyword("SELECT"),
		acceptWhitespaceToken,
		&conditionalTokenAcceptor{
			ifAccept: acceptKeyword("DISTINCT"),
			andThen:  acceptDistinctBody(query),
			orElse:   nopAcceptor,
		},
		acceptProperties(&query.Properties, true),
		acceptWhitespaceToken,
		acceptKeyword("FROM"),
		acceptWhitespaceToken,
		acceptEitherToken(
			func(tok *SymbolToken) error {
				query.Kind = Kind(tok.Content)
				return nil
			},
			func(tok *StringToken) error {
				if tok.Quote != '`' {
					return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, tok.Content, tok.Position)
				}
				query.Kind = Kind(tok.Content)
				return nil
			},
		),
		&conditionalTokenAcceptor{
			ifAccept: tokenAcceptors{
				acceptWhitespaceToken,
				acceptKeyword("WHERE"),
			},
			andThen: tokenAcceptors{
				acceptWhitespaceToken,
				acceptCondition(&query.Where),
			},
			orElse: nopAcceptor,
		},
		&conditionalTokenAcceptor{
			ifAccept: tokenAcceptors{
				acceptWhitespaceToken,
				acceptKeyword("ORDER"),
				acceptWhitespaceToken,
				acceptKeyword("BY"),
			},
			andThen: tokenAcceptors{
				acceptWhitespaceToken,
				acceptOrderByBody(&query.OrderBy),
			},
			orElse: nopAcceptor,
		},
		&conditionalTokenAcceptor{
			ifAccept: tokenAcceptors{
				acceptWhitespaceToken,
				acceptKeyword("LIMIT"),
			},
			andThen: tokenAcceptors{
				acceptWhitespaceToken,
				deferAcceptor(func() tokenAcceptor {
					query.Limit = new(Limit)
					return acceptLimitBody(query.Limit)
				}),
			},
			orElse: nopAcceptor,
		},
		&conditionalTokenAcceptor{
			ifAccept: tokenAcceptors{
				acceptWhitespaceToken,
				acceptKeyword("OFFSET"),
			},
			andThen: tokenAcceptors{
				acceptWhitespaceToken,
				deferAcceptor(func() tokenAcceptor {
					query.Offset = new(Offset)
					return acceptOffsetBody(query.Offset)
				}),
			},
			orElse: nopAcceptor,
		},
		skipWhitespaceToken,
	}
}

func acceptDistinctBody(query *Query) tokenAcceptor {
	return tokenAcceptors{
		acceptWhitespaceToken,
		&conditionalTokenAcceptor{
			ifAccept: acceptKeyword("ON"),
			andThen: tokenAcceptors{
				acceptWhitespaceToken,
				acceptOperator("("),
				skipWhitespaceToken,
				acceptProperties(&query.DistinctOn, false),
				skipWhitespaceToken,
				acceptOperator(")"),
				skipWhitespaceToken,
			},
			orElse: tokenAcceptors{
				notAcceptor(acceptWildcardToken),
				deferAcceptor(func() tokenAcceptor {
					query.Distinct = true
					return nopAcceptor
				}),
			},
		},
	}
}

func acceptProperties(props *[]Property, wildcard bool) tokenAcceptor {
	if wildcard {
		return tokenAcceptors{
			acceptTokenFromAny3(
				func(*WildcardToken) error {
					*props = nil
					return nil
				},
				func(tok *SymbolToken) error {
					*props = append(*props, Property(tok.Content))
					return nil
				},
				func(tok *StringToken) error {
					if tok.Quote != '`' {
						return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, tok.Content, tok.Position)
					}
					*props = append(*props, Property(tok.Content))
					return nil
				},
			),
			&conditionalTokenAcceptor{
				ifAccept: tokenAcceptors{
					skipWhitespaceToken,
					acceptOperator(","),
					skipWhitespaceToken,
				},
				andThen: deferAcceptor(func() tokenAcceptor {
					return acceptProperties(props, false)
				}),
				orElse: nopAcceptor,
			},
		}
	} else {
		return tokenAcceptors{
			acceptEitherToken(
				func(tok *SymbolToken) error {
					*props = append(*props, Property(tok.Content))
					return nil
				},
				func(tok *StringToken) error {
					if tok.Quote != '`' {
						return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, tok.Content, tok.Position)
					}
					*props = append(*props, Property(tok.Content))
					return nil
				},
			),
			&conditionalTokenAcceptor{
				ifAccept: tokenAcceptors{
					skipWhitespaceToken,
					acceptOperator(","),
					skipWhitespaceToken,
				},
				andThen: deferAcceptor(func() tokenAcceptor {
					return acceptProperties(props, false)
				}),
				orElse: nopAcceptor,
			},
		}
	}
}

func ParseCondition(ts TokenSource) (Condition, error) {
	var condition Condition
	acceptor := acceptCondition(&condition)
	if err := acceptor.accept(ts); err != nil {
		return nil, err
	}
	if ts.Next() {
		tok, _ := ts.Read()
		return nil, fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, tok.GetContent(), tok.GetPosition())
	}
	return condition, nil
}

func acceptCondition(cond *Condition) tokenAcceptor {
	return tokenAcceptorFn(func(tr tokenReader) error {
		ast, err := constructAST(tr, 0)
		if err != nil {
			return err
		}

		if c, err := ast.toCondition(); err != nil {
			return err
		} else {
			*cond = c
			return nil
		}
	})
}

func ParseKey(ts TokenSource) (*Key, error) {
	var key Key
	acceptor := tokenAcceptors{
		acceptKeyword("KEY"),
		acceptKeyBody(&key),
	}
	if err := acceptor.accept(ts); err != nil {
		return nil, err
	}
	if ts.Next() {
		tok, _ := ts.Read()
		return nil, fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, tok.GetContent(), tok.GetPosition())
	}
	return &key, nil
}

func acceptKeyBody(result *Key) tokenAcceptor {
	return tokenAcceptors{
		acceptOperator("("),
		skipWhitespaceToken,
		&conditionalTokenAcceptor{
			ifAccept: acceptKeyword("PROJECT"),
			andThen: tokenAcceptors{
				acceptOperator("("),
				skipWhitespaceToken,
				acceptSingleToken(func(token *StringToken) error {
					if token.Quote == '`' {
						return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, token.GetContent(), token.GetPosition())
					}
					result.ProjectID = ProjectID(token.Content)
					return nil
				}),
				skipWhitespaceToken,
				acceptOperator(")"),
				skipWhitespaceToken,
				acceptOperator(","),
				skipWhitespaceToken,
			},
			orElse: nopAcceptor,
		},
		&conditionalTokenAcceptor{
			ifAccept: acceptKeyword("NAMESPACE"),
			andThen: tokenAcceptors{
				acceptOperator("("),
				skipWhitespaceToken,
				acceptSingleToken(func(token *StringToken) error {
					if token.Quote == '`' {
						return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, token.GetContent(), token.GetPosition())
					}
					result.Namespace = token.Content
					return nil
				}),
				skipWhitespaceToken,
				acceptOperator(")"),
				skipWhitespaceToken,
				acceptOperator(","),
				skipWhitespaceToken,
			},
			orElse: nopAcceptor,
		},
		acceptKeyPath(&result.Path),
		acceptOperator(")"),
	}
}

func acceptKeyPath(keyPaths *[]*KeyPath) tokenAcceptor {
	var keyPath KeyPath
	return tokenAcceptors{
		acceptSingleToken(func(token *SymbolToken) error {
			keyPath.Kind = Kind(token.Content)
			return nil
		}),
		skipWhitespaceToken,
		acceptOperator(","),
		skipWhitespaceToken,
		acceptEitherToken(
			func(token *StringToken) error {
				if token.Quote == '`' {
					return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, token.GetContent(), token.GetPosition())
				}
				keyPath.Name = token.Content
				return nil
			},
			func(token *NumericToken) error {
				if token.Floating {
					return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, token.GetContent(), token.GetPosition())
				}
				keyPath.ID = token.Int64
				return nil
			},
		),
		skipWhitespaceToken,
		deferAcceptor(func() tokenAcceptor {
			*keyPaths = append(*keyPaths, &keyPath)
			return nopAcceptor
		}),
		&conditionalTokenAcceptor{
			ifAccept: tokenAcceptors{
				acceptOperator(","),
				skipWhitespaceToken,
			},
			andThen: deferAcceptor(func() tokenAcceptor {
				return acceptKeyPath(keyPaths)
			}),
			orElse: nopAcceptor,
		},
	}
}

func acceptArrayBody(result *[]conditionValuer) tokenAcceptor {
	var v conditionValuer
	return tokenAcceptors{
		acceptOperator("("),
		skipWhitespaceToken,
		acceptConditionValue(&v),
		skipWhitespaceToken,
		deferAcceptor(func() tokenAcceptor {
			*result = append(*result, v)
			return nopAcceptor
		}),
		&conditionalTokenAcceptor{
			ifAccept: acceptOperator(","),
			andThen:  acceptMoreArrayBody(result),
			orElse:   nopAcceptor,
		},
		acceptOperator(")"),
	}
}

func acceptMoreArrayBody(result *[]conditionValuer) tokenAcceptor {
	var v conditionValuer
	return tokenAcceptors{
		skipWhitespaceToken,
		acceptConditionValue(&v),
		skipWhitespaceToken,
		deferAcceptor(func() tokenAcceptor {
			*result = append(*result, v)
			return nopAcceptor
		}),
		&conditionalTokenAcceptor{
			ifAccept: acceptOperator(","),
			andThen: deferAcceptor(func() tokenAcceptor {
				return acceptMoreArrayBody(result)
			}),
			orElse: nopAcceptor,
		},
	}
}

func acceptBlobBody(result *[]byte) tokenAcceptor {
	return tokenAcceptors{
		acceptOperator("("),
		skipWhitespaceToken,
		acceptSingleToken(func(token *StringToken) error {
			if token.Quote == '`' {
				return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, token.GetContent(), token.GetPosition())
			}

			b, err := base64.RawURLEncoding.DecodeString(token.Content)
			if err != nil {
				return fmt.Errorf("%w: %s at %d (%w)", ErrUnexpectedToken, token.GetContent(), token.GetPosition(), err)
			}

			*result = b
			return nil
		}),
		skipWhitespaceToken,
		acceptOperator(")"),
	}
}

func acceptDateTimeBody(result *time.Time) tokenAcceptor {
	return tokenAcceptors{
		acceptOperator("("),
		skipWhitespaceToken,
		acceptSingleToken(func(token *StringToken) error {
			if token.Quote == '`' {
				return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, token.GetContent(), token.GetPosition())
			}

			t, err := time.Parse(time.RFC3339Nano, token.Content)
			if err != nil {
				return fmt.Errorf("%w: %s at %d (%w)", ErrUnexpectedToken, token.GetContent(), token.GetPosition(), err)
			}

			*result = t
			return nil
		}),
		skipWhitespaceToken,
		acceptOperator(")"),
	}
}

func acceptOrderByBody(orderBy *[]OrderBy) tokenAcceptor {
	var prop Property
	return tokenAcceptors{
		acceptEitherToken(
			func(tok *SymbolToken) error {
				prop = Property(tok.Content)
				return nil
			},
			func(tok *StringToken) error {
				if tok.Quote != '`' {
					return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, tok.Content, tok.Position)
				}
				prop = Property(tok.Content)
				return nil
			},
		),
		&conditionalTokenAcceptor{
			ifAccept: tokenAcceptors{
				acceptWhitespaceToken,
				acceptSingleToken(func(token *OrderToken) error {
					*orderBy = append(*orderBy, OrderBy{Descending: token.Descending, Property: prop})
					return nil
				}),
			},
			andThen: nopAcceptor,
			orElse: deferAcceptor(func() tokenAcceptor {
				*orderBy = append(*orderBy, OrderBy{Descending: false, Property: prop})
				return nopAcceptor
			}),
		},
		&conditionalTokenAcceptor{
			ifAccept: tokenAcceptors{
				skipWhitespaceToken,
				acceptOperator(","),
				skipWhitespaceToken,
			},
			andThen: deferAcceptor(func() tokenAcceptor {
				return acceptOrderByBody(orderBy)
			}),
			orElse: nopAcceptor,
		},
	}
}

func acceptLimitBody(limit *Limit) tokenAcceptor {
	var wantNextCursor bool
	return &conditionalTokenAcceptor{
		ifAccept: acceptKeyword("FIRST"),
		andThen: tokenAcceptors{
			skipWhitespaceToken,
			acceptOperator("("),
			skipWhitespaceToken,
			acceptEitherToken(
				func(token *NumericToken) error {
					if token.Floating {
						return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, token.GetContent(), token.GetPosition())
					}
					limit.Position = token.Int64
					wantNextCursor = true
					return nil
				},
				func(token *BindingToken) error {
					limit.Cursor = parseBindingToken(token)
					return nil
				},
			),
			skipWhitespaceToken,
			acceptOperator(","),
			skipWhitespaceToken,
			acceptEitherToken(
				func(token *NumericToken) error {
					if token.Floating {
						return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, token.GetContent(), token.GetPosition())
					}
					if wantNextCursor {
						return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, token.GetContent(), token.GetPosition())
					}
					limit.Position = token.Int64
					return nil
				},
				func(token *BindingToken) error {
					if !wantNextCursor {
						return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, token.GetContent(), token.GetPosition())
					}
					limit.Cursor = parseBindingToken(token)
					return nil
				},
			),
			skipWhitespaceToken,
			acceptOperator(")"),
		},
		orElse: acceptSingleToken(func(token *NumericToken) error {
			if token.Floating {
				return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, token.GetContent(), token.GetPosition())
			}
			limit.Position = token.Int64
			return nil
		}),
	}
}

func acceptOffsetBody(offset *Offset) tokenAcceptor {
	return tokenAcceptors{
		acceptEitherToken(
			func(token *NumericToken) error {
				if token.Floating {
					return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, token.GetContent(), token.GetPosition())
				}
				offset.Position = token.Int64
				return nil
			},
			func(token *BindingToken) error {
				offset.Cursor = parseBindingToken(token)
				return nil
			},
		),
		&conditionalTokenAcceptor{
			ifAccept: tokenAcceptors{
				skipWhitespaceToken,
				acceptOperator("+"),
				skipWhitespaceToken,
			},
			andThen: acceptSingleToken(func(token *NumericToken) error {
				if token.Floating {
					return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, token.GetContent(), token.GetPosition())
				}
				if offset.Cursor == nil {
					return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, token.GetContent(), token.GetPosition())
				}
				offset.Position = token.Int64
				return nil
			}),
			orElse: nopAcceptor,
		},
	}
}

func parseBindingToken(bind *BindingToken) BindingVariable {
	if bind.Index == 0 {
		return &NamedBinding{Name: bind.Name}
	}
	return &IndexedBinding{Index: bind.Index}
}
