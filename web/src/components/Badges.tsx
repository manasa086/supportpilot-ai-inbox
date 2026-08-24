import type { TicketPriority, TicketStatus } from "../graphql/types";

const STATUS_LABEL: Record<TicketStatus, string> = {
  NEW: "New",
  TRIAGED: "Triaged",
  IN_PROGRESS: "In Progress",
  RESOLVED: "Resolved",
};

const PRIORITY_LABEL: Record<TicketPriority, string> = {
  LOW: "Low",
  MEDIUM: "Medium",
  HIGH: "High",
  URGENT: "Urgent",
};

export function StatusBadge({ status }: { status: TicketStatus }) {
  return <span className={`badge badge-status-${status.toLowerCase()}`}>{STATUS_LABEL[status]}</span>;
}

export function PriorityBadge({ priority }: { priority: TicketPriority | null }) {
  if (!priority) return <span className="badge badge-priority-none">Untriaged</span>;
  return <span className={`badge badge-priority-${priority.toLowerCase()}`}>{PRIORITY_LABEL[priority]}</span>;
}
