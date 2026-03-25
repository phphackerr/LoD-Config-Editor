package theming

import (
	"fmt"
	"sort"
	"strings"
)

func CompileCSS(theme ResolvedTheme) (string, error) {
	var b strings.Builder

	b.WriteString(":root {\n")

	tokenIDs := make([]string, 0, len(theme.Tokens))
	for id := range theme.Tokens {
		tokenIDs = append(tokenIDs, id)
	}
	sort.Strings(tokenIDs)

	for _, id := range tokenIDs {
		token := theme.Tokens[id]
		states := make([]string, 0, len(token.States))
		for state := range token.States {
			states = append(states, state)
		}
		sort.Strings(states)

		for _, state := range states {
			lit := token.States[state]
			name := cssVarName(token.ID, state)
			value, err := literalToCSS(lit)
			if err != nil {
				return "", fmt.Errorf("failed to compile token %s[%s]: %w", token.ID, state, err)
			}
			b.WriteString(fmt.Sprintf("  %s: %s;\n", name, value))
		}
	}

	b.WriteString("}\n")
	return b.String(), nil
}

func cssVarName(tokenID, state string) string {
	base := strings.ReplaceAll(tokenID, ".", "-")

	if state == "default" {
		return "--" + base
	}

	return "--" + base + "--" + state
}

func literalToCSS(l Literal) (string, error) {
	switch {
	case l.Color != nil:
		return colorToCSS(*l.Color), nil

	case l.Number != nil:
		return fmt.Sprintf("%g%s", l.Number.Value, l.Number.Unit), nil

	case l.Gradient != nil:
		return gradientToCSS(*l.Gradient)

	case len(l.Shadow) > 0:
		return shadowToCSS(l.Shadow)
	}

	return "", fmt.Errorf("invalid literal")
}

func colorToCSS(c ColorLiteral) string {
	return fmt.Sprintf(
		"rgba(%d, %d, %d, %g)",
		int(c.R),
		int(c.G),
		int(c.B),
		c.A,
	)
}

func shadowToCSS(layers []ShadowLayer) (string, error) {
	var parts []string

	for _, s := range layers {
		color, err := resolveColorLiteral(s.Color, "shadow color")
		if err != nil {
			return "", err
		}

		colorValue, err := literalToCSS(Literal{Color: &color})
		if err != nil {
			return "", err
		}

		part := fmt.Sprintf(
			"%s %gpx %gpx %gpx %gpx %s",
			boolInset(s.Inset),
			s.X,
			s.Y,
			s.Blur,
			s.Spread,
			colorValue,
		)
		parts = append(parts, strings.TrimSpace(part))
	}

	return strings.Join(parts, ", "), nil
}

func boolInset(inset bool) string {
	if inset {
		return "inset"
	}
	return ""
}

func resolveColorLiteral(v TokenValue, context string) (ColorLiteral, error) {
	if v.Literal == nil || v.Literal.Color == nil {
		return ColorLiteral{}, fmt.Errorf("%s must be resolved color literal", context)
	}
	return *v.Literal.Color, nil
}

func gradientToCSS(g GradientLiteral) (string, error) {
	switch g.Kind {
	case "linear":
		var stops []string
		for _, s := range g.Stops {
			color, err := resolveColorLiteral(s.Color, "gradient stop color")
			if err != nil {
				return "", err
			}

			colorValue, err := literalToCSS(Literal{Color: &color})
			if err != nil {
				return "", err
			}

			stops = append(stops,
				fmt.Sprintf(
					"%s %d%%",
					colorValue,
					int(s.Pos*100),
				),
			)
		}
		return fmt.Sprintf(
			"linear-gradient(%gdeg, %s)",
			g.Angle,
			strings.Join(stops, ", "),
		), nil
	}

	return "", fmt.Errorf("unsupported gradient kind: %s", g.Kind)
}
