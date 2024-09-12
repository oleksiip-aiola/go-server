import useSWR from "swr";
import { AddTodo } from "./AddTodo";
import { TodoList } from "./TodoList";
import { useFetchData } from "../../utils/fetcher";
import { Todo } from "./types";
import { memo, useContext } from "react";
import { UserContext } from "../../contexts/UserContext";

export const Todos = memo(() => {
  const { user } = useContext(UserContext);
  const { protectedFetcher } = useFetchData();
  const { data, mutate } = useSWR<Todo[]>(
    [user, "todos"],
    protectedFetcher({ url: "todos", token: user?.accessToken }),
    {
      revalidateOnFocus: false,
    }
  );

  return (
    <>
      <TodoList {...{ data, mutate }} />
      <AddTodo {...{ mutate }} />
    </>
  );
});
