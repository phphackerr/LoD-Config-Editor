<script>
  import { onMount, onDestroy, getContext, createEventDispatcher } from 'svelte';
  import { t } from 'svelte-i18n';
  import { configStore } from '../../lib/store/config';
  import { tt, buildTooltipContent } from '../../lib/tooltip';
  import { searchableItems } from '../../lib/store/search';

  const tabId = getContext('tabId');
  const dispatch = createEventDispatcher();

  export let label = '';
  export let ttKey = '';
  export let ttImage = '';
  export let ttPlace = 'auto';
  export let className = '';
  export let style = '';

  export let section = '';
  export let option = '';

  let element;
  const id = Symbol();

  let configAvailable = false;
  let configData = null;
  let keyMissing = false;

  let tooltipContent = null;
  $: {
    const items = [];
    if (keyMissing) {
      items.push({ key: `Key not found: [${section}] ${option}`, color: '#ff4444' });
    } else {
      if (ttKey) items.push({ key: $t(ttKey) });
      if (ttImage) items.push({ image: ttImage });
    }
    tooltipContent = buildTooltipContent(items);
  }

  // Регистрация в поиске
  $: if (element) {
    searchableItems.update(id, {
      label: $t(label),
      element,
      tabId
    });
  }

  // Следим за доступностью конфига
  $: {
    const storeValue = $configStore;
    configAvailable = !!storeValue?.path && storeValue.error === null;
    configData = storeValue?.data;

    // Check if key exists (case-insensitive)
    if (configAvailable && section && option && configData) {
      if (option === 'ExtraSlot1') {
        keyMissing = false;
      } else {
        const sectionLower = section.toLowerCase();
        const optionLower = option.toLowerCase();

        let foundSection = null;
        // Find section case-insensitively
        for (const secKey of Object.keys(configData)) {
          if (secKey.toLowerCase() === sectionLower) {
            foundSection = configData[secKey];
            break;
          }
        }

        if (!foundSection) {
          keyMissing = true;
        } else {
          // Find option case-insensitively
          let foundOption = false;
          for (const optKey of Object.keys(foundSection)) {
            if (optKey.toLowerCase() === optionLower) {
              foundOption = true;
              break;
            }
          }
          keyMissing = !foundOption;
        }
      }
    } else {
      keyMissing = false;
    }
  }

  onMount(() => {
    searchableItems.register({
      id,
      label: $t(label),
      element,
      tabId
    });
  });

  onDestroy(() => {
    searchableItems.unregister(id);
  });
</script>

<!-- 
  $$restProps позволяет передавать любые атрибуты (role, tabindex, etc.) 
  события (on:keydown) пробрасываются автоматически
-->
<div
  bind:this={element}
  class="{className} {keyMissing ? 'disabled' : ''}"
  {style}
  use:tt={{ content: tooltipContent, placement: ttPlace }}
  {...$$restProps}
  on:keydown
  on:click
  on:focus
  on:blur
  on:mouseup
>
  <slot {configAvailable} {configData} {keyMissing} />
</div>

<style>
  div.disabled > :global(*) {
    pointer-events: none;
    opacity: 0.5;
    filter: grayscale(100%);
  }
</style>
