# Skills Folder Upload Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an Upload tab to the Skills create dialog that lets users select a local folder and auto-create a skill from its text files.

**Architecture:** New `upload-tab.tsx` component handles folder selection, file reading/filtering, and preview. It plugs into the existing `CreateSkillDialog` in `skills-page.tsx` alongside the Create and Import tabs. No backend changes — reuses `api.createSkill()`.

**Tech Stack:** React, TypeScript, shadcn/ui Tabs/Dialog/Input/Button, browser `<input webkitdirectory>` API.

---

## File Structure

| File | Action | Responsibility |
|---|---|---|
| `packages/views/skills/components/upload-tab.tsx` | **Create** | Upload tab: folder picker, file reading, filtering, preview, confirm |
| `packages/views/skills/components/upload-tab.test.tsx` | **Create** | Unit tests for upload-tab |
| `packages/views/skills/components/skills-page.tsx` | **Modify** | Add Upload tab trigger + content to CreateSkillDialog |

---

### Task 1: Create upload-tab.tsx — folder reading and filtering logic

**Files:**
- Create: `packages/views/skills/components/upload-tab.tsx`

- [ ] **Step 1: Create the upload-tab component with folder reading logic**

```tsx
"use client";

import { useState, useRef } from "react";
import { FolderOpen, X, AlertCircle, Upload } from "lucide-react";
import type { CreateSkillRequest } from "@multica/core/types";
import { Button } from "@multica/ui/components/ui/button";
import { Input } from "@multica/ui/components/ui/input";
import { Label } from "@multica/ui/components/ui/label";

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const BINARY_EXTENSIONS = new Set([
  ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".ico", ".svg", ".webp",
  ".mp3", ".mp4", ".wav", ".ogg", ".avi", ".mov", ".webm",
  ".zip", ".tar", ".gz", ".rar", ".7z",
  ".pdf", ".doc", ".docx", ".ppt", ".pptx", ".xls", ".xlsx",
  ".exe", ".dll", ".so", ".dylib", ".bin",
  ".woff", ".woff2", ".ttf", ".eot", ".otf",
  ".sqlite", ".db",
]);

const MAX_FILES = 50;
const MAX_SINGLE_FILE_SIZE = 1 * 1024 * 1024; // 1 MB
const MAX_TOTAL_SIZE = 5 * 1024 * 1024; // 5 MB

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function getExtension(filename: string): string {
  const dot = filename.lastIndexOf(".");
  return dot >= 0 ? filename.slice(dot).toLowerCase() : "";
}

function isBinaryFile(filename: string): boolean {
  return BINARY_EXTENSIONS.has(getExtension(filename));
}

interface ParsedFolder {
  name: string;
  skillContent: string;
  files: { path: string; content: string }[];
  skippedCount: number;
}

function parseFolderFiles(fileList: FileList): Promise<ParsedFolder> {
  const files = Array.from(fileList);

  // Derive folder name from first file's relative path
  const firstPath = files[0]?.webkitRelativePath ?? "";
  const folderName = firstPath.split("/")[0] ?? "untitled";

  // Filter binary files
  const textFiles = files.filter((f) => !isBinaryFile(f.name));
  const skippedCount = files.length - textFiles.length;

  // Check file count
  if (textFiles.length > MAX_FILES) {
    throw new Error(`文件数量超过 ${MAX_FILES} 个上限（共 ${textFiles.length} 个文本文件）`);
  }

  // Check total size
  const totalSize = textFiles.reduce((sum, f) => sum + f.size, 0);
  if (totalSize > MAX_TOTAL_SIZE) {
    throw new Error(`文件总大小超过 5 MB 上限（共 ${(totalSize / 1024 / 1024).toFixed(1)} MB）`);
  }

  // Read all text files
  return Promise.all(
    textFiles.map(
      (f) =>
        new Promise<{ file: File; content: string | null }>((resolve) => {
          if (f.size > MAX_SINGLE_FILE_SIZE) {
            resolve({ file: f, content: null });
            return;
          }
          const reader = new FileReader();
          reader.onload = () =>
            resolve({ file: f, content: reader.result as string });
          reader.onerror = () => resolve({ file: f, content: null });
          reader.readAsText(f);
        }),
    ),
  ).then((results) => {
    let skillContent = "";
    const supportingFiles: { path: string; content: string }[] = [];
    let readSkipped = 0;

    for (const { file, content } of results) {
      if (content === null) {
        readSkipped++;
        continue;
      }
      // Strip top-level folder name from path
      const relativePath = file.webkitRelativePath.split("/").slice(1).join("/");

      if (relativePath === "SKILL.md" || relativePath.endsWith("/SKILL.md")) {
        skillContent = content;
      } else {
        supportingFiles.push({ path: relativePath, content });
      }
    }

    return {
      name: folderName,
      skillContent,
      files: supportingFiles,
      skippedCount: skippedCount + readSkipped,
    };
  });
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

interface UploadTabProps {
  onCreate: (data: CreateSkillRequest) => Promise<void>;
  onCancel: () => void;
}

export function UploadTab({ onCreate, onCancel }: UploadTabProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [files, setFiles] = useState<{ path: string; content: string }[]>([]);
  const [skillContent, setSkillContent] = useState<string | null>(null);
  const [skippedCount, setSkippedCount] = useState(0);
  const [reading, setReading] = useState(false);
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [folderSelected, setFolderSelected] = useState(false);

  const handleFolderSelect = () => {
    inputRef.current?.click();
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const fileList = e.target.files;
    if (!fileList || fileList.length === 0) return;

    setReading(true);
    setError(null);

    parseFolderFiles(fileList)
      .then((result) => {
        if (!result.skillContent) {
          setError("文件夹中未找到 SKILL.md 文件，请确保包含 SKILL.md 后重试");
          setFolderSelected(false);
          return;
        }
        setName(result.name);
        setSkillContent(result.skillContent);
        setFiles(result.files);
        setSkippedCount(result.skippedCount);
        setFolderSelected(true);
      })
      .catch((err) => {
        setError(err instanceof Error ? err.message : "读取文件夹失败");
        setFolderSelected(false);
      })
      .finally(() => {
        setReading(false);
        // Reset input so the same folder can be re-selected
        e.target.value = "";
      });
  };

  const handleRemoveFile = (path: string) => {
    setFiles((prev) => prev.filter((f) => f.path !== path));
  };

  const handleCreate = async () => {
    if (skillContent === null) return;
    setCreating(true);
    try {
      await onCreate({
        name: name.trim(),
        description: description.trim() || undefined,
        content: skillContent,
        files,
      });
    } catch {
      setCreating(false);
    }
  };

  return (
    <div className="space-y-4 mt-4 min-h-[180px]">
      {/* Hidden folder input */}
      <input
        ref={inputRef}
        type="file"
        className="hidden"
        {...({ webkitdirectory: "", directory: "" } as Record<string, string>)}
        onChange={handleInputChange}
      />

      {!folderSelected ? (
        /* Upload zone */
        <button
          type="button"
          onClick={handleFolderSelect}
          disabled={reading}
          className="flex w-full flex-col items-center justify-center gap-2 rounded-lg border-2 border-dashed border-muted-foreground/25 px-6 py-8 text-muted-foreground transition-colors hover:border-muted-foreground/50 hover:bg-accent/50 disabled:opacity-50"
        >
          {reading ? (
            <>
              <Upload className="h-6 w-6 animate-pulse" />
              <span className="text-sm">读取文件中...</span>
            </>
          ) : (
            <>
              <FolderOpen className="h-6 w-6" />
              <span className="text-sm font-medium">点击选择文件夹</span>
              <span className="text-xs">
                需包含 SKILL.md，支持 .md .ts .py .json 等文本文件
              </span>
            </>
          )}
        </button>
      ) : (
        /* Preview */
        <>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <Label className="text-xs text-muted-foreground">名称</Label>
              <Input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="mt-1"
              />
            </div>
            <div>
              <Label className="text-xs text-muted-foreground">描述</Label>
              <Input
                type="text"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                className="mt-1"
                placeholder="可选"
              />
            </div>
          </div>

          {skippedCount > 0 && (
            <div className="flex items-center gap-2 rounded-md bg-yellow-500/10 px-3 py-1.5 text-xs text-yellow-600">
              <AlertCircle className="h-3.5 w-3.5 shrink-0" />
              已跳过 {skippedCount} 个非文本文件
            </div>
          )}

          <div>
            <div className="flex items-center justify-between mb-2">
              <Label className="text-xs text-muted-foreground">
                文件列表 ({files.length + 1} 个文件)
              </Label>
              <Button
                variant="ghost"
                size="xs"
                onClick={() => {
                  setFolderSelected(false);
                  setName("");
                  setDescription("");
                  setFiles([]);
                  setSkillContent(null);
                  setSkippedCount(0);
                  setError(null);
                }}
              >
                重新选择
              </Button>
            </div>
            <div className="max-h-40 overflow-y-auto rounded-lg border">
              {/* SKILL.md always first */}
              <div className="flex items-center gap-2 px-3 py-1.5 text-xs border-b last:border-b-0">
                <span className="font-mono truncate flex-1">SKILL.md</span>
                <span className="text-muted-foreground text-[11px]">主文件</span>
              </div>
              {files.map((f) => (
                <div
                  key={f.path}
                  className="flex items-center gap-2 px-3 py-1.5 text-xs border-b last:border-b-0"
                >
                  <span className="font-mono truncate flex-1">{f.path}</span>
                  <button
                    type="button"
                    onClick={() => handleRemoveFile(f.path)}
                    className="shrink-0 rounded p-0.5 text-muted-foreground hover:text-destructive"
                  >
                    <X className="h-3 w-3" />
                  </button>
                </div>
              ))}
            </div>
          </div>
        </>
      )}

      {/* Error display */}
      {error && (
        <div className="flex items-center gap-2 rounded-md bg-destructive/10 px-3 py-2 text-xs text-destructive">
          <AlertCircle className="h-3.5 w-3.5 shrink-0" />
          {error}
        </div>
      )}

      {/* Footer buttons */}
      <div className="flex justify-end gap-2 pt-2">
        <Button variant="ghost" onClick={onCancel}>
          取消
        </Button>
        <Button
          onClick={handleCreate}
          disabled={creating || !folderSelected || skillContent === null || !name.trim()}
        >
          {creating ? "创建中..." : "创建 Skill"}
        </Button>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Commit the upload tab component**

```bash
git add packages/views/skills/components/upload-tab.tsx
git commit -m "feat(views): add upload-tab component for skills folder upload"
```

---

### Task 2: Write tests for upload-tab

**Files:**
- Create: `packages/views/skills/components/upload-tab.test.tsx`

- [ ] **Step 1: Write the test file**

```tsx
import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { UploadTab } from "./upload-tab";

