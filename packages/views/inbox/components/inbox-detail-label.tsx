"use client";

import { STATUS_CONFIG, PRIORITY_CONFIG } from "@multica/core/issues/config";
import { useActorName } from "@multica/core/workspace/hooks";
import { StatusIcon, PriorityIcon } from "../../issues/components";
import { useAppLocale } from "../../i18n";
import type { InboxItem, InboxItemType, IssueStatus, IssuePriority } from "@multica/core/types";

const typeKeyMap: Record<InboxItemType, string> = {
  issue_assigned: "typeAssigned",
  unassigned: "typeUnassigned",
  assignee_changed: "typeAssigneeChanged",
  status_changed: "typeStatusChanged",
  priority_changed: "typePriorityChanged",
  due_date_changed: "typeDueDateChanged",
  new_comment: "typeNewComment",
  mentioned: "typeMentioned",
  review_requested: "typeReviewRequested",
  task_completed: "typeTaskCompleted",
  task_failed: "typeTaskFailed",
  agent_blocked: "typeAgentBlocked",
  agent_completed: "typeAgentCompleted",
  reaction_added: "typeReactionAdded",
};

export function getTypeLabel(t: Record<string, string>, type: InboxItemType): string {
  return t[typeKeyMap[type]] ?? type;
}

export { typeKeyMap };

function shortDate(dateStr: string): string {
  if (!dateStr) return "";
  return new Date(dateStr).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
  });
}

export function InboxDetailLabel({ item }: { item: InboxItem }) {
  const { getActorName } = useActorName();
  const { t } = useAppLocale();
  const details = item.details ?? {};

  switch (item.type) {
    case "status_changed": {
      if (!details.to) return <span>{getTypeLabel(t.inbox, item.type)}</span>;
      const label = STATUS_CONFIG[details.to as IssueStatus]?.label ?? details.to;
      return (
        <span className="inline-flex items-center gap-1">
          {t.inbox.setStatusTo}
          <StatusIcon status={details.to as IssueStatus} className="h-3 w-3" />
          {label}
        </span>
      );
    }
    case "priority_changed": {
      if (!details.to) return <span>{getTypeLabel(t.inbox, item.type)}</span>;
      const label = PRIORITY_CONFIG[details.to as IssuePriority]?.label ?? details.to;
      return (
        <span className="inline-flex items-center gap-1">
          {t.inbox.setPriorityTo}
          <PriorityIcon priority={details.to as IssuePriority} className="h-3 w-3" />
          {label}
        </span>
      );
    }
    case "issue_assigned": {
      if (details.new_assignee_id) {
        return <span>{t.inbox.assignedTo} {getActorName(details.new_assignee_type ?? "member", details.new_assignee_id)}</span>;
      }
      return <span>{getTypeLabel(t.inbox, item.type)}</span>;
    }
    case "unassigned":
      return <span>{t.inbox.removedAssignee}</span>;
    case "assignee_changed": {
      if (details.new_assignee_id) {
        return <span>{t.inbox.assignedTo} {getActorName(details.new_assignee_type ?? "member", details.new_assignee_id)}</span>;
      }
      return <span>{getTypeLabel(t.inbox, item.type)}</span>;
    }
    case "due_date_changed": {
      if (details.to) return <span>{t.inbox.setDueDateTo} {shortDate(details.to)}</span>;
      return <span>{t.inbox.removedDueDate}</span>;
    }
    case "new_comment": {
      if (item.body) return <span>{item.body}</span>;
      return <span>{getTypeLabel(t.inbox, item.type)}</span>;
    }
    case "reaction_added": {
      const emoji = details.emoji;
      if (emoji) return <span>{getTypeLabel(t.inbox, item.type)} {emoji}</span>;
      return <span>{getTypeLabel(t.inbox, item.type)}</span>;
    }
    default:
      return <span>{getTypeLabel(t.inbox, item.type) ?? item.type}</span>;
  }
}
