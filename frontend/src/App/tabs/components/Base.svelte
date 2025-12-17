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

  let element;
  const id = Symbol();

  let configAvailable = false;
  let configData = null;

  let tooltipContent = null;
  $: {
    const items = [];
    if (ttKey) items.push({ key: $t(ttKey) });
    if (ttImage) items.push({ image: ttImage });
    tooltipContent = buildTooltipContent(items);
  }

  // Регистрация в поиске
  $: if (element) {
    searchableItems.update(id, {
      label: $t(label.toLowerCase()),
      element,
      tabId
    });
  }

  // Следим за доступностью конфига
  $: {
    const storeValue = $configStore;
    configAvailable = !!storeValue?.path && storeValue.error === null;
    configData = storeValue?.data;
  }

  onMount(() => {
    searchableItems.register({
      id,
      label: $t(label.toLowerCase()),
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
  class={className}
  {style}
  use:tt={{ content: tooltipContent, placement: ttPlace }}
  {...$$restProps}
  on:keydown
  on:click
  on:focus
  on:blur
  on:mouseup
>
  <slot {configAvailable} {configData} />
</div>
