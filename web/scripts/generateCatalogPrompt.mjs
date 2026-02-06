import { createCatalog, generateCatalogPrompt } from "@json-render/core";
import { z } from "zod";

const fileEntrySchema = z.object({
  name: z.string(),
  path: z.string(),
  is_dir: z.boolean(),
});

const catalog = createCatalog({
  components: {
    Shell: { props: z.object({}), hasChildren: true },
    Sidebar: { props: z.object({ slot: z.literal("sidebar").optional() }), hasChildren: true },
    Main: { props: z.object({ slot: z.literal("main").optional() }), hasChildren: true },
    Footer: { props: z.object({ slot: z.literal("footer").optional() }), hasChildren: true },
    FileTree: {
      props: z.object({
        entries: z.array(fileEntrySchema),
        childrenByPath: z.record(z.array(fileEntrySchema)),
        expanded: z.array(z.string()),
        selectedDir: z.string().nullable().optional(),
        rootId: z.string().nullable().optional(),
        managedRoots: z.array(z.string()).optional(),
      }),
    },
    DefaultListView: { props: z.object({ entries: z.array(fileEntrySchema) }) },
    FileViewer: {
      props: z.object({
        file: z
          .object({
            name: z.string(),
            path: z.string(),
            content: z.string(),
            encoding: z.string(),
            truncated: z.boolean(),
            size: z.number(),
            ext: z.string().optional(),
            mime: z.string().optional(),
          })
          .nullable()
          .optional(),
      }),
    },
    ActionBar: {
      props: z.object({
        status: z.string().optional(),
        pendingView: z.boolean().optional(),
        onAcceptView: z.any().optional(),
        onRevertView: z.any().optional(),
      }),
    },
  },
  actions: {
    open: { params: z.object({ path: z.string(), root: z.string().optional() }) },
    open_dir: { params: z.object({ path: z.string(), root: z.string().optional() }) },
  },
});

const prompt = generateCatalogPrompt(catalog);
process.stdout.write(prompt);
