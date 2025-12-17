<script>
  export let show = false;
  export let text = "Загрузка..."; // Текст по умолчанию, можно переопределить
</script>

{#if show}
  <div class="loading-overlay">
    <div class="loading-content">
      <div class="eyes"></div>
      <span class="loading-text">{text}</span>
    </div>
  </div>
{/if}

<style>
  .loading-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.7); /* Затемненный фон */
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1000; /* Поверх всех остальных элементов */
  }

  .loading-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    color: white;
    font-size: 1.2em;
  }

  /* HTML: <div class="loader"></div> */
  .eyes {
    display: inline-flex;
    gap: 10px;
  }
  .eyes:before,
  .eyes:after {
    content: "";
    height: 20px;
    aspect-ratio: 1;
    border-radius: 50%;
    background: radial-gradient(farthest-side, #000 95%, #0000) 35% 35%/6px 6px
      no-repeat #fff;
    animation: l5 3s infinite;
  }
  @keyframes l5 {
    0%,
    11% {
      background-position: 35% 35%;
    }
    14%,
    36% {
      background-position: 65% 35%;
    }
    38%,
    61% {
      background-position: 65% 65%;
    }
    64%,
    86% {
      background-position: 35% 65%;
    }
    88%,
    100% {
      background-position: 35% 35%;
    }
  }

  .loading-text {
    margin-top: 10px;
  }
</style>
