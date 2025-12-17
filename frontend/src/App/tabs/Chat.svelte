<script module>
    export const tabMetadata = {
        order: 5,
    };
</script>

<script>
    // @ts-nocheck
    import Chat from "./components/Chat.svelte";
    import { setContext } from "svelte";

    const tabId = "Chat"; // <-- ВАЖНО: Замените это на ID для каждого файла!
    setContext("tabId", tabId);

    const SECTIONS = ["QUICKCHAT", "ONSTART"];

    // Генерация ключей для QuickChat
    const quickchatKeys = Array.from(
        { length: 10 },
        (_, i) => `QuickChatText${i + 1}`,
    );
    const quickchatHotkeys = Array.from(
        { length: 10 },
        (_, i) => `QuickChatHotkey${i + 1}`,
    );

    // Генерация ключей для OnStart
    const onstartKeys = Array.from(
        { length: 21 },
        (_, i) => `StartChatString${i + 1}`,
    );

    let quickchatControls = $state(
        quickchatKeys.map((key, i) => ({
            key,
            hotkeyKey: quickchatHotkeys[i],
        })),
    );

    let onstartControls = $state(onstartKeys.map((key) => ({ key })));
</script>

<div class="tab-page">
    <div class="chat-container">
        <!-- Левая колонка - QuickChat -->
        <div class="chat-column">
            <div class="controls-container">
                {#each quickchatControls as { key, hotkeyKey }}
                    <Chat
                        label="QuickChatText"
                        section={SECTIONS[0]}
                        option={key}
                        hotkeyOption={hotkeyKey}
                        width={480}
                        isOnStart={false}
                    />
                {/each}
            </div>
        </div>

        <!-- Правая колонка - OnStart -->
        <div class="chat-column">
            <div class="controls-container">
                {#each onstartControls as { key }}
                    <Chat
                        label="StartChatString"
                        section={SECTIONS[1]}
                        option={key}
                        width={480}
                        isOnStart={true}
                    />
                {/each}
            </div>
        </div>
    </div>
</div>

<style>
    .tab-page {
        padding: 20px;
        height: 100%;
        box-sizing: border-box;
    }

    .chat-container {
        display: flex;
        gap: 20px;
        height: 100%;
    }

    .chat-column {
        flex: 1;
        display: flex;
        flex-direction: column;
        min-width: 0; /* Для корректной работы overflow */
    }

    .controls-container {
        flex: 1;
        overflow-y: auto;
        display: flex;
        flex-direction: column;
        gap: 12px;
        padding-right: 8px; /* Для скроллбара */
    }
</style>