// Mock @multica/core/types — just re-export since UploadTab only uses the type
vi.mock("@multica/core/types", () => ({}));
// Mock @multica/ui components
vi.mock("@multica/ui/components/ui/button", () => ({
  Button: ({ children, onClick, disabled, ...props }: any) => (
    <button onClick={onClick} disabled={disabled} {...props}>
      {children}
    </button>
  ),
}));
vi.mock("@multica/ui/components/ui/input", () => ({
  Input: (props: any) => <input {...props} />,
}));
vi.mock("@multica/ui/components/ui/label", () => ({
  Label: ({ children, ...props }: any) => <label {...props}>{children}</label>,
}));

describe("UploadTab", () => {
  it("renders the upload zone initially", () => {
    render(<UploadTab onCreate={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByText("点击选择文件夹")).toBeInTheDocument();
  });

  it("shows cancel and create buttons", () => {
    render(<UploadTab onCreate={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByText("取消")).toBeInTheDocument();
    expect(screen.getByText("创建 Skill")).toBeInTheDocument();
  });

  it("create button is disabled before folder selection", () => {
    render(<UploadTab onCreate={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByText("创建 Skill")).toBeDisabled();
  });

  it("calls onCancel when cancel is clicked", () => {
    const onCancel = vi.fn();
    render(<UploadTab onCreate={vi.fn()} onCancel={onCancel} />);
    fireEvent.click(screen.getByText("取消"));
    expect(onCancel).toHaveBeenCalledOnce();
  });

  it("shows error when no SKILL.md is found", async () => {
    const onCreate = vi.fn();
    render(<UploadTab onCreate={onCreate} onCancel={vi.fn()} />);

    // Simulate file input change with files that don't include SKILL.md
    const file = new File(["hello"], "readme.txt", {
      type: "text/plain",
    });
    Object.defineProperty(file, "webkitRelativePath", {
      value: "my-skill/readme.txt",
      writable: false,
    });

    const input = document.querySelector('input[type="file"]')!;
    fireEvent.change(input, { target: { files: [file] } });

    await waitFor(() => {
      expect(
        screen.getByText(/文件夹中未找到 SKILL.md/),
      ).toBeInTheDocument();
    });
    expect(onCreate).not.toHaveBeenCalled();
  });

  it("calls onCreate with parsed data when folder with SKILL.md is selected", async () => {
    const onCreate = vi.fn().mockResolvedValue(undefined);
    render(<UploadTab onCreate={onCreate} onCancel={vi.fn()} />);

    const skillMd = new File("# My Skill\n\nDoes things.", "SKILL.md", {
      type: "text/markdown",
    });
    Object.defineProperty(skillMd, "webkitRelativePath", {
      value: "my-skill/SKILL.md",
      writable: false,
    });

    const helper = new File("export const x = 1;", "helper.ts", {
      type: "text/typescript",
    });
    Object.defineProperty(helper, "webkitRelativePath", {
      value: "my-skill/src/helper.ts",
      writable: false,
    });

    const input = document.querySelector('input[type="file"]')!;
    fireEvent.change(input, { target: { files: [skillMd, helper] } });

    await waitFor(() => {
      expect(screen.getByText("创建 Skill")).toBeEnabled();
    });

    fireEvent.click(screen.getByText("创建 Skill"));

    await waitFor(() => {
      expect(onCreate).toHaveBeenCalledOnce();
    });

    const call = onCreate.mock.calls[0][0];
    expect(call.name).toBe("my-skill");
    expect(call.content).toBe("# My Skill\n\nDoes things.");
    expect(call.files).toHaveLength(1);
    expect(call.files[0].path).toBe("src/helper.ts");
  });

  it("allows removing files from the list", async () => {
    const onCreate = vi.fn().mockResolvedValue(undefined);
    render(<UploadTab onCreate={onCreate} onCancel={vi.fn()} />);

    const skillMd = new File("skill content", "SKILL.md", {
      type: "text/markdown",
    });
    Object.defineProperty(skillMd, "webkitRelativePath", {
      value: "my-skill/SKILL.md",
      writable: false,
    });

    const extra = new File("extra", "extra.txt", { type: "text/plain" });
    Object.defineProperty(extra, "webkitRelativePath", {
      value: "my-skill/extra.txt",
      writable: false,
    });

    const input = document.querySelector('input[type="file"]')!;
    fireEvent.change(input, { target: { files: [skillMd, extra] } });

    await waitFor(() => {
      expect(screen.getByText("extra.txt")).toBeInTheDocument();
    });

    // Click the X button next to extra.txt
    const removeBtn = screen.getByText("extra.txt")
      .closest("div")!
      .querySelector("button")!;
    fireEvent.click(removeBtn);

    expect(screen.queryByText("extra.txt")).not.toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run the tests**

Run: `pnpm --filter @multica/views exec vitest run skills/components/upload-tab.test.tsx`
Expected: All tests pass.

- [ ] **Step 3: Commit tests**

```bash
git add packages/views/skills/components/upload-tab.test.tsx
git commit -m "test(views): add upload-tab component tests"
```

---

### Task 3: Integrate Upload tab into CreateSkillDialog

**Files:**
- Modify: `packages/views/skills/components/skills-page.tsx`

- [ ] **Step 1: Add the Upload import**

Add to the imports at the top of `skills-page.tsx`:

```tsx
import { UploadTab } from "./upload-tab";
```

- [ ] **Step 2: Extend the tab state type and add Upload tab trigger**

In `CreateSkillDialog`, change the tab state from `"create" | "import"` to `"create" | "import" | "upload"`:

```tsx
const [tab, setTab] = useState<"create" | "import" | "upload">("create");
```

Add `Upload` icon to imports:

```tsx
import { Sparkles, Plus, Trash2, Save, AlertCircle, Download, FolderUp } from "lucide-react";
```

Add the Upload trigger to the TabsList (after the Import trigger):

```tsx
<TabsTrigger value="upload" className="flex-1">
  <FolderUp className="mr-1.5 h-3 w-3" />
  Upload
</TabsTrigger>
```

- [ ] **Step 3: Add Upload TabsContent**

Add after the Import TabsContent:

```tsx
<TabsContent value="upload" className="mt-0">
  <UploadTab onCreate={onCreate} onCancel={onClose} />
</TabsContent>
```

- [ ] **Step 4: Update DialogFooter to handle upload tab**

The DialogFooter currently shows different buttons for create vs import. With upload tab, the upload tab has its own footer buttons, so hide the dialog footer when upload tab is active:

```tsx
{tab !== "upload" && (
  <DialogFooter>
    {/* ... existing create/import buttons ... */}
  </DialogFooter>
)}
```

- [ ] **Step 5: Commit the integration**

```bash
git add packages/views/skills/components/skills-page.tsx
git commit -m "feat(views): integrate upload tab into skills create dialog"
```

---

### Task 4: Manual verification

- [ ] **Step 1: Run typecheck**

Run: `pnpm typecheck`
Expected: No errors.

- [ ] **Step 2: Run the full test suite**

Run: `pnpm --filter @multica/views exec vitest run`
Expected: All tests pass.

- [ ] **Step 3: Start dev server and manually test**

Run: `pnpm dev:web`

Verify in browser:
1. Navigate to Skills page
2. Click "Create skill" — dialog shows 3 tabs: Create, Import, Upload
3. Click "Upload" tab — shows click-to-select zone
4. Select a folder without SKILL.md — shows error
5. Select a folder with SKILL.md and other files — shows preview with file list
6. Remove a file, edit name — changes reflected
7. Click "创建 Skill" — skill created, dialog closes
