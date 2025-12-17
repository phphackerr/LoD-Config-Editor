<script>
  import { t } from 'svelte-i18n';

  export let group; // сюда прилетит bind:group
  export let options = []; // [{ value: "cast", label: "Cast" }, ...]

  let inputElement;

  function handleKeyDown(event, value) {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      group = value; // напрямую обновляем bind:group
    }
  }
</script>

<div class="radio" role="radiogroup">
  {#each options as { value, label }}
    <div
      tabindex="0"
      role="radio"
      on:keydown={(event) => handleKeyDown(event, value)}
      on:click={() => (group = value)}
      aria-checked={group === value}
    >
      <input type="radio" bind:group {value} tabindex="-1" bind:this={inputElement} />
      <span>{$t(label)}</span>
    </div>
  {/each}
</div>

<style>
  .radio {
    display: flex;
    justify-content: center;
    flex-wrap: wrap;
    gap: 10px;
    margin: 10px 0;
    padding: 0 5px;
  }

  .radio div {
    display: flex;
    align-items: center;
    padding: 5px 10px;
    white-space: nowrap;
    cursor: pointer;
    background: var(--element-bg-color);
    border-radius: 4px;
    transition:
      background-color 0.2s,
      border-color 0.2s;
    border: 1px solid transparent;
  }

  .radio input[type='radio']:checked + span {
    color: #ffd700;
  }

  .radio label:has(input[type='radio']:checked) {
    background: rgba(255, 215, 0, 0.15);
    border-color: #ffd700;
  }

  .radio input[type='radio'] {
    appearance: none;
    -webkit-appearance: none;
    width: 16px;
    height: 16px;
    border: 2px solid #fff;
    border-radius: 50%;
    margin-right: 8px;
    position: relative;
    cursor: pointer;
    transition: border-color 0.2s;
  }

  .radio input[type='radio']:checked {
    border-color: #ffd700;
  }

  .radio input[type='radio']:checked::after {
    content: '';
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: 8px;
    height: 8px;
    background: #ffd700;
    border-radius: 50%;
  }
</style>
