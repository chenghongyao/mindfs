import React, { useState, useRef, useEffect, useCallback } from "react";

export type Skill = {
  id: string;
  name: string;
  description: string;
  source: "agent" | "directory";
  agent?: string;
};

type SkillSelectorProps = {
  skills: Skill[];
  selectedSkill: Skill | null;
  onSelect: (skill: Skill) => void;
  disabled?: boolean;
};

export function SkillSelector({
  skills,
  selectedSkill,
  onSelect,
  disabled,
}: SkillSelectorProps): JSX.Element {
  const [isOpen, setIsOpen] = useState(false);
  const [search, setSearch] = useState("");
  const dropdownRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  // 点击外部关闭
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };
    if (isOpen) {
      document.addEventListener("mousedown", handleClickOutside);
      return () => document.removeEventListener("mousedown", handleClickOutside);
    }
  }, [isOpen]);

  // 打开时聚焦搜索框
  useEffect(() => {
    if (isOpen && inputRef.current) {
      inputRef.current.focus();
    }
  }, [isOpen]);

  const handleSelect = useCallback(
    (skill: Skill) => {
      onSelect(skill);
      setIsOpen(false);
      setSearch("");
    },
    [onSelect]
  );

  // 按来源分组
  const agentSkills = skills.filter((s) => s.source === "agent");
  const directorySkills = skills.filter((s) => s.source === "directory");

  // 搜索过滤
  const filterSkills = (list: Skill[]) =>
    list.filter(
      (s) =>
        s.name.toLowerCase().includes(search.toLowerCase()) ||
        s.description.toLowerCase().includes(search.toLowerCase())
    );

  const filteredAgentSkills = filterSkills(agentSkills);
  const filteredDirectorySkills = filterSkills(directorySkills);

  return (
    <div ref={dropdownRef} style={{ position: "relative" }}>
      {/* 触发按钮 */}
      <button
        type="button"
        onClick={() => !disabled && setIsOpen(!isOpen)}
        disabled={disabled}
        style={{
          display: "flex",
          alignItems: "center",
          gap: "8px",
          padding: "8px 12px",
          borderRadius: "8px",
          border: "1px solid var(--border-color)",
          background: "#fff",
          cursor: disabled ? "not-allowed" : "pointer",
          fontSize: "13px",
          color: "var(--text-primary)",
          minWidth: "180px",
          opacity: disabled ? 0.6 : 1,
        }}
      >
        <span>⚡</span>
        <span style={{ flex: 1, textAlign: "left" }}>
          {selectedSkill ? selectedSkill.name : "选择技能..."}
        </span>
        <span
          style={{
            fontSize: "10px",
            color: "var(--text-secondary)",
            transform: isOpen ? "rotate(180deg)" : "rotate(0deg)",
            transition: "transform 0.15s",
          }}
        >
          ▼
        </span>
      </button>

      {/* 下拉面板 */}
      {isOpen && (
        <div
          style={{
            position: "absolute",
            top: "calc(100% + 4px)",
            left: 0,
            background: "#fff",
            border: "1px solid var(--border-color)",
            borderRadius: "12px",
            boxShadow: "0 8px 24px rgba(0,0,0,0.12)",
            zIndex: 100,
            minWidth: "280px",
            maxHeight: "360px",
            overflow: "hidden",
            display: "flex",
            flexDirection: "column",
          }}
        >
          {/* 搜索框 */}
          <div style={{ padding: "8px", borderBottom: "1px solid var(--border-color)" }}>
            <input
              ref={inputRef}
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="搜索技能..."
              style={{
                width: "100%",
                padding: "8px 10px",
                borderRadius: "6px",
                border: "1px solid var(--border-color)",
                fontSize: "13px",
                outline: "none",
              }}
            />
          </div>

          {/* 技能列表 */}
          <div style={{ flex: 1, overflow: "auto", padding: "4px 0" }}>
            {/* Agent 能力 */}
            {filteredAgentSkills.length > 0 && (
              <div>
                <div
                  style={{
                    padding: "8px 12px 4px",
                    fontSize: "11px",
                    fontWeight: 600,
                    color: "var(--text-secondary)",
                    textTransform: "uppercase",
                  }}
                >
                  Agent 能力
                </div>
                {filteredAgentSkills.map((skill) => (
                  <SkillItem
                    key={skill.id}
                    skill={skill}
                    selected={selectedSkill?.id === skill.id}
                    onSelect={handleSelect}
                  />
                ))}
              </div>
            )}

            {/* 目录技能 */}
            {filteredDirectorySkills.length > 0 && (
              <div>
                <div
                  style={{
                    padding: "8px 12px 4px",
                    fontSize: "11px",
                    fontWeight: 600,
                    color: "var(--text-secondary)",
                    textTransform: "uppercase",
                    borderTop:
                      filteredAgentSkills.length > 0 ? "1px solid var(--border-color)" : "none",
                    marginTop: filteredAgentSkills.length > 0 ? "4px" : 0,
                  }}
                >
                  当前目录
                </div>
                {filteredDirectorySkills.map((skill) => (
                  <SkillItem
                    key={skill.id}
                    skill={skill}
                    selected={selectedSkill?.id === skill.id}
                    onSelect={handleSelect}
                  />
                ))}
              </div>
            )}

            {/* 无结果 */}
            {filteredAgentSkills.length === 0 && filteredDirectorySkills.length === 0 && (
              <div
                style={{
                  padding: "20px",
                  textAlign: "center",
                  color: "var(--text-secondary)",
                  fontSize: "13px",
                }}
              >
                {search ? "未找到匹配的技能" : "暂无可用技能"}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

type SkillItemProps = {
  skill: Skill;
  selected: boolean;
  onSelect: (skill: Skill) => void;
};

function SkillItem({ skill, selected, onSelect }: SkillItemProps): JSX.Element {
  return (
    <button
      type="button"
      onClick={() => onSelect(skill)}
      style={{
        display: "flex",
        flexDirection: "column",
        gap: "2px",
        width: "100%",
        padding: "10px 12px",
        border: "none",
        background: selected ? "rgba(59, 130, 246, 0.08)" : "transparent",
        cursor: "pointer",
        textAlign: "left",
      }}
      onMouseEnter={(e) => {
        if (!selected) e.currentTarget.style.background = "rgba(0,0,0,0.03)";
      }}
      onMouseLeave={(e) => {
        if (!selected) e.currentTarget.style.background = "transparent";
      }}
    >
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: "8px",
        }}
      >
        <span
          style={{
            fontSize: "13px",
            fontWeight: selected ? 500 : 400,
            color: selected ? "#3b82f6" : "var(--text-primary)",
          }}
        >
          {skill.name}
        </span>
        {skill.agent && (
          <span
            style={{
              fontSize: "10px",
              padding: "2px 6px",
              borderRadius: "4px",
              background: "rgba(0,0,0,0.05)",
              color: "var(--text-secondary)",
            }}
          >
            {skill.agent}
          </span>
        )}
        {selected && (
          <span style={{ marginLeft: "auto", fontSize: "12px", color: "#3b82f6" }}>✓</span>
        )}
      </div>
      <div
        style={{
          fontSize: "11px",
          color: "var(--text-secondary)",
          overflow: "hidden",
          textOverflow: "ellipsis",
          whiteSpace: "nowrap",
        }}
      >
        {skill.description}
      </div>
    </button>
  );
}
