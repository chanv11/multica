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
  skillContent: string | undefined;
  files: { path: string; content: string }[];
  skippedCount: number;
}

async function parseFolderFiles(fileList: FileList): Promise<ParsedFolder> {
  const files = Array.from(fileList);

  // Derive folder name from first file's relative path
  const firstPath = files[0]?.webkitRelativePath ?? "";
  const folderName = firstPath.split("/")[0] ?? "untitled";

  // Filter binary files
  const textFiles = files.filter((f) => !isBinaryFile(f.name));
  const skippedCount = files.length - textFiles.length;

  // Check empty folder
  if (textFiles.length === 0) {
    throw new Error("No recognized text files in the folder");
  }

  // Check file count
  if (textFiles.length > MAX_FILES) {
    throw new Error(`File count exceeds limit of ${MAX_FILES} (found ${textFiles.length} text files)`);
  }

  // Check total size
  const totalSize = textFiles.reduce((sum, f) => sum + f.size, 0);
  if (totalSize > MAX_TOTAL_SIZE) {
    throw new Error(`Total size exceeds 5 MB limit (${(totalSize / 1024 / 1024).toFixed(1)} MB)`);
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
    let skillContent: string | undefined;
    const supportingFiles: { path: string; content: string }[] = [];
    const seenPaths = new Set<string>();
    let readSkipped = 0;

    for (const { file, content } of results) {
      if (content === null) {
        readSkipped++;
        continue;
      }
      // Strip top-level folder name from path
      const relativePath = file.webkitRelativePath.split("/").slice(1).join("/");

      // Deduplicate by normalized path
      if (seenPaths.has(relativePath)) continue;
      seenPaths.add(relativePath);

      // Only match root-level SKILL.md — subdirectory SKILL.md files
      // are treated as regular supporting files
      if (relativePath === "SKILL.md") {
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
  const [skillContent, setSkillContent] = useState<string | undefined>(undefined);
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
        if (result.skillContent === undefined) {
          setError("No SKILL.md found in folder. Please include a SKILL.md file and try again");
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
        setError(err instanceof Error ? err.message : "Failed to read folder");
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
    if (skillContent === undefined) return;
    setCreating(true);
    try {
      await onCreate({
        name: name.trim(),
        description: description.trim() || undefined,
        content: skillContent,
        files,
      });
    } catch {
      // Error handled by parent component; keep dialog open for retry
    } finally {
      setCreating(false);
    }
  };

  return (
    <div className="space-y-4 mt-4 min-h-[180px]">
      {/* Hidden folder input — webkitdirectory enables folder selection (non-standard but widely supported) */}
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
              <span className="text-sm">Reading files...</span>
            </>
          ) : (
            <>
              <FolderOpen className="h-6 w-6" />
              <span className="text-sm font-medium">Click to select folder</span>
              <span className="text-xs">
                Must contain SKILL.md. Supports .md .ts .py .json and other text files
              </span>
            </>
          )}
        </button>
      ) : (
        /* Preview */
        <>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <Label className="text-xs text-muted-foreground">Name</Label>
              <Input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="mt-1"
              />
            </div>
            <div>
              <Label className="text-xs text-muted-foreground">Description</Label>
              <Input
                type="text"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                className="mt-1"
                placeholder="Optional"
              />
            </div>
          </div>

          {skippedCount > 0 && (
            <div className="flex items-center gap-2 rounded-md bg-yellow-500/10 px-3 py-1.5 text-xs text-yellow-600">
              <AlertCircle className="h-3.5 w-3.5 shrink-0" />
              Skipped {skippedCount} non-text file{skippedCount !== 1 ? "s" : ""}
            </div>
          )}

          <div>
            <div className="flex items-center justify-between mb-2">
              <Label className="text-xs text-muted-foreground">
                Files ({files.length + 1} file{files.length + 1 !== 1 ? "s" : ""})
              </Label>
              <Button
                variant="ghost"
                size="xs"
                onClick={() => {
                  setFolderSelected(false);
                  setName("");
                  setDescription("");
                  setFiles([]);
                  setSkillContent(undefined);
                  setSkippedCount(0);
                  setError(null);
                }}
              >
                Reselect
              </Button>
            </div>
            <div className="max-h-40 overflow-y-auto rounded-lg border">
              {/* SKILL.md always first */}
              <div className="flex items-center gap-2 px-3 py-1.5 text-xs border-b last:border-b-0">
                <span className="font-mono truncate flex-1">SKILL.md</span>
                <span className="text-muted-foreground text-[11px]">Main file</span>
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
          Cancel
        </Button>
        <Button
          onClick={handleCreate}
          disabled={creating || !folderSelected || skillContent === undefined || !name.trim()}
        >
          {creating ? "Creating..." : "Create Skill"}
        </Button>
      </div>
    </div>
  );
}
