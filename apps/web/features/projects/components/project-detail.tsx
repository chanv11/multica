"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Check, Trash2 } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { projectDetailOptions } from "@core/projects/queries";
import { useUpdateProject, useDeleteProject } from "@core/projects/mutations";
import { issueListOptions } from "@core/issues/queries";
import { useWorkspaceId } from "@core/hooks";
import { PROJECT_STATUS_ORDER, PROJECT_STATUS_CONFIG } from "../config/status";
import { StatusIcon } from "@/features/issues/components/status-icon";
import { STATUS_CONFIG } from "@/features/issues/config/status";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import type { Issue } from "@/shared/types";

function IssueRow({ issue }: { issue: Issue }) {
  return (
    <Link
      href={`/issues/${issue.id}`}
      className="flex items-center gap-2 rounded-md px-3 py-2 text-sm hover:bg-accent/50 transition-colors"
    >
      <StatusIcon status={issue.status} className="h-3.5 w-3.5 shrink-0" />
      <span className="text-muted-foreground shrink-0 text-xs">{issue.identifier}</span>
      <span className="truncate">{issue.title}</span>
      <span className={`ml-auto shrink-0 text-xs ${STATUS_CONFIG[issue.status].iconColor}`}>
        {STATUS_CONFIG[issue.status].label}
      </span>
    </Link>
  );
}

export function ProjectDetail({ projectId }: { projectId: string }) {
  const wsId = useWorkspaceId();
  const router = useRouter();
  const { data: project, isLoading } = useQuery(projectDetailOptions(wsId, projectId));
  const { data: allIssues = [] } = useQuery(issueListOptions(wsId));
  const updateProject = useUpdateProject();
  const deleteProject = useDeleteProject();
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

  const projectIssues = useMemo(
    () => allIssues.filter((i) => i.project_id === projectId),
    [allIssues, projectId],
  );

  if (isLoading) {
    return (
      <div className="p-6 space-y-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-4 w-96" />
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (!project) {
    return (
      <div className="flex items-center justify-center h-full text-muted-foreground">
        Project not found
      </div>
    );
  }

  const statusCfg = PROJECT_STATUS_CONFIG[project.status];

  const handleDelete = () => {
    deleteProject.mutate(project.id, {
      onSuccess: () => router.push("/projects"),
    });
  };

  return (
    <div className="flex h-full flex-col">
      {/* Header */}
      <div className="flex items-center gap-3 border-b px-6 py-3">
        <Button
          variant="ghost"
          size="sm"
          className="h-7 w-7 p-0"
          onClick={() => router.push("/projects")}
        >
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <span className="text-lg">{project.icon || "📁"}</span>
        <h1 className="text-sm font-medium flex-1 truncate">{project.title}</h1>

        {/* Status dropdown */}
        <DropdownMenu>
          <DropdownMenuTrigger
            render={
              <Button variant="outline" size="sm">
                <span className={`inline-flex items-center gap-1 ${statusCfg.color}`}>
                  {statusCfg.label}
                </span>
              </Button>
            }
          />
          <DropdownMenuContent align="end" className="w-44">
            {PROJECT_STATUS_ORDER.map((s) => (
              <DropdownMenuItem
                key={s}
                onClick={() => updateProject.mutate({ id: project.id, status: s })}
              >
                <span className={PROJECT_STATUS_CONFIG[s].color}>
                  {PROJECT_STATUS_CONFIG[s].label}
                </span>
                {s === project.status && <Check className="ml-auto h-3.5 w-3.5" />}
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>

        {/* Delete */}
        <Button
          variant="ghost"
          size="sm"
          className="h-7 w-7 p-0 text-muted-foreground hover:text-destructive"
          onClick={() => setDeleteDialogOpen(true)}
        >
          <Trash2 className="h-3.5 w-3.5" />
        </Button>
        <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Delete project</AlertDialogTitle>
              <AlertDialogDescription>
                This will delete the project. Issues will not be deleted but will be unlinked.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <AlertDialogAction onClick={handleDelete}>Delete</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>

      {/* Tabs */}
      <Tabs defaultValue="overview" className="flex-1 flex flex-col overflow-hidden">
        <TabsList className="mx-6 mt-3 w-fit">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="issues">
            Issues{projectIssues.length > 0 && ` (${projectIssues.length})`}
          </TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="flex-1 overflow-y-auto p-6 space-y-6">
          {/* Description */}
          <div>
            <label className="text-xs font-medium text-muted-foreground mb-1 block">
              Description
            </label>
            <Textarea
              placeholder="Add a description..."
              defaultValue={project.description ?? ""}
              onBlur={(e) => {
                const val = e.target.value;
                if (val !== (project.description ?? "")) {
                  updateProject.mutate({
                    id: project.id,
                    description: val || null,
                  });
                }
              }}
              className="min-h-[100px] resize-none"
            />
          </div>

          {/* Stats */}
          <div>
            <label className="text-xs font-medium text-muted-foreground mb-2 block">
              Summary
            </label>
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
              <div className="rounded-lg border p-3">
                <div className="text-2xl font-semibold">{projectIssues.length}</div>
                <div className="text-xs text-muted-foreground">Total issues</div>
              </div>
              <div className="rounded-lg border p-3">
                <div className="text-2xl font-semibold">
                  {projectIssues.filter((i) => i.status === "done").length}
                </div>
                <div className="text-xs text-muted-foreground">Completed</div>
              </div>
              <div className="rounded-lg border p-3">
                <div className="text-2xl font-semibold">
                  {projectIssues.filter((i) => i.status === "in_progress" || i.status === "in_review").length}
                </div>
                <div className="text-xs text-muted-foreground">In progress</div>
              </div>
              <div className="rounded-lg border p-3">
                <div className="text-2xl font-semibold">
                  {projectIssues.filter((i) => i.status === "todo" || i.status === "backlog").length}
                </div>
                <div className="text-xs text-muted-foreground">Not started</div>
              </div>
            </div>
          </div>
        </TabsContent>

        <TabsContent value="issues" className="flex-1 overflow-y-auto p-6">
          {projectIssues.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground text-sm">
              No issues linked to this project yet.
              <br />
              <span className="text-xs">
                Assign issues to this project from the issue detail panel.
              </span>
            </div>
          ) : (
            <div className="space-y-0.5">
              {projectIssues.map((issue) => (
                <IssueRow key={issue.id} issue={issue} />
              ))}
            </div>
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
}
