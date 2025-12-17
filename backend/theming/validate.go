package theming

import (
	"regexp"
	"strings"
)

// Регулярки для разных форматов
var (
	hexColor  = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`)
	rgbColor  = regexp.MustCompile(`^rgb\(\s*\d{1,3}\s*,\s*\d{1,3}\s*,\s*\d{1,3}\s*\)$`)
	rgbaColor = regexp.MustCompile(`^rgba\(\s*\d{1,3}\s*,\s*\d{1,3}\s*,\s*\d{1,3}\s*,\s*(0|1|0?\.\d+)\s*\)$`)
	hslColor  = regexp.MustCompile(`^hsl\(\s*\d{1,3}\s*,\s*\d{1,3}%\s*,\s*\d{1,3}%\s*\)$`)
	hslaColor = regexp.MustCompile(`^hsla\(\s*\d{1,3}\s*,\s*\d{1,3}%\s*,\s*\d{1,3}%\s*,\s*(0|1|0?\.\d+)\s*\)$`)
)

// Список всех стандартных CSS named colors (140+)
var namedColors = map[string]bool{
	"aliceblue": true, "antiquewhite": true, "aqua": true, "aquamarine": true,
	"azure": true, "beige": true, "bisque": true, "black": true, "blanchedalmond": true,
	"blue": true, "blueviolet": true, "brown": true, "burlywood": true, "cadetblue": true,
	"chartreuse": true, "chocolate": true, "coral": true, "cornflowerblue": true, "cornsilk": true,
	"crimson": true, "cyan": true, "darkblue": true, "darkcyan": true, "darkgoldenrod": true,
	"darkgray": true, "darkgreen": true, "darkgrey": true, "darkkhaki": true, "darkmagenta": true,
	"darkolivegreen": true, "darkorange": true, "darkorchid": true, "darkred": true, "darksalmon": true,
	"darkseagreen": true, "darkslateblue": true, "darkslategray": true, "darkslategrey": true,
	"darkturquoise": true, "darkviolet": true, "deeppink": true, "deepskyblue": true, "dimgray": true,
	"dimgrey": true, "dodgerblue": true, "firebrick": true, "floralwhite": true, "forestgreen": true,
	"fuchsia": true, "gainsboro": true, "ghostwhite": true, "gold": true, "goldenrod": true,
	"gray": true, "green": true, "greenyellow": true, "grey": true, "honeydew": true,
	"hotpink": true, "indianred": true, "indigo": true, "ivory": true, "khaki": true,
	"lavender": true, "lavenderblush": true, "lawngreen": true, "lemonchiffon": true, "lightblue": true,
	"lightcoral": true, "lightcyan": true, "lightgoldenrodyellow": true, "lightgray": true, "lightgreen": true,
	"lightgrey": true, "lightpink": true, "lightsalmon": true, "lightseagreen": true, "lightskyblue": true,
	"lightslategray": true, "lightslategrey": true, "lightsteelblue": true, "lightyellow": true, "lime": true,
	"limegreen": true, "linen": true, "magenta": true, "maroon": true, "mediumaquamarine": true,
	"mediumblue": true, "mediumorchid": true, "mediumpurple": true, "mediumseagreen": true, "mediumslateblue": true,
	"mediumspringgreen": true, "mediumturquoise": true, "mediumvioletred": true, "midnightblue": true,
	"mintcream": true, "mistyrose": true, "moccasin": true, "navajowhite": true, "navy": true,
	"oldlace": true, "olive": true, "olivedrab": true, "orange": true, "orangered": true,
	"orchid": true, "palegoldenrod": true, "palegreen": true, "paleturquoise": true, "palevioletred": true,
	"papayawhip": true, "peachpuff": true, "peru": true, "pink": true, "plum": true,
	"powderblue": true, "purple": true, "rebeccapurple": true, "red": true, "rosybrown": true,
	"royalblue": true, "saddlebrown": true, "salmon": true, "sandybrown": true, "seagreen": true,
	"seashell": true, "sienna": true, "silver": true, "skyblue": true, "slateblue": true,
	"slategray": true, "slategrey": true, "snow": true, "springgreen": true, "steelblue": true,
	"tan": true, "teal": true, "thistle": true, "tomato": true, "turquoise": true,
	"violet": true, "wheat": true, "white": true, "whitesmoke": true, "yellow": true,
	"yellowgreen": true,
	// спец. значения
	"transparent": true, "currentcolor": true, "inherit": true,
	"initial": true, "unset": true, "none": true,
}

// ValidateColor проверяет, является ли строка валидным CSS-цветом
func ValidateColor(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))

	// Проверка на простые цвета
	if hexColor.MatchString(value) ||
		rgbColor.MatchString(value) ||
		rgbaColor.MatchString(value) ||
		hslColor.MatchString(value) ||
		hslaColor.MatchString(value) ||
		namedColors[value] {
		return true
	}

	// Разрешаем сложные CSS-значения
	if strings.HasPrefix(value, "linear-gradient(") ||
		strings.HasPrefix(value, "radial-gradient(") ||
		strings.HasPrefix(value, "conic-gradient(") ||
		strings.Contains(value, "inset") ||
		strings.Contains(value, "background") ||
		strings.Contains(value, "color") ||
		strings.Contains(value, "px") ||
		strings.Contains(value, ",") {
		return true
	}

	return false
}
