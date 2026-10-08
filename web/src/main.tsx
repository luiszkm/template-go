import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App } from "./App";
import { makeQueryClient, makeRouter } from "./router";
import "./index.css";

const root = document.getElementById("root");
if (!root) throw new Error("#root not found");

const queryClient = makeQueryClient();

createRoot(root).render(
  <StrictMode>
    <App router={makeRouter(queryClient)} queryClient={queryClient} />
  </StrictMode>,
);
