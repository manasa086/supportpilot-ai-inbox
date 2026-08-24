import { useState } from "react";
import { useQuery } from "@apollo/client/react";
import { GET_TICKETS } from "../graphql/operations";
import type { TicketPriority, TicketStatus } from "../graphql/types";
import { PriorityBadge, StatusBadge } from "./Badges";

const STATUS_OPTIONS: TicketStatus[] = ["NEW", "TRIAGED", "IN_PROGRESS", "RESOLVED"];
const PRIORITY_OPTIONS: TicketPriority[] = ["URGENT", "HIGH", "MEDIUM", "LOW"];

function timeAgo(iso: string): string {
  const diffMs = Date.now() - new Date(iso).getTime();
  const mins = Math.round(diffMs / 60000);
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.round(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  return `${Math.round(hours / 24)}d ago`;
}

export function InboxList({
  selectedId,
  onSelect,
  onNewTicket,
}: {
  selectedId: string | null;
  onSelect: (id: string) => void;
  onNewTicket: () => void;
}) {
  const [status, setStatus] = useState<TicketStatus | "">("");
  const [priority, setPriority] = useState<TicketPriority | "">("");

  const { data, loading, error, refetch } = useQuery(GET_TICKETS, {
    variables: {
      status: status || null,
      priority: priority || null,
    },
  });

  return (
    <div className="pane inbox-pane">
      <div className="pane-header">
        <h1>Inbox</h1>
        <button className="btn btn-primary btn-sm" onClick={onNewTicket}>
          + New ticket
        </button>
      </div>

      <div className="filter-bar">
        <select value={status} onChange={(e) => setStatus(e.target.value as TicketStatus | "")}>
          <option value="">All statuses</option>
          {STATUS_OPTIONS.map((s) => (
            <option key={s} value={s}>
              {s.replace("_", " ")}
            </option>
          ))}
        </select>
        <select value={priority} onChange={(e) => setPriority(e.target.value as TicketPriority | "")}>
          <option value="">All priorities</option>
          {PRIORITY_OPTIONS.map((p) => (
            <option key={p} value={p}>
              {p}
            </option>
          ))}
        </select>
        <button className="btn btn-ghost btn-sm" onClick={() => refetch()} title="Refresh">
          ↻
        </button>
      </div>

      {loading && <div className="empty-state">Loading tickets…</div>}
      {error && <div className="empty-state error">Couldn't load tickets: {error.message}</div>}
      {!loading && !error && data?.tickets.length === 0 && (
        <div className="empty-state">No tickets match these filters.</div>
      )}

      <ul className="ticket-list">
        {data?.tickets.map((t) => (
          <li key={t.id}>
            <button
              className={`ticket-row ${t.id === selectedId ? "selected" : ""}`}
              onClick={() => onSelect(t.id)}
            >
              <div className="ticket-row-top">
                <span className="ticket-account">{t.account.name}</span>
                <span className="ticket-time">{timeAgo(t.createdAt)}</span>
              </div>
              <div className="ticket-subject">{t.subject}</div>
              <div className="ticket-row-bottom">
                <StatusBadge status={t.status} />
                <PriorityBadge priority={t.priority} />
              </div>
            </button>
          </li>
        ))}
      </ul>
    </div>
  );
}
