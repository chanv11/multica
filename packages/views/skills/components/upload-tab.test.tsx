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

function makeFile(name: string, content: string, relativePath: string, type = "text/plain") {
  const file = new File([content], name, { type });
  Object.defineProperty(file, "webkitRelativePath", { value: relativePath, writable: false });
  return file;
}

function selectFiles(...files: File[]) {
  const input = document.querySelector('input[type="file"]') as HTMLInputElement;
  fireEvent.change(input, { target: { files } });
}

describe("UploadTab", () => {
  it("renders the upload zone initially", () => {
    render(<UploadTab onCreate={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByText("Click to select folder")).toBeInTheDocument();
  });

  it("shows cancel and create buttons", () => {
    render(<UploadTab onCreate={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByText("Cancel")).toBeInTheDocument();
    expect(screen.getByText("Create Skill")).toBeInTheDocument();
  });

  it("create button is disabled before folder selection", () => {
    render(<UploadTab onCreate={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByText("Create Skill")).toBeDisabled();
  });

  it("calls onCancel when cancel is clicked", () => {
    const onCancel = vi.fn();
    render(<UploadTab onCreate={vi.fn()} onCancel={onCancel} />);
    fireEvent.click(screen.getByText("Cancel"));
    expect(onCancel).toHaveBeenCalledOnce();
  });

  it("shows error when no SKILL.md is found", async () => {
    const onCreate = vi.fn();
    render(<UploadTab onCreate={onCreate} onCancel={vi.fn()} />);

    const file = makeFile("readme.txt", "hello", "my-skill/readme.txt");
    selectFiles(file);

    await waitFor(() => {
      expect(screen.getByText(/No SKILL.md found/)).toBeInTheDocument();
    });
    expect(onCreate).not.toHaveBeenCalled();
  });

  it("calls onCreate with parsed data when folder with SKILL.md is selected", async () => {
    const onCreate = vi.fn().mockResolvedValue(undefined);
    render(<UploadTab onCreate={onCreate} onCancel={vi.fn()} />);

    const skillMd = makeFile("SKILL.md", "# My Skill\n\nDoes things.", "my-skill/SKILL.md", "text/markdown");
    const helper = makeFile("helper.ts", "export const x = 1;", "my-skill/src/helper.ts", "text/typescript");
    selectFiles(skillMd, helper);

    await waitFor(() => {
      expect(screen.getByText("Create Skill")).toBeEnabled();
    });

    fireEvent.click(screen.getByText("Create Skill"));

    await waitFor(() => {
      expect(onCreate).toHaveBeenCalledOnce();
    });

    const call = onCreate.mock.calls[0]![0];
    expect(call.name).toBe("my-skill");
    expect(call.content).toBe("# My Skill\n\nDoes things.");
    expect(call.files).toHaveLength(1);
    expect(call.files[0].path).toBe("src/helper.ts");
  });

  it("allows removing files from the list", async () => {
    render(<UploadTab onCreate={vi.fn().mockResolvedValue(undefined)} onCancel={vi.fn()} />);

    const skillMd = makeFile("SKILL.md", "skill content", "my-skill/SKILL.md", "text/markdown");
    const extra = makeFile("extra.txt", "extra", "my-skill/extra.txt");
    selectFiles(skillMd, extra);

    await waitFor(() => {
      expect(screen.getByText("extra.txt")).toBeInTheDocument();
    });

    const removeBtn = screen.getByText("extra.txt")
      .closest("div")!
      .querySelector("button")!;
    fireEvent.click(removeBtn);

    expect(screen.queryByText("extra.txt")).not.toBeInTheDocument();
  });

  it("shows skipped count for binary files", async () => {
    render(<UploadTab onCreate={vi.fn().mockResolvedValue(undefined)} onCancel={vi.fn()} />);

    const skillMd = makeFile("SKILL.md", "content", "my-skill/SKILL.md", "text/markdown");
    const png = makeFile("image.png", "binary", "my-skill/image.png", "image/png");
    selectFiles(skillMd, png);

    await waitFor(() => {
      expect(screen.getByText(/Skipped 1 non-text file/)).toBeInTheDocument();
    });
  });

  it("shows error for empty folder with no text files", async () => {
    render(<UploadTab onCreate={vi.fn()} onCancel={vi.fn()} />);

    const png = makeFile("image.png", "binary", "my-skill/image.png", "image/png");
    selectFiles(png);

    await waitFor(() => {
      expect(screen.getByText(/No recognized text files/)).toBeInTheDocument();
    });
  });

  it("resets state when reselect is clicked", async () => {
    render(<UploadTab onCreate={vi.fn().mockResolvedValue(undefined)} onCancel={vi.fn()} />);

    const skillMd = makeFile("SKILL.md", "content", "my-skill/SKILL.md", "text/markdown");
    selectFiles(skillMd);

    await waitFor(() => {
      expect(screen.getByText("Reselect")).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText("Reselect"));

    expect(screen.getByText("Click to select folder")).toBeInTheDocument();
    expect(screen.queryByText("Reselect")).not.toBeInTheDocument();
  });

  it("recovers when onCreate rejects", async () => {
    const onCreate = vi.fn().mockRejectedValue(new Error("API error"));
    render(<UploadTab onCreate={onCreate} onCancel={vi.fn()} />);

    const skillMd = makeFile("SKILL.md", "content", "my-skill/SKILL.md", "text/markdown");
    selectFiles(skillMd);

    await waitFor(() => {
      expect(screen.getByText("Create Skill")).toBeEnabled();
    });

    fireEvent.click(screen.getByText("Create Skill"));

    await waitFor(() => {
      expect(onCreate).toHaveBeenCalledOnce();
    });

    // Button should be re-enabled after rejection
    await waitFor(() => {
      expect(screen.getByText("Create Skill")).toBeEnabled();
    });
  });
});
