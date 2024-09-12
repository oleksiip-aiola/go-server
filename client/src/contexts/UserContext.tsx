import React, { createContext, useEffect, useState } from "react";
import { authService } from "../services/authService";
import {
  LoginDtoType,
  RegisterDtoType,
  User,
  UserContextType,
} from "../types/User";
import useSWR from "swr";

// Create the user context
export const UserContext = createContext<UserContextType>({
  user: null,
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  login: async (_: LoginDtoType) => {
    return {} as User;
  },
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  register: async (_: RegisterDtoType) => {
    return Promise.resolve({} as User);
  },
  logout: () => {},
  loading: true,
  setUser: () => {},
});

// Provider component
export const UserProvider = ({ children }: { children: React.ReactNode }) => {
  const [user, setUser] = useState<User | null>(null); // Stores user info
  const [hasLoggedOut, setHasLoggedOut] = useState(false); // Flag to check if user has logged out

  const { data, isLoading, mutate } = useSWR(
    !user && !hasLoggedOut ? "/api/refresh-token" : null,
    authService.fetchCurrentUser
  );

  useEffect(() => {
    if (data && !isLoading) {
      setUser(data);
    }
  }, [data, isLoading]);

  const login = async ({ email, password }: LoginDtoType) => {
    const loggedInUser = await authService.login({ email, password });

    setUser(loggedInUser);
    await mutate(loggedInUser);

    setHasLoggedOut(false);
    return loggedInUser;
  };

  const register = async ({
    firstName,
    lastName,
    email,
    password,
  }: RegisterDtoType) => {
    const loggedInUser = await authService.register({
      firstName,
      lastName,
      email,
      password,
    });

    await mutate(loggedInUser);

    setUser(loggedInUser);
    setHasLoggedOut(false);

    return loggedInUser;
  };

  const logout = async () => {
    const loggedInUser = await authService.logout();

    setHasLoggedOut(true);
    setUser(null);
    mutate(null);

    return loggedInUser;
  };

  return (
    <UserContext.Provider
      value={{
        user: (user || data) as User,
        login,
        register,
        logout,
        loading: isLoading,
        setUser,
      }}
    >
      {children}
      {/* Render children only after user is loaded */}
    </UserContext.Provider>
  );
};
