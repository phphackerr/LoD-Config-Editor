<script module>
    export const tabMetadata = {
        order: 1,
    };
</script>

<script>
    import CommandPanel from "./components/hotkeys/command-panel.svelte";
    import InventoryPanel from "./components/hotkeys/inventory-panel.svelte";
    import OthersPanel from "./components/hotkeys/others-panel.svelte";
    import { setContext } from "svelte";

    const tabId = "Hotkeys"; // <-- Убедитесь, что здесь правильное имя для каждого файла!
    setContext("tabId", tabId);

    // @ts-nocheck
</script>

<div class="tab">
    <CommandPanel />
    <InventoryPanel />
    <OthersPanel />
</div>

<style>
    .tab {
        height: 100%;
        display: grid;
        grid-template-columns: 1.5fr 1fr 1.5fr;
        gap: 20px;
        align-items: center;
        justify-content: space-between;
        box-sizing: border-box;
        overflow-y: auto;
    }

    @media (max-width: 1100px) {
        .tab {
            grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
        }
    }
</style>
