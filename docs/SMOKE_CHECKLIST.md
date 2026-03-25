# Smoke Checklist

This checklist is intended for pre-release validation on a real desktop build.

## Automated Gate

Run one of the following:

```bash
task smoke
```

or on Windows:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\smoke.ps1
```

Expected result: all commands complete without errors.

## Manual Scenarios

1. First Run Path Setup
   - Start app with empty settings or clear `game_path`.
   - Verify settings auto-open on first run.
   - Verify path scanner can find/add at least one path.
   - Expected: selected path is saved and reused after restart.

2. Config Availability Notification
   - Set invalid `game_path`.
   - Expected: "config not found" notification is visible.
   - Click open settings from notification.
   - Set valid path.
   - Expected: notification disappears.

3. Core Config Editing
   - Change a checkbox, slider, and color picker value.
   - Close app and start again.
   - Expected: values persist in UI and in `config.lod.ini`.

4. Watcher External Change Flow
   - Keep app open and modify `config.lod.ini` externally.
   - Expected: watcher shows diff entries.
   - Test both actions:
   - Accept changes.
   - Discard changes.
   - Expected: no crashes and consistent config state.

5. Theme Apply and Editor Save
   - Switch current theme in settings.
   - Open Theme Editor window.
   - Change `color` token, switch value mode (`literal/ref/derived`) and save.
   - Edit `shadow` and `gradient` tokens (`app.panel.shadow`, `app.panel.gradient`) and save.
   - Restart app.
   - Expected: theme changes persist and apply on startup.

6. Language and Translation Tools
   - Switch language in settings.
   - Verify visible strings update in UI.
   - Create/update translation entry via language tools.
   - Expected: saved translation is available after reopen.

7. Map Download Flow
   - Open map download tab and run download.
   - Expected: request starts quickly, progress/status updates are shown.
   - For network failure, expected: user-facing error notification (no silent failure).

8. Updater Flow
   - Trigger check for updates from About tab.
   - If update is available, run update flow.
   - Expected: progress updates, restart button appears after download.
   - If no update is available, expected: no crash and clear status feedback.
