import { jwtDecode } from "jwt-decode";
import { LoginDtoType, RegisterDtoType, User } from "../types/User";
import axios, { AxiosError } from "axios";
import { ENDPOINT } from "../utils/fetcher";

export const authService = {
  login: async ({ email, password }: LoginDtoType): Promise<User> => {
    const { data } = await axios.post(
      `${ENDPOINT}/api/login`,
      {
        email,
        password,
      },
      {
        headers: {
          "Content-Type": "application/json",
        },
        withCredentials: true,
      }
    );

    return {
      ...(jwtDecode(data.access_token) as User),
      accessToken: data.access_token,
    };
  },

  register: async ({
    firstName,
    lastName,
    email,
    password,
  }: RegisterDtoType): Promise<User> => {
    const { data } = await axios.post(
      `${ENDPOINT}/api/register`,
      {
        firstName,
        lastName,
        email,
        password,
      },
      {
        headers: {
          "Content-Type": "application/json",
        },
        withCredentials: true,
      }
    );
    return {
      ...(jwtDecode(data.access_token) as User),
      accessToken: data.access_token,
    };
  },

  logout: async () => {
    return axios.post(`${ENDPOINT}/api/logout`, null, {
      withCredentials: true,
    });
  },

  fetchCurrentUser: async (): Promise<User | null> => {
    try {
      const response = await axios.post(`${ENDPOINT}/api/refresh-token`, null, {
        withCredentials: true,
      });

      if (response.status === 200) {
        const data = response.data;
        return {
          ...(jwtDecode(data.access_token) as User),
          accessToken: data.access_token,
        };
      } else {
        console.log("Token refresh failed");
        return null;
      }
    } catch (error) {
      console.log("Error occurred:", (error as AxiosError).message);
      return null;
    }
  },
};
