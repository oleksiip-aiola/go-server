import { ReactNode } from "react";
import { UserProvider } from "../contexts/UserContext";

const UserProviderWrapper = ({ children }: { children: ReactNode }) => {
  return <UserProvider>{children}</UserProvider>;
};

export default UserProviderWrapper;
