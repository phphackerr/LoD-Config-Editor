package theming

import "fmt"

type Theme struct {
	Meta       ThemeMeta        `json:"meta"`
	Tokens     []Token          `json:"tokens"`
	Components []ComponentStyle `json:"components"`
}

type ThemeMeta struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Author  string `json:"author"`
	Version int    `json:"version"`
}

type Token struct {
	ID     string                `json:"id"`
	Type   TokenType             `json:"type"`
	Value  TokenValue            `json:"value,omitempty"`
	States map[string]TokenValue `json:"states,omitempty"`
	Meta   *TokenMeta            `json:"meta,omitempty"`
}

type ComponentStyle struct {
	Component string                           `json:"component"`
	Slots     map[string]map[string]TokenValue `json:"slots"`
}

type TokenMeta struct {
	Label       string `json:"label,omitempty"`
	Group       string `json:"group,omitempty"`
	Order       int    `json:"order,omitempty"`
	Description string `json:"description,omitempty"`
}

type TokenType string

const (
	TokenColor    TokenType = "color"
	TokenNumber   TokenType = "number"
	TokenSpacing  TokenType = "spacing"
	TokenRadius   TokenType = "radius"
	TokenShadow   TokenType = "shadow"
	TokenGradient TokenType = "gradient"
)

type TokenValue struct {
	Literal *Literal `json:"literal,omitempty"`
	Ref     *Ref     `json:"ref,omitempty"`
	Derived *Derived `json:"derived,omitempty"`
}

func (v TokenValue) Kind() string {
	switch {
	case v.Literal != nil:
		return "literal"
	case v.Ref != nil:
		return "ref"
	case v.Derived != nil:
		return "derived"
	default:
		return "invalid"
	}
}

func (v TokenValue) Validate() error {
	count := 0

	if v.Literal != nil {
		count++
	}
	if v.Ref != nil {
		count++
	}
	if v.Derived != nil {
		count++
	}

	if count != 1 {
		return fmt.Errorf("token value must contain exactly one of: literal, ref, derived")
	}

	return nil
}

type Literal struct {
	Color    *ColorLiteral    `json:"color,omitempty"`
	Number   *NumberLiteral   `json:"number,omitempty"`
	Shadow   []ShadowLayer    `json:"shadow,omitempty"`
	Gradient *GradientLiteral `json:"gradient,omitempty"`
}

func (l Literal) Validate() error {
	count := 0

	if l.Color != nil {
		count++
	}
	if l.Number != nil {
		count++
	}
	if l.Gradient != nil {
		count++
	}
	if len(l.Shadow) > 0 {
		count++
	}

	if count != 1 {
		return fmt.Errorf("literal must contain exactly one concrete value")
	}

	return nil
}

type ColorLiteral struct {
	Space string  `json:"space"` // "srgb"
	R     float64 `json:"r"`
	G     float64 `json:"g"`
	B     float64 `json:"b"`
	A     float64 `json:"a"`
}

type NumberLiteral struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"` // px, rem, em, %
}

type ShadowLayer struct {
	X      float64    `json:"x"`
	Y      float64    `json:"y"`
	Blur   float64    `json:"blur"`
	Spread float64    `json:"spread"`
	Inset  bool       `json:"inset"`
	Color  TokenValue `json:"color"`
}

type GradientLiteral struct {
	Kind  string         `json:"kind"` // linear, radial
	Angle float64        `json:"angle,omitempty"`
	Stops []GradientStop `json:"stops"`
}

type GradientStop struct {
	Pos   float64    `json:"pos"` // 0..1
	Color TokenValue `json:"color"`
}

type Ref struct {
	ID string `json:"id"`
}

type Derived struct {
	Op     string     `json:"op"` // lighten, darken, alpha
	From   TokenValue `json:"from"`
	Amount float64    `json:"amount"`
}
