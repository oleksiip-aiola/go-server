import axios, { AxiosError } from "axios";
import { useContext } from "react";
import { UserContext } from "../contexts/UserContext";
import { jwtDecode } from "jwt-decode";
import { User } from "../types/User";

export const ENDPOINT = "http://localhost:4000";

export const useFetchData = () => {
  const { user, setUser } = useContext(UserContext);
  const protectedFetcher =
    ({ url = "/protected-resource" }) =>
    async () => {
      let localUser: User | null = null;
      if (!user) {
        try {
          const response = await axios.post(
            `${ENDPOINT}/api/refresh-token`,
            null,
            {
              withCredentials: true,
            }
          );

          if (response.status === 200) {
            const data = response.data;
            localUser = {
              ...(jwtDecode(data.access_token) as User),
              accessToken: data.access_token,
            };
            setUser(localUser);
          } else {
            console.log("Token refresh failed");
            return;
          }
        } catch (error) {
          console.log("Error occurred:", (error as AxiosError).message);
          return;
        }
      }

      try {
        const apiResponse = await axios.get(`${ENDPOINT}/${url}`, {
          headers: {
            Authorization: `Bearer ${
              user?.accessToken ?? localUser?.accessToken
            }`,
          },
        });

        const result = apiResponse.data;
        return result;
      } catch (error) {
        console.log("Error occurred:", (error as AxiosError).message);
        return error;
      }
    };

  const fetcher =
    ({ url = "/protected-resource" }) =>
    async () => {
      try {
        const apiResponse = await axios.get(`${ENDPOINT}/${url}`);

        const result = apiResponse.data;
        return result;
      } catch (error) {
        console.log("Error occurred:", (error as AxiosError).message);
        return;
      }
    };

  return { protectedFetcher, fetcher };
};
