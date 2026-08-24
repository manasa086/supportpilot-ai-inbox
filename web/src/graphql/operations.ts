import { gql, type TypedDocumentNode } from "@apollo/client";
import type {
  Account,
  AccountWithTickets,
  CreateTicketInput,
  Ticket,
  TicketPriority,
  TicketStatus,
} from "./types";

const TICKET_FIELDS = gql`
  fragment TicketFields on Ticket {
    id
    subject
    body
    status
    priority
    category
    requesterEmail
    createdAt
    updatedAt
    suggestion {
      category
      priority
      reply
      confidence
    }
    account {
      id
      name
      plan
    }
  }
`;

export const GET_TICKETS: TypedDocumentNode<
  { tickets: Ticket[] },
  { status?: TicketStatus | null; priority?: TicketPriority | null }
> = gql`
  query GetTickets($status: TicketStatus, $priority: TicketPriority) {
    tickets(status: $status, priority: $priority) {
      ...TicketFields
    }
  }
  ${TICKET_FIELDS}
`;

export const GET_TICKET: TypedDocumentNode<
  { ticket: Ticket | null },
  { id: string }
> = gql`
  query GetTicket($id: ID!) {
    ticket(id: $id) {
      ...TicketFields
    }
  }
  ${TICKET_FIELDS}
`;

export const GET_ACCOUNTS: TypedDocumentNode<{ accounts: Account[] }, Record<string, never>> = gql`
  query GetAccounts {
    accounts {
      id
      name
      plan
      seats
    }
  }
`;

export const GET_ACCOUNT: TypedDocumentNode<
  { account: AccountWithTickets | null },
  { id: string }
> = gql`
  query GetAccount($id: ID!) {
    account(id: $id) {
      id
      name
      plan
      seats
      createdAt
      summary
      tickets {
        id
        subject
        status
        priority
        createdAt
      }
    }
  }
`;

export const CREATE_TICKET: TypedDocumentNode<
  { createTicket: Ticket },
  { input: CreateTicketInput }
> = gql`
  mutation CreateTicket($input: CreateTicketInput!) {
    createTicket(input: $input) {
      ...TicketFields
    }
  }
  ${TICKET_FIELDS}
`;

export const TRIAGE_TICKET: TypedDocumentNode<{ triageTicket: Ticket }, { id: string }> = gql`
  mutation TriageTicket($id: ID!) {
    triageTicket(id: $id) {
      ...TicketFields
    }
  }
  ${TICKET_FIELDS}
`;

export const RESOLVE_TICKET: TypedDocumentNode<{ resolveTicket: Ticket }, { id: string }> = gql`
  mutation ResolveTicket($id: ID!) {
    resolveTicket(id: $id) {
      ...TicketFields
    }
  }
  ${TICKET_FIELDS}
`;
