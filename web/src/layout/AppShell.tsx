import React, { useState, useEffect, useCallback } from "react";

type AppShellProps = {
  sidebar: React.ReactNode;
  main: React.ReactNode;
  rightSidebar?: React.ReactNode;
  rightCollapsed?: boolean;
  onToggleRight?: () => void;
  footer: React.ReactNode;
};

// Breakpoints
const MOBILE_BREAKPOINT = 768;
const TABLET_BREAKPOINT = 1024;

// Hook for responsive detection
function useResponsive() {
  const [isMobile, setIsMobile] = useState(false);
  const [isTablet, setIsTablet] = useState(false);

  useEffect(() => {
    const checkSize = () => {
      const width = window.innerWidth;
      setIsMobile(width < MOBILE_BREAKPOINT);
      setIsTablet(width >= MOBILE_BREAKPOINT && width < TABLET_BREAKPOINT);
    };

    checkSize();
    window.addEventListener("resize", checkSize);
    return () => window.removeEventListener("resize", checkSize);
  }, []);

  return { isMobile, isTablet };
}

const sidebarStyle: React.CSSProperties = {
  gridArea: "sidebar",
  borderRight: "1px solid var(--border-color)",
  overflow: "auto",
  background: "rgba(255,255,255,0.6)",
  backdropFilter: "blur(12px)",
  display: "flex",
  flexDirection: "column",
};

const mainStyle: React.CSSProperties = {
  gridArea: "main",
  overflow: "auto",
  padding: "0",
  background: "transparent",
  display: "flex",
  flexDirection: "column",
};

const footerStyle: React.CSSProperties = {
  gridArea: "footer",
  borderTop: "1px solid var(--border-color)",
  padding: "0 20px",
  display: "flex",
  alignItems: "center",
  gap: "12px",
  background: "rgba(255,255,255,0.8)",
  backdropFilter: "blur(12px)",
  fontSize: "13px",
  color: "var(--text-secondary)",
};

export function AppShell({
  sidebar,
  main,
  rightSidebar,
  rightCollapsed = false,
  onToggleRight,
  footer,
}: AppShellProps) {
  const { isMobile, isTablet } = useResponsive();
  const [mobileNav, setMobileNav] = useState<"files" | "main" | "sessions">("main");
  const [sidebarOpen, setSidebarOpen] = useState(false);

  const toggleSidebar = useCallback(() => {
    setSidebarOpen((prev) => !prev);
  }, []);

  // Mobile layout
  if (isMobile) {
    return (
      <div
        style={{
          display: "flex",
          flexDirection: "column",
          height: "100vh",
          background: "linear-gradient(120deg, #fdfbfb 0%, #ebedee 100%)",
        }}
      >
        {/* Mobile header */}
        <header
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            padding: "12px 16px",
            borderBottom: "1px solid var(--border-color)",
            background: "rgba(255,255,255,0.9)",
            backdropFilter: "blur(12px)",
          }}
        >
          <button
            type="button"
            onClick={toggleSidebar}
            style={{
              padding: "8px",
              border: "none",
              background: "transparent",
              fontSize: "20px",
              cursor: "pointer",
            }}
          >
            ☰
          </button>
          <span style={{ fontWeight: 600, fontSize: "16px" }}>MindFS</span>
          <button
            type="button"
            onClick={onToggleRight}
            style={{
              padding: "8px",
              border: "none",
              background: "transparent",
              fontSize: "20px",
              cursor: "pointer",
            }}
          >
            💬
          </button>
        </header>

        {/* Mobile content */}
        <div style={{ flex: 1, overflow: "auto" }}>
          {mobileNav === "files" && sidebar}
          {mobileNav === "main" && main}
          {mobileNav === "sessions" && rightSidebar}
        </div>

        {/* Mobile footer with action bar */}
        <div style={{ borderTop: "1px solid var(--border-color)" }}>{footer}</div>

        {/* Mobile bottom navigation */}
        <nav
          style={{
            display: "flex",
            borderTop: "1px solid var(--border-color)",
            background: "rgba(255,255,255,0.95)",
          }}
        >
          <MobileNavButton
            icon="📁"
            label="Files"
            active={mobileNav === "files"}
            onClick={() => setMobileNav("files")}
          />
          <MobileNavButton
            icon="🏠"
            label="View"
            active={mobileNav === "main"}
            onClick={() => setMobileNav("main")}
          />
          <MobileNavButton
            icon="💬"
            label="Sessions"
            active={mobileNav === "sessions"}
            onClick={() => setMobileNav("sessions")}
          />
        </nav>

        {/* Mobile sidebar overlay */}
        {sidebarOpen && (
          <div
            style={{
              position: "fixed",
              inset: 0,
              zIndex: 1000,
              display: "flex",
            }}
          >
            <div
              style={{
                position: "absolute",
                inset: 0,
                background: "rgba(0,0,0,0.3)",
              }}
              onClick={toggleSidebar}
            />
            <div
              style={{
                position: "relative",
                width: "280px",
                maxWidth: "80vw",
                background: "#fff",
                boxShadow: "4px 0 24px rgba(0,0,0,0.1)",
                overflow: "auto",
              }}
            >
              {sidebar}
            </div>
          </div>
        )}
      </div>
    );
  }

  // Tablet layout - narrower sidebar
  const sidebarWidth = isTablet ? "200px" : "260px";
  const rightWidth = rightSidebar ? (rightCollapsed ? "0px" : isTablet ? "240px" : "280px") : "0px";

  const shellStyle: React.CSSProperties = {
    display: "grid",
    gridTemplateColumns: `${sidebarWidth} 1fr ${rightWidth}`,
    gridTemplateRows: "1fr 48px",
    gridTemplateAreas: `"sidebar main right" "sidebar footer right"`,
    height: "100vh",
    background: "linear-gradient(120deg, #fdfbfb 0%, #ebedee 100%)",
    color: "var(--text-primary)",
  };

  const rightStyle: React.CSSProperties = {
    gridArea: "right",
    borderLeft: "1px solid var(--border-color)",
    overflow: "auto",
    background: "rgba(255,255,255,0.6)",
    backdropFilter: "blur(12px)",
    display: rightSidebar ? "flex" : "none",
    flexDirection: "column",
  };

  return (
    <div style={shellStyle}>
      <aside style={sidebarStyle}>{sidebar}</aside>
      <main style={mainStyle}>{main}</main>
      <aside style={rightStyle}>{rightSidebar}</aside>
      <footer style={footerStyle}>{footer}</footer>
      {rightSidebar && rightCollapsed ? (
        <button
          type="button"
          onClick={onToggleRight}
          style={{
            position: "fixed",
            right: 12,
            top: "50%",
            transform: "translateY(-50%)",
            padding: "6px 8px",
            borderRadius: "8px",
            border: "1px solid var(--border-color)",
            background: "#fff",
            fontSize: "12px",
            cursor: "pointer",
            boxShadow: "0 6px 16px rgba(15,23,42,0.12)",
          }}
        >
          会话
        </button>
      ) : null}
    </div>
  );
}

// Mobile navigation button component
function MobileNavButton({
  icon,
  label,
  active,
  onClick,
}: {
  icon: string;
  label: string;
  active: boolean;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      style={{
        flex: 1,
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        gap: "4px",
        padding: "8px",
        border: "none",
        background: "transparent",
        color: active ? "var(--accent-color)" : "var(--text-secondary)",
        fontSize: "10px",
        cursor: "pointer",
      }}
    >
      <span style={{ fontSize: "20px" }}>{icon}</span>
      <span>{label}</span>
    </button>
  );
}
