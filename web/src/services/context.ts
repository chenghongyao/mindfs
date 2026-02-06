export type ClientContext = {
  current_root: string;
  current_path?: string;
  selection?: {
    file_path: string;
    start: number;
    end: number;
    text: string;
  };
  current_view?: {
    rule_id: string;
    version: string;
  };
};

type ContextInput = {
  currentRoot: string;
  currentPath?: string | null;
  selection?: {
    filePath: string;
    start: number;
    end: number;
    text: string;
  } | null;
  currentView?: {
    ruleId: string;
    version: string;
  } | null;
};

export function buildClientContext(input: ContextInput): ClientContext {
  const ctx: ClientContext = {
    current_root: input.currentRoot,
  };
  if (input.currentPath) {
    ctx.current_path = input.currentPath;
  }
  if (input.selection) {
    ctx.selection = {
      file_path: input.selection.filePath,
      start: input.selection.start,
      end: input.selection.end,
      text: input.selection.text,
    };
  }
  if (input.currentView) {
    ctx.current_view = {
      rule_id: input.currentView.ruleId,
      version: input.currentView.version,
    };
  }
  return ctx;
}
