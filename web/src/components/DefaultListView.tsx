import React from "react";

type FileEntry = {
  name: string;
  path: string;
  is_dir: boolean;
};

type DefaultListViewProps = {
  entries: FileEntry[];
};

export function DefaultListView({ entries }: DefaultListViewProps) {
  return (
    <div style={{ padding: "32px 40px" }}>
      <header style={{ marginBottom: "32px" }}>
        <h2 style={{ margin: "0 0 8px 0", fontSize: "24px", fontWeight: 600 }}>
          Overview
        </h2>
        <p style={{ margin: 0, color: "var(--text-secondary)", fontSize: "14px" }}>
          {entries.length} items in this directory
        </p>
      </header>
      
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fill, minmax(200px, 1fr))",
          gap: "16px",
        }}
      >
        {entries.map((entry) => (
          <div
            key={entry.path}
            style={{
              background: "#fff",
              border: "1px solid var(--border-color)",
              borderRadius: "8px",
              padding: "16px",
              display: "flex",
              flexDirection: "column",
              gap: "12px",
              transition: "transform 0.1s, box-shadow 0.1s",
              cursor: "default",
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.transform = "translateY(-2px)";
              e.currentTarget.style.boxShadow = "0 4px 12px rgba(0,0,0,0.05)";
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.transform = "translateY(0)";
              e.currentTarget.style.boxShadow = "none";
            }}
          >
            <div
              style={{
                width: "40px",
                height: "40px",
                borderRadius: "8px",
                background: entry.is_dir ? "#eff6ff" : "#f1f5f9",
                color: entry.is_dir ? "#3b82f6" : "#64748b",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                fontSize: "20px",
              }}
            >
              {entry.is_dir ? "📂" : "📄"}
            </div>
            
            <div style={{ minWidth: 0 }}>
              <div
                style={{
                  fontWeight: 500,
                  fontSize: "14px",
                  marginBottom: "4px",
                  whiteSpace: "nowrap",
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                }}
                title={entry.name}
              >
                {entry.name}
              </div>
              <div
                style={{
                  fontSize: "12px",
                  color: "var(--text-secondary)",
                  whiteSpace: "nowrap",
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                }}
              >
                {entry.path}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}