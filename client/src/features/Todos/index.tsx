import useSWR from "swr";
import { AddTodo } from "./AddTodo";
import { TodoList } from "./TodoList";
import { useFetchData } from "../../utils/fetcher";
import { Todo } from "./types";

export const Todos = () => {
  const { protectedFetcher } = useFetchData();
  const { data, mutate } = useSWR<Todo[]>(
    "api/todos",
    protectedFetcher({ url: "todos" })
  );
  console.log("todos");
  return (
    <>
      <TodoList {...{ data, mutate }} />
      <AddTodo {...{ mutate }} />
    </>
  );
};
