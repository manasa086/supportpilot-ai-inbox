// Hand-written types mirroring api/graph/schema.graphqls. A weekend-MVP
// stand-in for full `graphql-codegen` generation — see
// ../../node_modules/@apollo/client/skills/apollo-client/references/typescript-codegen.md
// for how to wire that up later if the schema grows.

export type TicketStatus = "NEW" | "TRIAGED" | "IN_PROGRESS" | "RESOLVED";
export type TicketPriority = "LOW" | "MEDIUM" | "HIGH" | "URGENT";

export interface Suggestion {
  category: string;
  priority: TicketPriority;
  reply: string;
  confidence: number;
}

export interface Account {
  id: string;
  name: string;
  plan: string;
  seats: number;
  createdAt: string;
  summary?: string | null;
}

export interface TicketSummary {
  id: string;
  subject: string;
  status: TicketStatus;
  priority: TicketPriority | null;
  createdAt: string;
}

export interface AccountWithTickets extends Account {
  tickets: TicketSummary[];
}

export interface AccountSummary {
  id: string;
  name: string;
  plan: string;
}

export interface Ticket {
  id: string;
  subject: string;
  body: string;
  status: TicketStatus;
  priority: TicketPriority | null;
  category: string | null;
  requesterEmail: string;
  createdAt: string;
  updatedAt: string;
  suggestion: Suggestion | null;
  account: AccountSummary;
}

export interface CreateTicketInput {
  accountId: string;
  subject: string;
  body: string;
  requesterEmail: string;
}
