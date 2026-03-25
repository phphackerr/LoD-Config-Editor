package theming

import (
	"strings"
	"testing"
)

func TestResolveTheme_KeepsDefaultValueWhenStatesPresent(t *testing.T) {
	theme := Theme{
		Tokens: []Token{
			{
				ID:   "titlebar.button.bg",
				Type: TokenColor,
				Value: TokenValue{
					Literal: &Literal{
						Color: &ColorLiteral{Space: "srgb", R: 10, G: 20, B: 30, A: 1},
					},
				},
				States: map[string]TokenValue{
					"hover": {
						Literal: &Literal{
							Color: &ColorLiteral{Space: "srgb", R: 40, G: 50, B: 60, A: 1},
						},
					},
				},
			},
		},
	}

	resolved, errs := ResolveTheme(theme)
	if len(errs) > 0 {
		t.Fatalf("unexpected resolver errors: %v", errs)
	}

	token, ok := resolved.Tokens["titlebar.button.bg"]
	if !ok {
		t.Fatalf("resolved token not found")
	}

	defaultColor := token.States["default"].Color
	if defaultColor == nil {
		t.Fatalf("default state is missing")
	}
	if defaultColor.R != 10 || defaultColor.G != 20 || defaultColor.B != 30 {
		t.Fatalf("unexpected default color: %+v", *defaultColor)
	}

	hoverColor := token.States["hover"].Color
	if hoverColor == nil {
		t.Fatalf("hover state is missing")
	}
	if hoverColor.R != 40 || hoverColor.G != 50 || hoverColor.B != 60 {
		t.Fatalf("unexpected hover color: %+v", *hoverColor)
	}
}

func TestResolveTheme_ResolvesRefInDefaultAndState(t *testing.T) {
	theme := Theme{
		Tokens: []Token{
			{
				ID:   "titlebar.text",
				Type: TokenColor,
				Value: TokenValue{
					Literal: &Literal{
						Color: &ColorLiteral{Space: "srgb", R: 33, G: 33, B: 33, A: 1},
					},
				},
			},
			{
				ID:   "titlebar.button.icon.color",
				Type: TokenColor,
				Value: TokenValue{
					Ref: &Ref{ID: "titlebar.text"},
				},
				States: map[string]TokenValue{
					"hover": {
						Ref: &Ref{ID: "titlebar.text"},
					},
				},
			},
		},
	}

	resolved, errs := ResolveTheme(theme)
	if len(errs) > 0 {
		t.Fatalf("unexpected resolver errors: %v", errs)
	}

	token, ok := resolved.Tokens["titlebar.button.icon.color"]
	if !ok {
		t.Fatalf("resolved token not found")
	}

	defaultColor := token.States["default"].Color
	if defaultColor == nil {
		t.Fatalf("default state is missing")
	}
	if defaultColor.R != 33 || defaultColor.G != 33 || defaultColor.B != 33 {
		t.Fatalf("unexpected default color: %+v", *defaultColor)
	}

	hoverColor := token.States["hover"].Color
	if hoverColor == nil {
		t.Fatalf("hover state is missing")
	}
	if hoverColor.R != 33 || hoverColor.G != 33 || hoverColor.B != 33 {
		t.Fatalf("unexpected hover color: %+v", *hoverColor)
	}
}

func TestResolveTheme_ResolvesRefInsideShadowLiteral(t *testing.T) {
	theme := Theme{
		Tokens: []Token{
			{
				ID:   "app.shadow.color",
				Type: TokenColor,
				Value: TokenValue{
					Literal: &Literal{
						Color: &ColorLiteral{Space: "srgb", R: 12, G: 24, B: 36, A: 0.5},
					},
				},
			},
			{
				ID:   "app.panel.shadow",
				Type: TokenShadow,
				Value: TokenValue{
					Literal: &Literal{
						Shadow: []ShadowLayer{
							{
								X:      0,
								Y:      8,
								Blur:   16,
								Spread: 0,
								Color: TokenValue{
									Ref: &Ref{ID: "app.shadow.color"},
								},
							},
						},
					},
				},
			},
		},
	}

	resolved, errs := ResolveTheme(theme)
	if len(errs) > 0 {
		t.Fatalf("unexpected resolver errors: %v", errs)
	}

	token := resolved.Tokens["app.panel.shadow"]
	if len(token.States["default"].Shadow) != 1 {
		t.Fatalf("expected one shadow layer, got: %d", len(token.States["default"].Shadow))
	}

	layerColor := token.States["default"].Shadow[0].Color.Literal
	if layerColor == nil || layerColor.Color == nil {
		t.Fatalf("shadow color was not resolved into literal color")
	}

	got := layerColor.Color
	if got.R != 12 || got.G != 24 || got.B != 36 || got.A != 0.5 {
		t.Fatalf("unexpected resolved shadow color: %+v", *got)
	}

	css, err := CompileCSS(resolved)
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}
	if !strings.Contains(css, "--app-panel-shadow:") {
		t.Fatalf("compiled css does not contain app.panel.shadow variable")
	}
}

func TestResolveTheme_ResolvesDerivedInsideGradientStop(t *testing.T) {
	theme := Theme{
		Tokens: []Token{
			{
				ID:   "app.base",
				Type: TokenColor,
				Value: TokenValue{
					Literal: &Literal{
						Color: &ColorLiteral{Space: "srgb", R: 100, G: 100, B: 100, A: 1},
					},
				},
			},
			{
				ID:   "app.panel.gradient",
				Type: TokenGradient,
				Value: TokenValue{
					Literal: &Literal{
						Gradient: &GradientLiteral{
							Kind:  "linear",
							Angle: 90,
							Stops: []GradientStop{
								{
									Pos: 0,
									Color: TokenValue{
										Ref: &Ref{ID: "app.base"},
									},
								},
								{
									Pos: 1,
									Color: TokenValue{
										Derived: &Derived{
											Op: "lighten",
											From: TokenValue{
												Ref: &Ref{ID: "app.base"},
											},
											Amount: 10,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	resolved, errs := ResolveTheme(theme)
	if len(errs) > 0 {
		t.Fatalf("unexpected resolver errors: %v", errs)
	}

	css, err := CompileCSS(resolved)
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}
	if !strings.Contains(css, "--app-panel-gradient:") {
		t.Fatalf("compiled css does not contain app.panel.gradient variable")
	}
}
