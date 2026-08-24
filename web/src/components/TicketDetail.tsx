import { useQuery, useMutation } from "@apollo/client/react";
import { GET_TICKET, RESOLVE_TICKET, TRIAGE_TICKET } from "../graphql/operations";
import { PriorityBadge, StatusBadge } from "./Badges";

export function TicketDetail({ ticketId }: { ticketId: string }) {
  const { data, loading, error } = useQuery(GET_TICKET, {
    variables: { id: ticketId },
  });

  const [triageTicket, { loading: triaging, error: triageError }] = useMutation(TRIAGE_TICKET, {
    refetchQueries: ["GetTickets"],
  });
  const [resolveTicket, { loading: resolving }] = useMutation(RESOLVE_TICKET, {
    refetchQueries: ["GetTickets"],
  });

  if (loading) return <div className="pane detail-pane empty-state">Loading ticket…</div>;
  if (error) return <div className="pane detail-pane empty-state error">Couldn't load ticket: {error.message}</div>;
  const ticket = data?.ticket;
  if (!ticket) return <div className="pane detail-pane empty-state">Ticket not found.</div>;

  const canTriage = ticket.status === "NEW";
  const canResolve = ticket.status !== "RESOLVED";

  return (
    <div className="pane detail-pane">
      <div className="pane-header">
        <h1>{ticket.subject}</h1>
        <div className="detail-actions">
          <button
            className="btn btn-primary btn-sm"
            disabled={!canTriage || triaging}
            onClick={() => triageTicket({ variables: { id: ticket.id } })}
            title={canTriage ? "Classify with AI and draft a reply" : "Already triaged"}
          >
            {triaging ? "Triaging…" : "✨ Triage with AI"}
          </button>
          <button
            className="btn btn-secondary btn-sm"
            disabled={!canResolve || resolving}
            onClick={() => resolveTicket({ variables: { id: ticket.id } })}
          >
            {resolving ? "Resolving…" : "Resolve"}
          </button>
        </div>
      </div>

      <div className="detail-meta">
        <StatusBadge status={ticket.status} />
        <PriorityBadge priority={ticket.priority} />
        {ticket.category && <span className="badge badge-category">{ticket.category}</span>}
        <span className="detail-meta-sep">·</span>
        <span className="detail-requester">{ticket.requesterEmail}</span>
        <span className="detail-meta-sep">·</span>
        <span className="detail-account">{ticket.account.name}</span>
      </div>

      {triageError && <div className="inline-error">Triage failed: {triageError.message}</div>}

      <div className="ticket-body">{ticket.body}</div>

      {ticket.suggestion && (
        <div className="suggestion-panel">
          <div className="suggestion-header">
            <span className="suggestion-title">🤖 AI suggestion</span>
            <span className="suggestion-confidence">
              {Math.round(ticket.suggestion.confidence * 100)}% confidence
            </span>
          </div>
          <p className="suggestion-reply">{ticket.suggestion.reply}</p>
          <button
            className="btn btn-ghost btn-sm"
            onClick={() => navigator.clipboard.writeText(ticket.suggestion!.reply)}
          >
            Copy reply
          </button>
        </div>
      )}
    </div>
  );
}
