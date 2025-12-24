<script module>
  export const tabMetadata = {
    order: 3
  };
</script>

<script>
  import { onMount } from 'svelte';
  import { t } from 'svelte-i18n';
  import { appSettings, updateLanguage } from '../../lib/store/appSettings';
  import { changeLanguage, getAvailableLanguages } from '../../lib/store/i18n';

  let languages = [];
  let currentLang = 'en';

  onMount(async () => {
    languages = await getAvailableLanguages();
    currentLang = $appSettings.language;
  });

  async function handleLanguageChange(event) {
    const newLang = event.target.value;
    currentLang = newLang;
    await updateLanguage(newLang);
    await changeLanguage(newLang);
  }
</script>

<div class="content">
  <div class="setting-row">
    <div class="label">{$t('SETTINGS.LANGUAGE.lang')}</div>
    <div class="dropdown-wrapper">
      <label class="dropdown">
        <select value={currentLang} on:change={handleLanguageChange}>
          {#each languages as lang}
            <option value={lang.code}>{lang.name}</option>
          {/each}
        </select>
      </label>
    </div>
  </div>
</div>

<style>
  .content {
    padding: 20px;
    color: var(--text-color);
  }

  .setting-row {
    display: flex;
    align-items: center;
    gap: 20px;
    margin-bottom: 20px;
  }

  .label {
    font-size: 16px;
    min-width: 100px;
  }

  /* Styles borrowed from Dropdown.svelte */
  .dropdown-wrapper {
    display: inline-flex;
    width: fit-content;
  }

  .dropdown {
    display: flex;
    gap: 10px;
    align-items: center;
    cursor: pointer;
    user-select: none;
    padding: 5px 10px;
    background: var(--element-bg-color);
    border-radius: 4px;
    border: 1px solid transparent;
    transition:
      background-color 0.2s,
      border-color 0.2s;
    width: fit-content;
  }

  .dropdown:hover {
    background: var(--element-bg-hover-color);
  }

  .dropdown select {
    background-color: rgba(255, 255, 255, 0.1);
    border: 1px solid var(--dd-select-border-color);
    border-radius: 4px;
    color: var(--color);
    cursor: pointer;
    transition: all 0.2s ease;
    padding: 5px;
    font-size: 15px;
    min-width: 150px;
  }

  .dropdown select:hover {
    background-color: rgba(255, 255, 255, 0.15);
    border-color: rgba(255, 255, 255, 0.3);
  }

  .dropdown select:focus {
    outline: none;
    border-color: #ffd700;
    box-shadow: 0 0 0 2px rgba(255, 215, 0, 0.2);
  }

  .dropdown select option {
    background-color: #2a2a2a;
    color: #fff;
  }
</style>
