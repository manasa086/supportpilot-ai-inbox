import { useQuery } from "@apollo/client/react";
import { GET_ACCOUNT, GET_TICKET } from "../graphql/operations";
import { PriorityBadge, StatusBadge } from "./Badges";

export function AccountSidebar({
  ticketId,
  onSelectTicket,
}: {
  ticketId: string;
  onSelectTicket: (id: string) => void;
}) {
  // Look up the account ID via the currently-selected ticket, then load
  // that account's profile + full ticket history.
  const { data: ticketData } = useQuery(GET_TICKET, { variables: { id: ticketId } });
  const accountId = ticketData?.ticket?.account.id;

  const { data, loading, error } = useQuery(GET_ACCOUNT, {
    skip: !accountId,
    variables: { id: accountId ?? "" },
  });

  if (!accountId || loading) return <div className="pane account-pane empty-state">Loading account…</div>;
  if (error) return <div className="pane account-pane empty-state error">Couldn't load account: {error.message}</div>;
  const account = data?.account;
  if (!account) return <div className="pane account-pane empty-state">Account not found.</div>;

  return (
    <div className="pane account-pane">
      <div className="pane-header">
        <h1>{account.name}</h1>
      </div>

      <dl className="account-facts">
        <div>
          <dt>Plan</dt>
          <dd>{account.plan}</dd>
        </div>
        <div>
          <dt>Seats</dt>
          <dd>{account.seats}</dd>
        </div>
        <div>
          <dt>Customer since</dt>
          <dd>{new Date(account.createdAt).toLocaleDateString()}</dd>
        </div>
      </dl>

      {account.summary && (
        <div className="account-summary">
          <div className="account-summary-title">Account summary</div>
          <p>{account.summary}</p>
        </div>
      )}

      <div className="account-tickets-title">
        Ticket history ({account.tickets.length})
      </div>
      <ul className="account-ticket-list">
        {account.tickets
          .slice()
          .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
          .map((t) => (
            <li key={t.id}>
              <button
                className={`account-ticket-row ${t.id === ticketId ? "selected" : ""}`}
                onClick={() => onSelectTicket(t.id)}
              >
                <span className="account-ticket-subject">{t.subject}</span>
                <span className="account-ticket-badges">
                  <StatusBadge status={t.status} />
                  <PriorityBadge priority={t.priority} />
                </span>
              </button>
            </li>
          ))}
      </ul>
    </div>
  );
}
