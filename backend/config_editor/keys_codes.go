package config_editor

import (
	"fmt"
	"strconv"
)

// KeyCodes — словарь для горячих клавиш (код -> название).
var KeyCodes = initKeyCodes()

func initKeyCodes() map[string]string {
	m := make(map[string]string)

	// Мышь
	m["0x04"] = "midlemouse"
	m["0x05"] = "x1"
	m["0x06"] = "x2"

	// Специальные клавиши
	m["0x08"] = "backspace"
	m["0x09"] = "tab"
	m["0x0D"] = "enter"
	m["0x14"] = "caps lock"
	m["0x1B"] = "esc"
	m["0x20"] = "space"
	m["0x21"] = "page up"
	m["0x22"] = "page down"
	m["0x23"] = "end"
	m["0x24"] = "home"
	m["0x25"] = "left"
	m["0x26"] = "up"
	m["0x27"] = "right"
	m["0x28"] = "down"
	m["0x2C"] = "print screen"
	m["0x2D"] = "insert"
	m["0x2E"] = "delete"

	// Цифры 0–9
	for i := 0; i <= 9; i++ {
		key := fmt.Sprintf("0x3%X", i)
		m[key] = strconv.Itoa(i)
	}

	// Буквы A–Z
	for i, c := range "abcdefghijklmnopqrstuvwxyz" {
		key := fmt.Sprintf("0x%02X", 0x41+i) // 0x41 = A
		m[key] = string(c)
	}

	// F-клавиши (F1–F24)
	for i := 1; i <= 24; i++ {
		key := fmt.Sprintf("0x%02X", 0x6F+i) // 0x70 = F1
		m[key] = "f" + strconv.Itoa(i)
	}

	// Numpad
	m["0x60"] = "numpad 0"
	m["0x61"] = "numpad 1"
	m["0x62"] = "numpad 2"
	m["0x63"] = "numpad 3"
	m["0x64"] = "numpad 4"
	m["0x65"] = "numpad 5"
	m["0x66"] = "numpad 6"
	m["0x67"] = "numpad 7"
	m["0x68"] = "numpad 8"
	m["0x69"] = "numpad 9"
	m["0x6A"] = "numpad *"
	m["0x6B"] = "numpad +"
	m["0x6D"] = "numpad -"
	m["0x6E"] = "numpad ."
	m["0x6F"] = "numpad /"

	// Дополнительные клавиши
	m["0x90"] = "num lock"
	m["0x91"] = "scroll lock"

	// Специальные символы
	m["0xBA"] = ";"
	m["0xBB"] = "+"
	m["0xBC"] = ","
	m["0xBD"] = "-"
	m["0xBE"] = "."
	m["0xBF"] = "/"
	m["0xC0"] = "`"
	m["0xDB"] = "["
	m["0xDC"] = "\\"
	m["0xDD"] = "]"

	// Модификаторы
	m["0x10"] = "shift"
	m["0x11"] = "ctrl"
	m["0x12"] = "alt"

	return m
}

// Lookup возвращает название клавиши по hex-коду (например, "0x41" -> "a").
func Lookup(code string) string {
	if name, ok := KeyCodes[code]; ok {
		return name
	}
	return ""
}

// ReverseLookup возвращает hex-код по названию клавиши (например, "a" -> "0x41").
// Полезно, если нужно сохранять обратно в INI.
func ReverseLookup(name string) string {
	for code, keyName := range KeyCodes {
		if keyName == name {
			return code
		}
	}
	return ""
}
