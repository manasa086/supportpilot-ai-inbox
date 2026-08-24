import { useState } from "react";
import { useMutation, useQuery } from "@apollo/client/react";
import { CREATE_TICKET, GET_ACCOUNTS } from "../graphql/operations";

export function NewTicketForm({
  onClose,
  onCreated,
}: {
  onClose: () => void;
  onCreated: (ticketId: string) => void;
}) {
  const { data: accountsData } = useQuery(GET_ACCOUNTS);
  const [accountId, setAccountId] = useState("");
  const [subject, setSubject] = useState("");
  const [body, setBody] = useState("");
  const [requesterEmail, setRequesterEmail] = useState("");

  const [createTicket, { loading, error }] = useMutation(CREATE_TICKET, {
    // GetTickets picks up the new row; GetAccount picks up the new entry
    // in that account's ticket-history list (Apollo's normalized cache
    // updates shared fields on existing entities automatically, but a
    // brand-new list item needs an explicit refetch).
    refetchQueries: ["GetTickets", "GetAccount"],
  });

  const canSubmit = accountId && subject.trim() && body.trim() && requesterEmail.trim();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!canSubmit) return;
    const { data } = await createTicket({
      variables: { input: { accountId, subject, body, requesterEmail } },
    });
    if (data) onCreated(data.createTicket.id);
  };

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <form className="modal" onClick={(e) => e.stopPropagation()} onSubmit={handleSubmit}>
        <h2>New ticket</h2>

        <label>
          Account
          <select value={accountId} onChange={(e) => setAccountId(e.target.value)} required>
            <option value="" disabled>
              Select an account…
            </option>
            {accountsData?.accounts.map((a) => (
              <option key={a.id} value={a.id}>
                {a.name} ({a.plan})
              </option>
            ))}
          </select>
        </label>

        <label>
          Requester email
          <input
            type="email"
            value={requesterEmail}
            onChange={(e) => setRequesterEmail(e.target.value)}
            placeholder="user@customer.com"
            required
          />
        </label>

        <label>
          Subject
          <input value={subject} onChange={(e) => setSubject(e.target.value)} required />
        </label>

        <label>
          Body
          <textarea value={body} onChange={(e) => setBody(e.target.value)} rows={5} required />
        </label>

        {error && <div className="inline-error">{error.message}</div>}

        <div className="modal-actions">
          <button type="button" className="btn btn-ghost btn-sm" onClick={onClose}>
            Cancel
          </button>
          <button type="submit" className="btn btn-primary btn-sm" disabled={!canSubmit || loading}>
            {loading ? "Creating…" : "Create ticket"}
          </button>
        </div>
      </form>
    </div>
  );
}
