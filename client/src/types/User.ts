export type User = {
  firstName: string;
  lastName: string;
  email: string;
  accessToken: string;
};

export type LoginDtoType = {
  email: string;
  password: string;
};

export type RegisterDtoType = LoginDtoType & {
  firstName: string;
  lastName: string;
};

export type UserContextType = {
  user: User | null;
  login: (data: LoginDtoType) => Promise<User>;
  register: (data: RegisterDtoType) => Promise<User>;
  logout: () => void;
  loading: boolean;
  setUser: (user: User | null) => void;
};
