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

    const skillMd = new File(["# My Skill\n\nDoes things."], "SKILL.md", {
      type: "text/markdown",
    });
    Object.defineProperty(skillMd, "webkitRelativePath", {
      value: "my-skill/SKILL.md",
      writable: false,
    });

    const helper = new File(["export const x = 1;"], "helper.ts", {
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

    const call = onCreate.mock.calls[0]![0];
    expect(call.name).toBe("my-skill");
    expect(call.content).toBe("# My Skill\n\nDoes things.");
    expect(call.files).toHaveLength(1);
    expect(call.files[0].path).toBe("src/helper.ts");
  });

  it("allows removing files from the list", async () => {
    const onCreate = vi.fn().mockResolvedValue(undefined);
    render(<UploadTab onCreate={onCreate} onCancel={vi.fn()} />);

    const skillMd = new File(["skill content"], "SKILL.md", {
      type: "text/markdown",
    });
    Object.defineProperty(skillMd, "webkitRelativePath", {
      value: "my-skill/SKILL.md",
      writable: false,
    });

    const extra = new File(["extra"], "extra.txt", { type: "text/plain" });
    Object.defineProperty(extra, "webkitRelativePath", {
      value: "my-skill/extra.txt",
      writable: false,
    });

    const input = document.querySelector('input[type="file"]')!;
    fireEvent.change(input, { target: { files: [skillMd, extra] } });

    await waitFor(() => {
      expect(screen.getByText("extra.txt")).toBeInTheDocument();
    });

    const removeBtn = screen.getByText("extra.txt")
      .closest("div")!
      .querySelector("button")!;
    fireEvent.click(removeBtn);

    expect(screen.queryByText("extra.txt")).not.toBeInTheDocument();
  });
});
