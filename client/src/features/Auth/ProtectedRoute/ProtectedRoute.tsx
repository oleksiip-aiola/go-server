/* eslint-disable @typescript-eslint/no-explicit-any */

import { ComponentType } from "react";
import { Navigate, Outlet } from "react-router-dom";
import { useUserContext } from "../../../hooks/useUserContext";
import { Loader } from "@mantine/core";

const ProtectedRoute = ({
  component: Component,
}: {
  component: ComponentType<any>;
  [key: string]: any;
}) => {
  const { user, loading } = useUserContext();

  if (loading) {
    return <Loader />;
  }

  if (!loading && user) {
    // If user exists, render the passed component or the Outlet if no children
    return <Component /> || <Outlet />;
  }

  return <Navigate to="/login" />;
};

export default ProtectedRoute;
