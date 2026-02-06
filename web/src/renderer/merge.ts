import type { UITree, UIElement } from "./defaultTree";

export type Shortcut = {
  id: string;
  label: string;
  action: string;
  position?: "left" | "center" | "right";
  type?: "button" | "text";
  icon?: string;
  params?: Record<string, unknown>;
  disabled?: boolean;
};

const cloneTree = (tree: UITree): UITree => {
  const elements: Record<string, UIElement> = {};
  for (const [key, element] of Object.entries(tree.elements)) {
    elements[key] = {
      ...element,
      props: element.props ? { ...element.props } : undefined,
      children: element.children ? [...element.children] : undefined,
    };
  }
  return { root: tree.root, elements };
};

const rekeyElements = (tree: UITree, prefix: string): UITree => {
  const elements: Record<string, UIElement> = {};
  for (const [key, element] of Object.entries(tree.elements)) {
    const newKey = `${prefix}${key}`;
    elements[newKey] = {
      ...element,
      key: newKey,
      children: element.children?.map((child) => `${prefix}${child}`),
      parentKey: element.parentKey ? `${prefix}${element.parentKey}` : element.parentKey,
    } as UIElement;
  }
  return { root: `${prefix}${tree.root}`, elements };
};

// Extract shortcuts from view data
export function extractShortcuts(view: Record<string, unknown> | null): Shortcut[] {
  if (!view) return [];

  const shortcuts = view.shortcuts;
  if (!Array.isArray(shortcuts)) return [];

  return shortcuts.map((s: unknown) => {
    const shortcut = s as Record<string, unknown>;
    return {
      id: String(shortcut.id || ""),
      label: String(shortcut.label || ""),
      action: String(shortcut.action || ""),
      position: (shortcut.position as Shortcut["position"]) || "center",
      type: (shortcut.type as Shortcut["type"]) || "button",
      icon: shortcut.icon ? String(shortcut.icon) : undefined,
      params: shortcut.params as Record<string, unknown> | undefined,
      disabled: Boolean(shortcut.disabled),
    };
  }).filter(s => s.id && s.label && s.action);
}

export function mergeViewIntoShell(shell: UITree, view: UITree | null): UITree {
  if (!view) return shell;
  const shellCopy = cloneTree(shell);
  const viewRekeyed = rekeyElements(view, "view:");
  const main = shellCopy.elements.main;
  if (!main) return shellCopy;
  main.children = [viewRekeyed.root];
  shellCopy.elements = { ...shellCopy.elements, ...viewRekeyed.elements };
  return shellCopy;
}

// Merge view with shortcuts support
export function mergeViewWithShortcuts(
  shell: UITree,
  view: UITree | null,
  shortcuts: Shortcut[]
): { tree: UITree; shortcuts: Shortcut[] } {
  const tree = mergeViewIntoShell(shell, view);
  return { tree, shortcuts };
}
