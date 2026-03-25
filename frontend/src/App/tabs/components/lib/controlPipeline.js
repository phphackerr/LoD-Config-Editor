import { saveConfigValue } from '../../../lib/store/config';
import { isInternalChange } from '../../../lib/store/internalChange';
import { tr } from '../../../lib/store/storeUtils';

export async function persistControlValue(section, option, value, fallbackMessage) {
  const sectionName = String(section || '').trim();
  const optionName = String(option || '').trim();

  if (!sectionName || !optionName) {
    return {
      ok: false,
      error: tr('ERRORS.controls.invalid_target')
    };
  }

  isInternalChange.mark();
  return await saveConfigValue(sectionName, optionName, value, fallbackMessage);
}
