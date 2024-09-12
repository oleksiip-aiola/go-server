import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App.tsx";
import "./index.css";
import { MantineProvider } from "@mantine/core";
import { Notifications } from "@mantine/notifications";
import "@mantine/core/styles.css";
import "@mantine/notifications/styles.css";
import UserProviderWrapper from "./provider/userProvider.tsx";
import {
  createBrowserRouter,
  createRoutesFromElements,
  Route,
  RouterProvider,
} from "react-router-dom";
import Login from "./features/Auth/Login/Login.tsx";
import ProtectedRoute from "./features/Auth/ProtectedRoute/ProtectedRoute.tsx";
import { Todos } from "./features/Todos/index.tsx";

const router = createBrowserRouter(
  createRoutesFromElements(
    <>
      <Route path="/" element={<App />}>
        <Route path="/" element={<Login />} />
        <Route path="dashboard">
          <Route path="todo" element={<ProtectedRoute component={Todos} />} />
        </Route>
      </Route>
      <Route path="/login" element={<App />} />
    </>
  )
);

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <MantineProvider>
      <UserProviderWrapper>
        <Notifications />
        <RouterProvider router={router} />
      </UserProviderWrapper>
    </MantineProvider>
  </StrictMode>
);
