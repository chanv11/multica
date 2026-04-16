# Skills Folder Upload Design

## Summary

Add a folder upload feature to the Skills page's create dialog, allowing users to select a local folder and automatically create a skill from its contents (SKILL.md as main content, remaining text files as supporting files).

## Motivation

Current skill creation requires manually entering content in the UI. Users with existing skill folders (e.g., from local development or exported configurations) must recreate files one by one. Folder upload eliminates this friction.

## Trigger Location

Skills page → Create Skill dialog → new "Upload" tab alongside existing Create / Import tabs.

## User Flow

1. User clicks "Create skill" button on Skills page
2. Dialog opens with three tabs: Create | Import | Upload
3. User selects "Upload" tab
4. Upload area shows a click-to-select zone
5. User clicks and selects a local folder via browser folder picker
6. Frontend reads all files, filters to text-only, validates SKILL.md presence
7. Preview displays: skill name (folder name, editable), description (optional), file list with remove option
8. User confirms → `api.createSkill()` called with parsed data
9. Success: dialog closes, skill list refreshes, new skill selected

## Technical Design

### File Reading & Filtering (Frontend Only)

- `<input type="file" webkitdirectory>` triggers browser folder picker
- Returns flat `File[]` with `webkitRelativePath` on each file
- **Binary filtering**: blacklist approach — skip files with known binary extensions (`.png`, `.jpg`, `.gif`, `.zip`, `.pdf`, `.exe`, `.bin`, `.woff`, `.ttf`, `.ico`, `.mp3`, `.mp4`, `.webp`, `.svg`, `.ogg`, `.wav`, `.avi`, `.mov`, `.woff2`, `.eot`)
- **SKILL.md detection**: scan all text files for one whose path ends with `SKILL.md`. Not found → show error "文件夹中未找到 SKILL.md 文件，请确保包含 SKILL.md 后重试"
- **Path normalization**: strip top-level folder name from `webkitRelativePath`. E.g., `my-skill/src/index.ts` → `src/index.ts`

### Limits

| Constraint | Value | Enforcement |
|---|---|---|
| Max files per upload | 50 | Frontend: reject with error message |
| Max single file size | 1 MB | Frontend: skip file, count in skipped |
| Max total size | 5 MB | Frontend: reject with error message |
| Text-only | Blacklist binary extensions | Frontend: skip, report skipped count |

### State Management

```ts
// upload-tab.tsx local state
const [files, setFiles] = useState<{ path: string; content: string }[]>([]);
const [skillContent, setSkillContent] = useState<string>("");
const [defaultName, setDefaultName] = useState<string>("");
const [skippedCount, setSkippedCount] = useState<number>(0);
const [reading, setReading] = useState<boolean>(false);
const [readError, setReadError] = useState<string | null>(null);
```

### API Usage

Reuses existing `api.createSkill()` with `CreateSkillRequest`:

```ts
api.createSkill({
  name: name || defaultName,   // folder name, user can edit
  description,                  // optional
  content: skillContent,        // SKILL.md content
  files: files,                 // remaining text files
});
```

No backend changes required.

## Component Structure

### File Changes

| File | Action | Description |
|---|---|---|
| `packages/views/skills/components/upload-tab.tsx` | **New** | Upload tab component |
| `packages/views/skills/components/skills-page.tsx` | **Modify** | Add Upload tab to CreateSkillDialog |

### upload-tab.tsx

**Props:**
```ts
interface UploadTabProps {
  onCreate: (data: CreateSkillRequest) => void;
  onCancel: () => void;
}
```

**Responsibilities:**
- Render click-to-select upload area
- Read folder files via `<input webkitdirectory>`
- Filter, validate, and parse file contents
- Display preview with editable name, description, and removable file list
- Call `onCreate` on confirm

**UI Elements:**
- Upload zone: bordered dashed area with folder icon and hint text, triggers hidden `<input webkitdirectory>` on click
- Preview section (shown after folder selected):
  - Skill name input (pre-filled with folder name)
  - Description textarea (optional)
  - Skipped files notice (if any binary files were skipped)
  - File list with X button per file to remove
- Action buttons: Cancel / Create Skill

### skills-page.tsx Changes

- Add `TabsTrigger value="upload"` to TabsList
- Add `TabsContent value="upload"` rendering `<UploadTab />`
- Tab type extended to `"create" | "import" | "upload"`
- Reuse existing `handleCreate` callback (no changes needed)

### UI Component Reuse

| Element | Pattern |
|---|---|
| Input / Textarea | shadcn Input / Textarea, same as Create tab |
| File list items | Simplified path display with X remove button, similar to file-tree.tsx path styling |
| Buttons | shadcn Button, consistent with dialog |
| Error display | Inline red text, same as Import tab `importError` pattern |
| Upload zone | Dashed border area with icon, similar to common drop zone patterns |

## Edge Cases

| Case | Behavior |
|---|---|
| No SKILL.md in folder | Show error, block creation |
| Empty folder (no text files) | Show error "文件夹中没有可识别的文本文件" |
| Folder with only binary files | Show error with skipped count |
| File read failure (encoding) | Skip file, include in skipped count |
| Very large folder (>50 files) | Show error "文件数量超过 50 个上限" |
| Duplicate path after normalization | Deduplicate, keep first occurrence |
| SKILL.md in subdirectory | Accept it (match by path suffix) |
