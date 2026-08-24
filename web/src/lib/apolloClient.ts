import { ApolloClient, InMemoryCache, HttpLink } from "@apollo/client";

// Points at the Go API's /query endpoint. Defaults to the local dev server
// (`go run ./cmd/server` in ../api); set VITE_API_URL once the API is
// deployed (see ../api/infra/template.yaml's ApiUrl output).
const API_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080/query";

const httpLink = new HttpLink({ uri: API_URL });

export const apolloClient = new ApolloClient({
  link: httpLink,
  cache: new InMemoryCache(),
});
