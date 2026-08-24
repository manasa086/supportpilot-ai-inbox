import { useState } from "react";
import { InboxList } from "./components/InboxList";
import { TicketDetail } from "./components/TicketDetail";
import { AccountSidebar } from "./components/AccountSidebar";
import { NewTicketForm } from "./components/NewTicketForm";
import "./App.css";

export default function App() {
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [showNewTicket, setShowNewTicket] = useState(false);

  return (
    <div className="app-shell">
      <header className="app-header">
        <span className="app-logo">🛟 SupportPilot</span>
      </header>

      <main className="app-body">
        <InboxList
          selectedId={selectedId}
          onSelect={setSelectedId}
          onNewTicket={() => setShowNewTicket(true)}
        />

        {selectedId ? (
          <>
            <TicketDetail ticketId={selectedId} />
            <AccountSidebar ticketId={selectedId} onSelectTicket={setSelectedId} />
          </>
        ) : (
          <div className="pane detail-pane empty-state large">
            Select a ticket from the inbox to see its details.
          </div>
        )}
      </main>

      {showNewTicket && (
        <NewTicketForm
          onClose={() => setShowNewTicket(false)}
          onCreated={(id) => {
            setShowNewTicket(false);
            setSelectedId(id);
          }}
        />
      )}
    </div>
  );
}
