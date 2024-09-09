import { Box, List, ThemeIcon } from "@mantine/core";
import "./App.css";
import useSWR from "swr";
import { AddTodo } from "./components/AddTodo";
import "@mantine/core/styles.css";
import { CheckCircleFillIcon } from "@primer/octicons-react";

export interface Todo {
  title: string;
  body: string;
  done: boolean;
  id: number;
}

export const ENDPOINT = "http://localhost:4000";
const fetcher = (url: string) =>
  fetch(`${ENDPOINT}/${url}`).then((res) => res.json());

function App() {
  const { data, mutate } = useSWR<Todo[]>("api/todos", fetcher);

  const handleStatusUpdate = async (id: number) => {
    const updated = await fetch(`${ENDPOINT}/api/todos/${id}/status`, {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
      },
    }).then((r) => {
      return r.json();
    });

    await mutate(updated);
  };

  return (
    <Box
      style={{
        padding: "2rem",
        width: "100%",
        maxWidth: "40rem",
        margin: "0 auto",
      }}
    >
      <List spacing={"xs"} size="sm" mb={12} center>
        {data?.map((todo) => {
          return (
            <List.Item
              key={`todo__${todo.id}`}
              icon={
                todo.done ? (
                  <ThemeIcon
                    color="teal"
                    size={24}
                    radius="xl"
                    onClick={() => handleStatusUpdate(todo.id)}
                  >
                    <CheckCircleFillIcon size={20} />
                  </ThemeIcon>
                ) : (
                  <ThemeIcon
                    color="gray"
                    size={24}
                    radius="xl"
                    onClick={() => handleStatusUpdate(todo.id)}
                  >
                    <CheckCircleFillIcon />
                  </ThemeIcon>
                )
              }
            >
              {todo.title}
            </List.Item>
          );
        })}
      </List>
      <AddTodo {...{ mutate }} />
    </Box>
  );
}

export default App;
