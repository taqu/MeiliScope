import { useEffect, useRef } from "react";

export interface ShortcutDef {
  key: string;
  ctrl?: boolean;
  shift?: boolean;
  alt?: boolean;
  description: string;
  skipInEditor?: boolean;
  handler: () => void;
}

export function useKeyboardShortcuts(shortcuts: ShortcutDef[]) {
  const shortcutsRef = useRef(shortcuts);
  shortcutsRef.current = shortcuts;

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement;
      const isEditable =
        target.tagName === "INPUT" ||
        target.tagName === "TEXTAREA" ||
        target.isContentEditable ||
        target.closest(".cm-editor") !== null;

      for (const sc of shortcutsRef.current) {
        if (
          e.key === sc.key &&
          !!e.ctrlKey === !!sc.ctrl &&
          !!e.shiftKey === !!sc.shift &&
          !!e.altKey === !!sc.alt
        ) {
          if (isEditable && sc.skipInEditor !== false) continue;
          e.preventDefault();
          sc.handler();
          return;
        }
      }
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, []);
}
