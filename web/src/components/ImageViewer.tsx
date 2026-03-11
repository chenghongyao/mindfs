import React from "react";

type ImageViewerProps = {
  path: string;
  root?: string;
};

export function ImageViewer({ path, root }: ImageViewerProps) {
  const url = root
    ? `/api/file?raw=1&root=${encodeURIComponent(root)}&path=${encodeURIComponent(path)}`
    : `/api/file?raw=1&path=${encodeURIComponent(path)}`;
  return (
    <div
      style={{
        padding: "24px",
        display: "flex",
        flex: 1,
        minHeight: 0,
        justifyContent: "center",
        alignItems: "center",
      }}
    >
      <img
        src={url}
        alt={path}
        style={{
          maxWidth: "100%",
          maxHeight: "100%",
          borderRadius: "12px",
          boxShadow: "0 12px 24px rgba(31, 37, 48, 0.1)",
        }}
      />
    </div>
  );
}
