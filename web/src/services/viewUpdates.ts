export type ViewUpdate =
  | { type: "full"; payload: unknown }
  | { type: "patch"; payload: JsonPatch[] };

export type JsonPatch = {
  op: "add" | "remove" | "replace" | "set";
  path: string;
  value?: unknown;
};

export type ViewState = {
  current: unknown | null;
  pending: unknown | null;
  previous: unknown | null;
};

export const emptyViewState = (): ViewState => ({
  current: null,
  pending: null,
  previous: null,
});

const clone = <T,>(value: T): T =>
  value
    ? (typeof structuredClone === "function"
        ? structuredClone(value)
        : JSON.parse(JSON.stringify(value)))
    : value;

const pointerSegments = (path: string) =>
  path
    .split("/")
    .filter(Boolean)
    .map((seg) => seg.replace(/~1/g, "/").replace(/~0/g, "~"));

const setByPointer = (obj: any, path: string, value: unknown) => {
  const segs = pointerSegments(path);
  if (segs.length === 0) return value;
  let curr = obj;
  for (let i = 0; i < segs.length - 1; i++) {
    const key = segs[i];
    if (!(key in curr) || typeof curr[key] !== "object") {
      curr[key] = {};
    }
    curr = curr[key];
  }
  curr[segs[segs.length - 1]] = value;
  return obj;
};

const removeByPointer = (obj: any, path: string) => {
  const segs = pointerSegments(path);
  if (segs.length === 0) return obj;
  let curr = obj;
  for (let i = 0; i < segs.length - 1; i++) {
    const key = segs[i];
    if (!(key in curr)) return obj;
    curr = curr[key];
  }
  delete curr[segs[segs.length - 1]];
  return obj;
};

export function applyViewUpdate(current: unknown | null, update: ViewUpdate): unknown | null {
  if (update.type === "full") {
    return update.payload ?? null;
  }
  const base = clone(current ?? {});
  update.payload.forEach((patch) => {
    switch (patch.op) {
      case "add":
      case "replace":
      case "set":
        setByPointer(base, patch.path, patch.value);
        break;
      case "remove":
        removeByPointer(base, patch.path);
        break;
    }
  });
  return base;
}

export function stageViewUpdate(state: ViewState, update: ViewUpdate): ViewState {
  const next = applyViewUpdate(state.current, update);
  return {
    current: state.current,
    previous: state.current,
    pending: next,
  };
}

export function acceptViewUpdate(state: ViewState): ViewState {
  return {
    current: state.pending ?? state.current,
    pending: null,
    previous: null,
  };
}

export function revertViewUpdate(state: ViewState): ViewState {
  return {
    current: state.previous ?? state.current,
    pending: null,
    previous: null,
  };
}
