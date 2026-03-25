package theming

type ResolvedTheme struct {
	Tokens map[string]ResolvedToken
}

type ResolvedToken struct {
	ID     string
	Type   TokenType
	States map[string]Literal
}

func ResolveTheme(theme Theme) (ResolvedTheme, []ThemeError) {
	index := map[string]Token{}
	for _, t := range theme.Tokens {
		index[t.ID] = t
	}

	resolved := map[string]ResolvedToken{}
	visiting := map[string]bool{}
	var errors []ThemeError

	for id := range index {
		if _, ok := resolved[id]; ok {
			continue
		}
		if err := resolveToken(id, index, resolved, visiting); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return ResolvedTheme{}, errors
	}

	return ResolvedTheme{Tokens: resolved}, nil
}

func resolveToken(
	id string,
	index map[string]Token,
	resolved map[string]ResolvedToken,
	visiting map[string]bool,
) ThemeError {

	if visiting[id] {
		return ErrCycle("tokens." + id)
	}

	if _, ok := resolved[id]; ok {
		return nil
	}

	token, ok := index[id]
	if !ok {
		return ErrUnknownToken(id)
	}

	visiting[id] = true
	states := map[string]Literal{}

	if token.States != nil {
		if hasTokenValue(token.Value) {
			lit, err := resolveValue(
				token.Value,
				"tokens."+id+".value",
				index,
				resolved,
				visiting,
			)
			if err != nil {
				return err
			}
			if err := validateType(token, lit); err != nil {
				return err
			}
			states["default"] = lit
		}

		for state, value := range token.States {
			lit, err := resolveValue(
				value,
				"tokens."+id+".states."+state,
				index,
				resolved,
				visiting,
			)
			if err != nil {
				return err
			}
			if err := validateType(token, lit); err != nil {
				return err
			}
			states[state] = lit
		}
	} else {
		lit, err := resolveValue(
			token.Value,
			"tokens."+id+".value",
			index,
			resolved,
			visiting,
		)
		if err != nil {
			return err
		}
		if err := validateType(token, lit); err != nil {
			return err
		}
		states["default"] = lit
	}

	resolved[id] = ResolvedToken{
		ID:     token.ID,
		Type:   token.Type,
		States: states,
	}

	delete(visiting, id)
	return nil
}

func hasTokenValue(v TokenValue) bool {
	return v.Literal != nil || v.Ref != nil || v.Derived != nil
}

func resolveValue(
	val TokenValue,
	path string,
	index map[string]Token,
	resolved map[string]ResolvedToken,
	visiting map[string]bool,
) (Literal, ThemeError) {

	if err := val.Validate(); err != nil {
		return Literal{}, ErrInvalidValue(path, err.Error())
	}

	switch {
	case val.Literal != nil:
		if err := val.Literal.Validate(); err != nil {
			return Literal{}, ErrInvalidValue(path, err.Error())
		}
		return resolveLiteral(*val.Literal, path+".literal", index, resolved, visiting)

	case val.Ref != nil:
		refID := val.Ref.ID
		if err := resolveToken(refID, index, resolved, visiting); err != nil {
			return Literal{}, err
		}
		rt := resolved[refID]
		return rt.States["default"], nil

	case val.Derived != nil:
		base, err := resolveValue(
			val.Derived.From,
			path+".derived.from",
			index,
			resolved,
			visiting,
		)
		if err != nil {
			return Literal{}, err
		}
		return applyDerived(base, val.Derived, path)

	default:
		return Literal{}, ErrInvalidValue(path, "invalid token value")
	}
}

func resolveLiteral(
	lit Literal,
	path string,
	index map[string]Token,
	resolved map[string]ResolvedToken,
	visiting map[string]bool,
) (Literal, ThemeError) {
	switch {
	case lit.Color != nil:
		return lit, nil

	case lit.Number != nil:
		return lit, nil

	case lit.Gradient != nil:
		next := *lit.Gradient
		if len(next.Stops) > 0 {
			stops := make([]GradientStop, len(next.Stops))
			for i, stop := range next.Stops {
				color, err := resolveColorValue(
					stop.Color,
					path+".gradient.stops",
					index,
					resolved,
					visiting,
				)
				if err != nil {
					return Literal{}, err
				}

				stop.Color = TokenValue{
					Literal: &Literal{
						Color: &color,
					},
				}
				stops[i] = stop
			}
			next.Stops = stops
		}

		return Literal{Gradient: &next}, nil

	case len(lit.Shadow) > 0:
		layers := make([]ShadowLayer, len(lit.Shadow))
		for i, layer := range lit.Shadow {
			color, err := resolveColorValue(
				layer.Color,
				path+".shadow.color",
				index,
				resolved,
				visiting,
			)
			if err != nil {
				return Literal{}, err
			}

			layer.Color = TokenValue{
				Literal: &Literal{
					Color: &color,
				},
			}
			layers[i] = layer
		}
		return Literal{Shadow: layers}, nil
	}

	return Literal{}, ErrInvalidValue(path, "invalid literal")
}

func resolveColorValue(
	value TokenValue,
	path string,
	index map[string]Token,
	resolved map[string]ResolvedToken,
	visiting map[string]bool,
) (ColorLiteral, ThemeError) {
	lit, err := resolveValue(value, path, index, resolved, visiting)
	if err != nil {
		return ColorLiteral{}, err
	}
	if lit.Color == nil {
		return ColorLiteral{}, ErrInvalidValue(path, "expected color literal")
	}
	return *lit.Color, nil
}

func validateType(token Token, lit Literal) ThemeError {
	switch token.Type {
	case TokenColor:
		if lit.Color == nil {
			return ErrInvalidValue(
				"tokens."+token.ID,
				"expected color literal",
			)
		}
	case TokenNumber, TokenSpacing, TokenRadius:
		if lit.Number == nil {
			return ErrInvalidValue(
				"tokens."+token.ID,
				"expected number literal",
			)
		}
	case TokenShadow:
		if len(lit.Shadow) == 0 {
			return ErrInvalidValue(
				"tokens."+token.ID,
				"expected shadow literal",
			)
		}
	case TokenGradient:
		if lit.Gradient == nil {
			return ErrInvalidValue(
				"tokens."+token.ID,
				"expected gradient literal",
			)
		}
	}
	return nil
}

func applyDerived(base Literal, d *Derived, path string) (Literal, ThemeError) {
	if base.Color == nil {
		return Literal{}, ErrInvalidValue(
			path,
			"derived operations supported only for colors",
		)
	}

	c := *base.Color

	switch d.Op {
	case "alpha":
		c.A = clamp(d.Amount, 0, 1)

	case "lighten":
		c.R = clamp(c.R+d.Amount, 0, 255)
		c.G = clamp(c.G+d.Amount, 0, 255)
		c.B = clamp(c.B+d.Amount, 0, 255)

	case "darken":
		c.R = clamp(c.R-d.Amount, 0, 255)
		c.G = clamp(c.G-d.Amount, 0, 255)
		c.B = clamp(c.B-d.Amount, 0, 255)

	default:
		return Literal{}, ErrInvalidValue(
			path,
			"unsupported derive op: "+d.Op,
		)
	}

	return Literal{Color: &c}, nil
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
