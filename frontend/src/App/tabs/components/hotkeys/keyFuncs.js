// 🔑 Справочник из твоего кода
import { KEY_TO_CODE } from './keyCodes';

// Хелпер для красивого отображения
function formatKeyName(name) {
  if (!name) return null;

  // Всегда с маленькой буквы для проверки
  const val = name.toLowerCase();

  if (val === 'wheelup') return 'wheelUp';
  if (val === 'wheeldn') return 'wheelDn';

  // Однобуквенные и цифры → просто в UpperCase
  if (val.length === 1) {
    return val.toUpperCase();
  }

  // Остальные → первая буква большая, остальное как есть
  return val.charAt(0).toUpperCase() + val.slice(1);
}

// Универсальная нормализация
export function normalizeKey(input) {
  if (!input) return '';

  let str = input.toString().trim();

  // --- 1. Разбиваем на сегменты: либо "0x.." либо текст ---
  // Например "Alt0x51" → ["Alt", "0x51"]
  // "0x210x57" → ["0x21", "0x57"]
  const parts = str.match(/0x[0-9a-f]+|[a-z]+/gi) || [];

  const modifierMap = {
    ctrl: 'Ctrl',
    control: 'Ctrl',
    shift: 'Shift',
    alt: 'Alt'
  };

  const result = [];

  for (const part of parts) {
    if (/^0x[0-9a-f]+$/i.test(part)) {
      // HEX → ищем в словаре
      const found = Object.entries(KEY_TO_CODE).find(
        ([, v]) => v.toLowerCase() === part.toLowerCase()
      );
      if (found) {
        const keyName = found[0].toLowerCase();
        if (keyName === 'space') {
          result.push('Space');
        } else if (keyName === 'wheelup' || keyName === 'wheeldn') {
          result.push(formatKeyName(keyName));
        } else {
          result.push(found[0].toUpperCase());
        }
      } else {
        result.push(part); // fallback если код неизвестен
      }
    } else {
      // текст → проверяем модификаторы
      const lower = part.toLowerCase();
      if (modifierMap[lower]) {
        result.push(modifierMap[lower]);
      } else if (lower === 'wheelup' || lower === 'wheeldn') {
        result.push(formatKeyName(lower));
      } else if (part.length === 1) {
        result.push(part.toUpperCase());
      } else {
        result.push(part.charAt(0).toUpperCase() + part.slice(1));
      }
    }
  }

  return result.join(' + ');
}

// Обратная функция: название → hex
export function encodeKey(display) {
  if (!display) return '';

  // "Ctrl + Q" → ["Ctrl", "Q"]
  const parts = display
    .split('+')
    .map((p) => p.trim())
    .filter(Boolean);

  if (parts.length === 0) return '';

  // === одиночная клавиша или модификатор ===
  if (parts.length === 1) {
    const key = parts[0].toLowerCase();
    const code = KEY_TO_CODE[key];
    if (!code) throw new Error(`Неизвестная клавиша: ${parts[0]}`);
    return code; // всегда hex
  }

  // === модификатор + клавиша ===
  if (parts.length === 2) {
    const modifier = parts[0].toLowerCase();
    const key = parts[1].toLowerCase();

    if (!(modifier === 'ctrl' || modifier === 'alt' || modifier === 'shift')) {
      throw new Error('Допустимы только Ctrl или Alt как модификаторы');
    }

    const code = KEY_TO_CODE[key];
    if (!code) throw new Error(`Неизвестная клавиша: ${parts[1]}`);

    return modifier.charAt(0).toUpperCase() + modifier.slice(1) + code;
  }

  throw new Error('Нельзя закодировать больше двух частей');
}
